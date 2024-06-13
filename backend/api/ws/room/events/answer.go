package events

import (
	"context"
	"encoding/json"
	"errors"
	"slices"

	"github.com/holdennekt/sgame/api"
	"github.com/holdennekt/sgame/api/ws"
	"github.com/holdennekt/sgame/entities"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const ANSWER ws.Event = "answer"

type AnswerMessage struct {
	Event ws.Event `json:"event"`
}

func HandleRdsAnswerMessage(rds *redis.Client, wsConn *ws.WsConn, pubSubConn *ws.PubSubConn, roomId primitive.ObjectID, msg ws.InternalMessage) {
	room, _ := entities.GetRoomById(rds, roomId)

	if room.FinalRoundState.IsActive {
		wsConn.PublishError(errors.New("not allowed to answer"))
		return
	}

	err := api.TryUpdateRoom(rds, roomId, func(tx *redis.Tx) error {
		room, httpErr := entities.GetRoomById(rds, roomId)
		if httpErr != nil {
			return httpErr
		}

		if room.AnsweringPlayer != nil || !slices.Contains(room.AllowedToAnswer, msg.From.Id) {
			return errors.New("not allowed to answer")
		}

		room.AnsweringPlayer = &msg.From.Id
		room.CurrentPlayer = &msg.From.Id
		room.AllowedToAnswer = slices.DeleteFunc(room.AllowedToAnswer, func(playerId primitive.ObjectID) bool {
			return msg.From.Id == playerId
		})

		_, err := tx.Pipelined(context.TODO(), func(p redis.Pipeliner) error {
			roomKey := entities.GetRoomRedisKey(room.Id.Hex())
			p.JSONSet(context.TODO(), roomKey, "$.answeringPlayer", room.AnsweringPlayer)
			p.JSONSet(context.TODO(), roomKey, "$.currentPlayer", room.CurrentPlayer)
			p.JSONSet(context.TODO(), roomKey, "$.allowedToAnswer", room.AllowedToAnswer)
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

const CORRECT_ANSWER ws.Event = "answer"

type CorrectAnswerMessage struct {
	Event   ws.Event `json:"event"`
	Payload []string `json:"payload"`
}

func NewCorrectAnswerInternalMessage(answers []string) ws.InternalMessage {
	payload, _ := json.Marshal(answers)
	return ws.InternalMessage{
		From: entities.SYSTEM,
		Message: ws.Message{
			Event:   CORRECT_ANSWER,
			Payload: payload,
		},
	}
}
