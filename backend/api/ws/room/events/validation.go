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

const VALIDATION ws.Event = "validation"

type ValidationMessage struct {
	Event   ws.Event          `json:"event"`
	Payload ValidationPayload `json:"payload"`
}

type ValidationPayload struct {
	IsCorrect bool `json:"isCorrect"`
}

func HandleRdsValidationMessage(rds *redis.Client, wsConn *ws.WsConn, pubSubConn *ws.PubSubConn, pack *entities.Pack, roomId primitive.ObjectID, msg ws.InternalMessage) {
	room, _ := entities.GetRoomById(rds, roomId)

	if room.FinalRoundState.IsActive || !room.IsUserHost(msg.From.Id) || room.AnsweringPlayer == nil {
		wsConn.PublishError(errors.New("can not validate"))
		return
	}

	var vp ValidationPayload
	if err := json.Unmarshal(msg.Payload, &vp); err != nil {
		wsConn.PublishError(err)
		return
	}

	currentQuestion := *room.CurrentQuestion

	err := api.TryUpdateRoom(rds, roomId, func(tx *redis.Tx) error {
		room, httpErr := entities.GetRoomById(rds, roomId)
		if httpErr != nil {
			return httpErr
		}

		playerIndex := slices.IndexFunc(room.Players, func(p entities.Player) bool {
			return *room.AnsweringPlayer == p.Id
		})
		if playerIndex == -1 {
			return errors.New("no such player in room")
		}

		roomKey := entities.GetRoomRedisKey(room.Id.Hex())
		_, err := tx.Pipelined(context.TODO(), func(p redis.Pipeliner) error {
			room.AnsweringPlayer = nil

			if vp.IsCorrect {
				room.Players[playerIndex].Score += room.CurrentQuestion.Value
			} else {
				room.Players[playerIndex].Score -= room.CurrentQuestion.Value
			}

			isEndOfQuestion := vp.IsCorrect || len(room.AllowedToAnswer) == 0
			if isEndOfQuestion {
				room.EndQuestion(pack)
			}

			p.JSONSet(context.TODO(), roomKey, "$", room)
			return nil
		})
		return err
	}, UPDATE_ROOM_RETRIES)
	if err != nil {
		wsConn.PublishError(err)
		return
	}

	correctAnswerMessage := NewCorrectAnswerInternalMessage(currentQuestion.Answers)
	if err := pubSubConn.Publish(correctAnswerMessage); err != nil {
		wsConn.PublishError(err)
		return
	}

	roomMessage := RoomInternalMessage()
	if err := pubSubConn.Publish(roomMessage); err != nil {
		wsConn.PublishError(err)
		return
	}
}
