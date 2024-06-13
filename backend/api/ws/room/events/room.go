package events

import (
	"github.com/holdennekt/sgame/api/ws"
	"github.com/holdennekt/sgame/entities"
)

const UPDATE_ROOM_RETRIES = 3

const ROOM ws.Event = "room"

type RoomMessage struct {
	Event ws.Event `json:"event"`
}

func RoomInternalMessage() ws.InternalMessage {
	return ws.InternalMessage{
		From: entities.SYSTEM,
		Message: ws.Message{
			Event: ROOM,
		},
	}
}
