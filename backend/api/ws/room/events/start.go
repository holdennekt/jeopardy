package events

import (
	"context"
	"errors"
	"math/rand"

	"github.com/holdennekt/sgame/api"
	"github.com/holdennekt/sgame/api/ws"
	"github.com/holdennekt/sgame/entities"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const START ws.Event = "start"

type StartMessage struct {
	Event ws.Event `json:"event"`
}

func HandleRdsStartMessage(rds *redis.Client, wsConn *ws.WsConn, pubSubConn *ws.PubSubConn, pack *entities.Pack, roomId primitive.ObjectID, msg ws.InternalMessage) {
	err := api.TryUpdateRoom(rds, roomId, func(tx *redis.Tx) error {
		room, httpErr := entities.GetRoomById(rds, roomId)
		if httpErr != nil {
			return httpErr
		}

		if !room.IsUserHost(msg.From.Id) || len(room.Players) == 0 {
			return errors.New("not allowed to start game")
		}

		room.StartNextRound(pack)
		room.CurrentPlayer = &room.Players[rand.Intn(len(room.Players))].Id

		_, err := tx.Pipelined(context.TODO(), func(p redis.Pipeliner) error {
			roomKey := entities.GetRoomRedisKey(roomId.Hex())
			p.JSONSet(context.TODO(), roomKey, "$", room)
			return nil
		})
		return err
	}, UPDATE_ROOM_RETRIES)
	if err != nil {
		wsConn.PublishError(err)
		return
	}

	roomMessage := RoomInternalMessage()
	if err := pubSubConn.Publish(roomMessage); err != nil {
		wsConn.PublishError(err)
		return
	}
}
