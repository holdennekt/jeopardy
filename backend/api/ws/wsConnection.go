package ws

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/holdennekt/sgame/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Event string

type Message struct {
	Event   Event           `json:"event"`
	Payload json.RawMessage `json:"payload"`
}

type Messageble interface {
	ToMessage() Message
}

type InternalMessage struct {
	From entities.User `json:"from"`
	Message
}

type WsConn struct {
	userId       primitive.ObjectID
	Conn         *websocket.Conn
	Messages     <-chan InternalMessage
	Publish      func(message Message) error
	PublishError func(err error) error
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func ConnectUserToWs(c *gin.Context, user entities.User) (*WsConn, error) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return nil, err
	}

	wc := &WsConn{
		userId:   user.Id,
		Conn:     conn,
		Messages: getMesasgesChannel(conn, user),
		Publish: func(message Message) error {
			return conn.WriteJSON(message)
		},
		PublishError: func(err error) error {
			return conn.WriteJSON(NewErrorMessage(err))
		},
	}

	return wc, nil
}

func getMesasgesChannel(conn *websocket.Conn, user entities.User) <-chan InternalMessage {
	messages := make(chan InternalMessage)
	go func() {
		for {
			var msg Message
			_, r, err := conn.NextReader()
			if err != nil {
				log.Printf("User \"%s\" has closed the connection", user.Id)
				break
			}
			if err := json.NewDecoder(r).Decode(&msg); err != nil {
				log.Println("Error while decoding incoming wsMessage:", err)
				continue
			}
			messages <- InternalMessage{Message: msg, From: user}
		}
		conn.Close()
		close(messages)
	}()
	return messages
}
