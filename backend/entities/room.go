package entities

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/holdennekt/sgame/custErrors"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const ROOM_PREFIX = "room:"
const ANSWERING_TIME = 5 * time.Second

type Room struct {
	Id primitive.ObjectID `json:"id"`
	RoomDTO
	PackPreview        PackPreview          `json:"packPreview"`
	CreatedBy          primitive.ObjectID   `json:"createdBy"`
	Host               *Host                `json:"host"`
	Players            []Player             `json:"players"`
	BanList            []User               `json:"banList"`
	CurrentRound       *string              `json:"currentRound"`
	AvailableQuestions AvailableQuestions   `json:"availableQuestions"`
	CurrentPlayer      *primitive.ObjectID  `json:"currentPlayer"`
	CurrentQuestion    *Question            `json:"currentQuestion"`
	AnsweringPlayer    *primitive.ObjectID  `json:"answeringPlayer"`
	AllowedToAnswer    []primitive.ObjectID `json:"allowedToAnswer"`
	FinalRoundState    FinalRoundState      `json:"finalRoundState"`
	DeadlineAt         time.Time            `json:"deadlineAt"`
	PausedState        PausedState          `json:"pausedState"`
}

type RoomDTO struct {
	Name    string             `json:"name" binding:"min=1,max=50"`
	PackId  primitive.ObjectID `json:"packId" binding:"required"`
	Options roomOptions        `json:"options"`
}

type roomOptions struct {
	MaxPlayers          int         `json:"maxPlayers" binding:"min=1,max=10"`
	Type                PrivacyType `json:"type" binding:"oneof=public private"`
	Password            *string     `json:"password" binding:"omitnil,min=4,max=16"`
	ThinkingTime        int         `json:"thinkingTime" binding:"min=1,max=30"`
	ThinkingTimeFinal   int         `json:"thinkingTimeFinal" binding:"min=1,max=120"`
	IsFalseStartAllowed bool        `json:"isFalseStartAllowed"`
}

type PrivacyType string

const (
	Public  PrivacyType = "public"
	Private PrivacyType = "private"
)

type FinalRoundState struct {
	IsActive           bool            `json:"isActive"`
	AvailableQuestions map[string]bool `json:"availableQuestions"`
	Question           *FinalQuestion  `json:"question"`
	Players            []FinalPlayer   `json:"players"`
}

type hiddenFinalRoundState struct {
	IsActive           bool                 `json:"isActive"`
	AvailableQuestions map[string]bool      `json:"availableQuestions"`
	Question           *HiddenFinalQuestion `json:"question"`
	Players            []FinalPlayer        `json:"players"`
}

type FinalPlayer struct {
	PlayerId  primitive.ObjectID `json:"playerId"`
	BetAmount int                `json:"amount"`
	HasBet    bool               `json:"isDone"`
	Answer    *string            `json:"answer"`
}

type AvailableQuestions map[string][]BoardQuestion
type BoardQuestion struct {
	Index         int  `json:"index"`
	Value         int  `json:"value"`
	HasBeenPlayed bool `json:"hasBeenPlayed"`
}

type PausedState struct {
	IsPaused bool      `json:"isPaused"`
	PausedAt time.Time `json:"pausedAt"`
}

func (r *Room) IsUserHost(userId primitive.ObjectID) bool {
	if r.Host == nil {
		return false
	}
	return userId == r.Host.Id
}

func (r *Room) IsUserPlayer(userId primitive.ObjectID) bool {
	return slices.ContainsFunc(r.Players, func(player Player) bool {
		return userId == player.Id
	})
}

func (r *Room) IsUserIn(userId primitive.ObjectID) bool {
	return r.IsUserHost(userId) || r.IsUserPlayer(userId)
}

func (r *Room) AnyAvailableQuestions() bool {
	for _, questions := range r.AvailableQuestions {
		notPlayedIndex := slices.IndexFunc(questions, func(bq BoardQuestion) bool {
			return !bq.HasBeenPlayed
		})
		if notPlayedIndex != -1 {
			return true
		}
	}
	return false
}

func (r *Room) GetProjection(userId primitive.ObjectID) any {
	if r.IsUserHost(userId) {
		return NewHostRoom(r)
	}
	return NewPlayerRoom(r)
}

func (r *Room) EndQuestion(pack *Pack) {
	r.CurrentQuestion = nil
	r.AllowedToAnswer = make([]primitive.ObjectID, 0)
	if !r.AnyAvailableQuestions() {
		r.StartNextRound(pack)
	}
}

func (r *Room) StartNextRound(pack *Pack) {
	currentRoundIndex := slices.IndexFunc(pack.Rounds, func(round Round) bool {
		return *r.CurrentRound == round.Name
	})
	nextRoundIndex := currentRoundIndex + 1
	if nextRoundIndex < len(pack.Rounds) {
		nextRound := pack.Rounds[nextRoundIndex]
		r.CurrentRound = &nextRound.Name
		r.InitAvailableQuestions(nextRound)
	} else {
		r.startFinalRound(pack)
	}
}

func (r *Room) startFinalRound(pack *Pack) {
	r.CurrentRound = nil
	r.FinalRoundState.IsActive = true
	r.InitAvailableFinalQuestions(pack.FinalRound)
	r.AllowedToAnswer = make([]primitive.ObjectID, 0)
	r.FinalRoundState.Players = make([]FinalPlayer, 0)
	for _, player := range r.Players {
		if player.Score > 0 {
			r.AllowedToAnswer = append(r.AllowedToAnswer, player.Id)
			r.FinalRoundState.Players = append(r.FinalRoundState.Players, FinalPlayer{
				PlayerId: player.Id,
			})
		}
	}
}

func (r *Room) InitAvailableQuestions(round Round) {
	availableQuestions := make(AvailableQuestions)
	for _, category := range round.Categories {
		categoryAvailableQuestions := make([]BoardQuestion, 0)
		for _, question := range category.Questions {
			categoryAvailableQuestions = append(categoryAvailableQuestions, BoardQuestion{
				Index:         question.Index,
				Value:         question.Value,
				HasBeenPlayed: false,
			})
		}
		availableQuestions[category.Name] = categoryAvailableQuestions
	}
	r.AvailableQuestions = availableQuestions
}

func (r *Room) InitAvailableFinalQuestions(finalRound FinalRound) {
	availableQuestions := make(map[string]bool)
	for _, finalCategory := range finalRound.Categories {
		availableQuestions[finalCategory.Name] = true
	}
	r.FinalRoundState.AvailableQuestions = availableQuestions
}

func GetRoomByKey(rds *redis.Client, key string) (*Room, custErrors.HttpError) {
	var room Room

	res, err := rds.JSONGet(context.TODO(), key).Result()
	if err != nil {
		return nil, custErrors.NewInternalError(err)
	}

	if res == "" {
		return nil, custErrors.NewHttpError(
			http.StatusNotFound,
			gin.H{"error": "no room with such id"},
		)
	}

	if err := json.Unmarshal([]byte(res), &room); err != nil {
		return nil, custErrors.NewInternalError(err)
	}

	return &room, nil
}

func GetRoomById(rds *redis.Client, id primitive.ObjectID) (*Room, custErrors.HttpError) {
	return GetRoomByKey(rds, GetRoomRedisKey(id.Hex()))
}

func GetRoomRedisKey(id string) string {
	return ROOM_PREFIX + id
}
