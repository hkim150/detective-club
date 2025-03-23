package model

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  socketBufferSize,
		WriteBufferSize: socketBufferSize,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
)

type room struct {
	clients      map[string]*client
	join         chan *client
	leave        chan *client
	hub          chan []byte
	activePlayer string
	conspirator  string
}

func (r *room) Run() {
	for {
		select {
		case client := <-r.join:
			r.clients[client.id] = client
			resMsg := message{
				Purpose: "inform-num-players",
				Content: strconv.Itoa(len(r.clients)),
			}
			b, err := json.Marshal(resMsg)
			if err != nil {
				log.Println("failed to marshal message: ", err)
				continue
			}
			for _, c := range r.clients {
				c.receive <- b
			}
		case client := <-r.leave:
			delete(r.clients, client.id)
			close(client.receive)
			resMsg := message{
				Purpose: "inform-num-players",
				Content: strconv.Itoa(len(r.clients)),
			}
			b, err := json.Marshal(resMsg)
			if err != nil {
				log.Println("failed to marshal message: ", err)
				continue
			}
			for _, c := range r.clients {
				c.receive <- b
			}
		case msg := <-r.hub:
			var msgObj message
			err := json.Unmarshal(msg, &msgObj)
			if err != nil {
				log.Println("failed to unmarshal message: ", err)
				continue
			}

			switch msgObj.Purpose {
			case "claim-active-player":
				r.activePlayer = msgObj.Content
				candidates := make([]string, 0, len(r.clients)-1)
				for id := range r.clients {
					if id != r.activePlayer {
						candidates = append(candidates, id)
					}
				}
				if len(candidates) > 1 {
					r.conspirator = candidates[rand.Intn(len(candidates))]
				}
				resMsg := message{
					Purpose: "set-active-player",
					Content: r.activePlayer,
				}
				b, err := json.Marshal(resMsg)
				if err != nil {
					log.Println("failed to marshal message: ", err)
					continue
				}
				for _, client := range r.clients {
					client.receive <- b
				}
			case "unclaim-active-player":
				r.activePlayer = ""
				r.conspirator = ""
				resMsg := message{
					Purpose: "set-active-player",
					Content: r.activePlayer,
				}
				b, err := json.Marshal(resMsg)
				if err != nil {
					log.Println("failed to marshal message: ", err)
					continue
				}
				for _, client := range r.clients {
					client.receive <- b
				}
			case "give-clue":
				for id, client := range r.clients {
					if id != r.conspirator {
						client.receive <- msg
					}
				}

				conspiratorMsg := message{
					Purpose: "give-clue",
					Content: "You are the conspirator. Good luck!",
				}

				b, err := json.Marshal(conspiratorMsg)
				if err != nil {
					log.Println("failed to marshal message: ", err)
					continue
				}
				if conspiratorClient, ok := r.clients[r.conspirator]; ok {
					conspiratorClient.receive <- b
				}
				// TODO: if conspirator is not in the room, reset the game
			}
		}
	}
}

func NewRoom() *room {
	return &room{
		clients: make(map[string]*client),
		join:    make(chan *client),
		leave:   make(chan *client),
		hub:     make(chan []byte),
	}
}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHttp: ", err)
		return
	}

	client := &client{
		id:      uuid.New().String(),
		socket:  socket,
		receive: make(chan []byte, messageBufferSize),
		room:    r,
	}

	msg := message{
		Purpose: "set-id",
		Content: client.id,
	}
	b, err := json.Marshal(msg)
	if err != nil {
		log.Println("ServeHttp: ", err)
	}
	client.receive <- b
	r.join <- client

	defer func() {
		r.leave <- client
	}()

	go client.writeToSocket()
	client.readFromSocket()
}
