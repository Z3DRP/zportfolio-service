package wsman

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Z3DRP/zportfolio-service/config"
	"github.com/Z3DRP/zportfolio-service/internal/controller"
	"github.com/Z3DRP/zportfolio-service/internal/dtos"
	"github.com/Z3DRP/zportfolio-service/internal/models"
	"github.com/Z3DRP/zportfolio-service/internal/utils"
	"github.com/Z3DRP/zportfolio-service/internal/zlogger"
	"github.com/gorilla/websocket"
)

var (
	pongWait     = 10 * time.Second
	pingInterval = (pongWait * 9) / 10
)

var (
	ErrEventNotSupported = errors.New("this event type is not supported")
)

var (
	WebsocketUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return config.IsValidOrigin(r.Header.Get("origin")) },
	}
)

type ErrFailedBroadcast struct {
	EffectedPeriodStart time.Time
	EffectedPeriodEnd   time.Time
	Client              *Client
	Err                 error
}

func (e ErrFailedBroadcast) Error() string {
	return fmt.Sprintf("failed to broadcast updated for period: {Start: %v, End: %v} on conn: %v", e.EffectedPeriodStart, e.EffectedPeriodEnd, e.Client.Connection.RemoteAddr())
}

func (e ErrFailedBroadcast) Unwrap() error {
	return e.Err
}

type Manager struct {
	Clients  ClientList
	logger   *zlogger.Zlogrus
	handlers map[string]EventHandler
	sync.RWMutex
}

func NewManager(logr *zlogger.Zlogrus) *Manager {
	return &Manager{
		Clients:  make(ClientList),
		handlers: make(map[string]EventHandler),
		logger:   logr,
	}
}

// NOTE might have to update this to either wrap the actual handlers in the anom func or update the anom func to the handler itself
func (m *Manager) setupEventHandlers() {
	m.handlers[EventFetchSchedule] = func(ctx context.Context, clnt *Client, evnt Event) error {
		if err := HandleGetSchedule(ctx, clnt, evnt); err != nil {
			m.logger.MustDebug(err.Error())
			return err
		}
		return nil
	}

	m.handlers[EventCreateTask] = HandleCreateTask
	m.handlers[EventUpdateTask] = HandleEditTask
	m.handlers[EventRemoveTask] = HandleRemoveTask
}

func (m *Manager) routeEvent(event Event, clnt *Client) error {
	if handler, ok := m.handlers[event.Type]; ok {
		if err := handler(context.TODO(), clnt, event); err != nil {
			return err
		}
		// anything other than fetching needs to broadcast changes
		if event.Type != EventFetchSchedule {
			go m.BroadcastScheduleUpdate(clnt.CurrentPeriod)
		}
		return nil

	} else {
		return ErrEventNotSupported
	}
}

func (m *Manager) AddClient(client *Client) {
	m.Lock()
	defer m.Unlock()
	m.Clients[client.Connection] = client
}

func (m *Manager) RemoveClient(client *Client) {
	m.Lock()
	defer m.Unlock()
	defer delete(m.Clients, client.Connection)

	if _, ok := m.Clients[client.Connection]; ok {
		client.Connection.Close()
	}
}

func (m *Manager) SetClientPeriod(client *Client, period *models.Period) {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.Clients[client.Connection]; ok {
		m.Clients[client.Connection].SetPeriod(period)
	}
}

func (m *Manager) BroadcastScheduleUpdate(prd *models.Period) {
	result, err := controller.FetchSchedule(context.TODO(), prd.StartDate, prd.EndDate)

	if err != nil {
		m.logger.MustDebug(err.Error())
	}

	resposne, ok := result.(*models.ScheduleResponse)

	if !ok {
		m.logger.MustDebug(fmt.Sprintf("failed to type assert [%T] as type [%T]", result, models.ScheduleResponse{}))
	}

	evnt := dtos.NewEventDto(EventBroadcastSchedule, resposne)

	for conn, clnt := range m.Clients {
		if utils.IsInRange(clnt.CurrentPeriod.StartDate, clnt.CurrentPeriod.EndDate, prd.StartDate, prd.EndDate) {
			if err := conn.WriteJSON(evnt); err != nil {
				failedBroadCast := ErrFailedBroadcast{EffectedPeriodStart: prd.StartDate, EffectedPeriodEnd: prd.EndDate, Client: clnt, Err: err}
				m.logger.MustDebug(failedBroadCast.Error())
			}
		}
	}
}
