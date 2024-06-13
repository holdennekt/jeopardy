package events

import (
	"encoding/json"

	"github.com/holdennekt/sgame/api/ws"
	"github.com/holdennekt/sgame/entities"
)

const CHAT ws.Event = "chat"

type chatMessage struct {
	Event   ws.Event    `json:"event"`
	Payload chatPayload `json:"payload"`
}

type chatPayload struct {
	From entities.User `json:"from"`
	incomingChatPayload
}

type incomingChatPayload struct {
	Text string `json:"text"`
}

func NewChatMessage(message ws.InternalMessage) (chatMessage, error) {
	var incomingPayload incomingChatPayload
	if err := json.Unmarshal(message.Payload, &incomingPayload); err != nil {
		return chatMessage{}, err
	}
	return chatMessage{
		Event: CHAT,
		Payload: chatPayload{
			incomingChatPayload: incomingPayload,
			From:                message.From,
		},
	}, nil
}

func (cm chatMessage) ToMessage() ws.Message {
	rawMessage, _ := json.Marshal(cm.Payload)
	return ws.Message{
		Event:   cm.Event,
		Payload: rawMessage,
	}
}

func HandleWsChatMessage(pubSubConn *ws.PubSubConn, msg ws.InternalMessage) {
	pubSubConn.Publish(msg)
}
