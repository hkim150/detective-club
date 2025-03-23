package model

import (
	"log"

	"github.com/gorilla/websocket"
)

type client struct {
	id      string
	socket  *websocket.Conn
	room    *room
	receive chan []byte
}

func (c *client) readFromSocket() {
	defer c.socket.Close()
	for {
		_, msg, err := c.socket.ReadMessage()
		if err != nil {
			log.Println("readFromSocket:", err)
			return
		}
		c.room.hub <- msg
	}
}

func (c *client) writeToSocket() {
	defer c.socket.Close()
	for msg := range c.receive {
		if err := c.socket.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Println("writeToSocket:", err)
			return
		}
	}
}
