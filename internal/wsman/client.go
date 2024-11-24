package wsman

import (
	"github.com/Z3DRP/zportfolio-service/enums"
	"github.com/Z3DRP/zportfolio-service/internal/models"
	"github.com/Z3DRP/zportfolio-service/internal/zlogger"
	"github.com/gorilla/websocket"
)

type Client struct {
	Connection    *websocket.Conn
	Manager       *Manager
	CurrentPeriod *models.Period
	Logger        *zlogger.Zlogrus
	MessageQue    chan []Event
}

type ClientList map[*websocket.Conn]*Client

func NewClient(conn *websocket.Conn, manager *Manager, logr *zlogger.Zlogrus) *Client {
	return &Client{
		Connection: conn,
		Manager:    manager,
		Logger:     logr,
		MessageQue: make(chan []Event),
	}
}

func (c *Client) SetPeriod(p *models.Period) {
	c.CurrentPeriod = p
}

func (c *Client) ReadMessages() {
	defer func() {
		c.Manager.RemoveClient(c)
	}()

	for {
		var message Event
		err := c.Connection.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure); err != nil {
				c.Logger.MustDebug(err.Error())
			}
			break
		}
	}
}

func (c *Client) WriteMessages() {
	defer func() {
		c.Manager.RemoveClient(c)
	}()

	for {
		select {
		case message, ok := <-c.MessageQue:
			if !ok {
				if err := c.Connection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(enums.ServerError, "unable to write message")); err != nil {
					c.Logger.MustDebug(err.Error())
				}
				return
			}

			if err := c.Connection.WriteJSON(message); err != nil {
				c.Logger.MustDebug(err.Error())
			}
		}
	}
}
