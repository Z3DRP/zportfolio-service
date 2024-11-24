package wsman

import (
	"net/http"
	"sync"

	"github.com/Z3DRP/zportfolio-service/config"
	"github.com/Z3DRP/zportfolio-service/internal/models"
	"github.com/gorilla/websocket"
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

type Manager struct {
	Clients ClientList
	sync.RWMutex
	handlers map[string]EventHandler
}

func NewManager() *Manager {
	return &Manager{
		Clients: make(ClientList),
	}
}

func (m *Manager) setupEventHandlers() {
	m.handlers[Event]
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
