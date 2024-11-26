package wsman

import (
	"fmt"
	"time"

	"github.com/Z3DRP/zportfolio-service/enums"
	"github.com/Z3DRP/zportfolio-service/internal/models"
	"github.com/gorilla/websocket"
)

type Client struct {
	Connection    *websocket.Conn
	Manager       *Manager
	CurrentPeriod *models.Period
	MessageQue    chan Event
}

type ClientList map[*websocket.Conn]*Client

func NewClient(conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		Connection: conn,
		Manager:    manager,
		MessageQue: make(chan Event),
	}
}

func (c *Client) SetPeriod(p *models.Period) {
	c.CurrentPeriod = p
}

func (c *Client) ReadMessages() {
	defer func() {
		c.Manager.RemoveClient(c)
	}()

	c.Connection.SetReadLimit(512)
	c.Connection.SetPongHandler(c.pongHandler)

	if err := c.Connection.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		c.Manager.logger.MustDebug(fmt.Sprintf("set ws read deadline err: %v", err))
		return
	}

	for {
		var message Event
		err := c.Connection.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.Manager.logger.MustDebug(err.Error())
			}
			break
		}

		if err = c.Manager.routeEvent(message, c); err != nil {
			c.Manager.logger.MustDebug(err.Error())
		}
	}
}

func (c *Client) WriteMessages() {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		c.Manager.RemoveClient(c)
	}()

	for {
		select {
		case message, ok := <-c.MessageQue:
			if !ok {
				if err := c.Connection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(enums.ServerError, "unable to write message")); err != nil {
					c.Manager.logger.MustDebug(err.Error())
				}
				return
			}

			if err := c.Connection.WriteJSON(message); err != nil {
				c.Manager.logger.MustDebug(err.Error())
			}
			c.Manager.logger.MustDebug("message written successfully")
		case <-ticker.C:
			c.Manager.logger.MustDebug("ping")

			if err := c.Connection.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				c.Manager.logger.MustDebug(fmt.Sprintf("ping write msg failed: %v", err))
				return
			}
		}
	}
}

func (c *Client) pongHandler(pongMsg string) error {
	c.Manager.logger.MustDebug("pong")
	// this keeps the connection alive everytime a pong is recieved
	return c.Connection.SetReadDeadline(time.Now().Add(pongWait))
}
