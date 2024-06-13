package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"github.com/holdennekt/sgame/api/ws"
	"github.com/holdennekt/sgame/entities"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const QUESTION ws.Event = "question"

type QuestionMessage struct {
	Event   ws.Event        `json:"event"`
	Payload QuestionPayload `json:"payload"`
}

type QuestionPayload struct {
	Category string `json:"category"`
	Index    int    `json:"index"`
}

func HandleRdsQuestionMessage(rds *redis.Client, wsConn *ws.WsConn, pubSubConn *ws.PubSubConn, pack *entities.Pack, roomId primitive.ObjectID, msg ws.InternalMessage) {
	room, _ := entities.GetRoomById(rds, roomId)
	if room.CurrentRound == nil || room.AvailableQuestions == nil || room.CurrentPlayer == nil || *room.CurrentPlayer != msg.From.Id {
		wsConn.PublishError(errors.New("not allowed to choose"))
		return
	}

	var qp QuestionPayload
	if err := json.Unmarshal(msg.Payload, &qp); err != nil {
		wsConn.PublishError(err)
		return
	}

	boardQuestionIndex := slices.IndexFunc(room.AvailableQuestions[qp.Category], func(bq entities.BoardQuestion) bool {
		return bq.Index == qp.Index
	})
	if room.AvailableQuestions[qp.Category] == nil || boardQuestionIndex == -1 {
		wsConn.PublishError(errors.New("no such question in current round"))
		return
	}
	if !room.AvailableQuestions[qp.Category][boardQuestionIndex].HasBeenPlayed {
		wsConn.PublishError(errors.New("question has already been played"))
		return
	}

	roundIndex := slices.IndexFunc(pack.Rounds, func(r entities.Round) bool {
		return *room.CurrentRound == r.Name
	})
	round := pack.Rounds[roundIndex]
	categoryIndex := slices.IndexFunc(round.Categories, func(c entities.Category) bool {
		return qp.Category == c.Name
	})
	category := round.Categories[categoryIndex]
	questionIndex := slices.IndexFunc(category.Questions, func(q entities.Question) bool {
		return qp.Index == q.Index
	})
	question := category.Questions[questionIndex]
	room.CurrentQuestion = &question

	allowedToAnswer := make([]primitive.ObjectID, len(room.Players))
	for i, player := range room.Players {
		allowedToAnswer[i] = player.Id
	}
	room.AllowedToAnswer = allowedToAnswer

	room.AvailableQuestions[qp.Category][boardQuestionIndex].HasBeenPlayed = true

	roomKey := entities.GetRoomRedisKey(room.Id.Hex())
	_, err := rds.TxPipelined(context.TODO(), func(p redis.Pipeliner) error {
		p.JSONSet(context.TODO(), roomKey, "$.currentQuestion", room.CurrentQuestion)
		p.JSONSet(context.TODO(), roomKey, "$.allowedToAnswer", room.AllowedToAnswer)
		path := fmt.Sprintf("$.availableQuestions.%s", qp.Category)
		p.JSONSet(context.TODO(), roomKey, path, room.AvailableQuestions[qp.Category])
		return nil
	})
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
