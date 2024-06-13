package events

import (
	"encoding/json"

	"github.com/holdennekt/sgame/api/ws"
	"github.com/holdennekt/sgame/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	LOBBY_ROOM   ws.Event = "lobby-room"
	ROOM_DELETED ws.Event = "room-deleted"
)

type LobbyRoomMessage struct {
	Event   ws.Event           `json:"event"`
	Payload entities.LobbyRoom `json:"payload"`
}

func NewLobbyRoomInternalMessage(room *entities.Room) ws.InternalMessage {
	payload, _ := json.Marshal(entities.NewLobbyRoom(room))
	return ws.InternalMessage{
		From: entities.SYSTEM,
		Message: ws.Message{
			Event:   LOBBY_ROOM,
			Payload: payload,
		},
	}
}

type RoomDeletedPayload struct {
	Id primitive.ObjectID `json:"id"`
}

type RoomDeletedMessage struct {
	Event   ws.Event           `json:"event"`
	Payload RoomDeletedPayload `json:"payload"`
}

func NewRoomDeletedInternalMessage(id primitive.ObjectID) ws.InternalMessage {
	payload, _ := json.Marshal(RoomDeletedPayload{Id: id})
	return ws.InternalMessage{
		From: entities.SYSTEM,
		Message: ws.Message{
			Event:   ROOM_DELETED,
			Payload: payload,
		},
	}
}
