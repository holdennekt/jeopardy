package events

import (
	"encoding/json"
	"time"

	"github.com/holdennekt/sgame/api/ws"
	"github.com/holdennekt/sgame/entities"
)

const DEADLINE ws.Event = "deadline"

// Sent by host of the room when current question was fully shown
// or sent by system to room participants, setting timer deadline
type DeadlineMessage struct {
	Event   ws.Event        `json:"event"`
	Payload QuestionPayload `json:"payload"`
}

type DeadlinePayload struct {
	Deadline time.Time `json:"deadline"`
}

func NewDeadlineInternalMessage(deadline time.Time) ws.InternalMessage {
	payload, _ := json.Marshal(deadline)
	return ws.InternalMessage{
		From: entities.SYSTEM,
		Message: ws.Message{
			Event:   DEADLINE,
			Payload: payload,
		},
	}
}

// func HandleRdsDeadlineMessage(rds *redis.Client, wsConn *ws.WsConn, pubSubConn *ws.PubSubConn, pack *entities.Pack, roomId primitive.ObjectID, msg ws.InternalMessage) {
// 	room, _ := entities.GetRoomById(rds, roomId)

// 	if !room.IsUserHost(msg.From.Id) || room.CurrentQuestion == nil && room.FinalRoundQuestion == nil {
// 		wsConn.PublishError(errors.New("can not set deadline"))
// 		return
// 	}

// 	var duration time.Duration
// 	if room.CurrentQuestion != nil {
// 		duration = time.Duration(room.Options.ThinkingTime) * time.Second
// 	} else {
// 		duration = time.Duration(room.Options.ThinkingTimeFinal) * time.Second
// 	}

// 	deadlineMessage := NewDeadlineInternalMessage(time.Now().Add(duration))

// 	time.AfterFunc(duration, func() {
// 		api.TryUpdateRoom(rds, roomId, func(tx *redis.Tx) error {
// 			newRoom, httpErr := entities.GetRoomById(rds, roomId)
// 			if httpErr != nil {
// 				return httpErr
// 			}

// 			triedToAnswer := *room.CurrentPlayer != *newRoom.CurrentPlayer
// 			questionChanged := newRoom.CurrentQuestion == nil || *newRoom.CurrentQuestion != *room.CurrentQuestion
// 			if triedToAnswer || questionChanged {
// 				return
// 			}

// 			roomKey := entities.GetRoomRedisKey(room.Id.Hex())
// 			_, err := tx.Pipelined(context.TODO(), func(p redis.Pipeliner) error {
// 				newRoom.EndQuestion(pack)

// 				p.JSONSet(context.TODO(), roomKey, "$", newRoom)
// 				return nil
// 			})
// 			return err
// 		}, UPDATE_ROOM_RETRIES)
// 	})
// }
