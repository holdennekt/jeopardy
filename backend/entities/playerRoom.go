package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PlayerRoom struct {
	Id                 primitive.ObjectID    `json:"id"`
	Name               string                `json:"name"`
	PackPreview        PackPreview           `json:"packPreview"`
	Host               *Host                 `json:"host"`
	Players            []Player              `json:"players"`
	CurrentRound       *string               `json:"currentRound"`
	AvailableQuestions AvailableQuestions    `json:"availableQuestions"`
	CurrentPlayer      *primitive.ObjectID   `json:"currentPlayer"`
	CurrentQuestion    *HiddenQuestion       `json:"currentQuestion"`
	AnsweringPlayer    *primitive.ObjectID   `json:"answeringPlayer"`
	AllowedToAnswer    []primitive.ObjectID  `json:"allowedToAnswer"`
	FinalRoundState    hiddenFinalRoundState `json:"finalRoundState"`
	DeadlineAt         time.Time             `json:"deadlineAt"`
	PausedState        PausedState           `json:"pausedState"`
}

func NewPlayerRoom(room *Room) PlayerRoom {
	var currentQuestion *HiddenQuestion
	if room.CurrentQuestion == nil {
		currentQuestion = nil
	} else {
		currentQuestion = &HiddenQuestion{
			Index:      room.CurrentQuestion.Index,
			Value:      room.CurrentQuestion.Value,
			Text:       room.CurrentQuestion.Text,
			Attachment: room.CurrentQuestion.Attachment,
		}
	}
	var finalQuestion *HiddenFinalQuestion
	if room.FinalRoundState.Question == nil {
		finalQuestion = nil
	} else {
		finalQuestion = &HiddenFinalQuestion{
			Text:       room.FinalRoundState.Question.Text,
			Attachment: room.FinalRoundState.Question.Attachment,
		}
	}
	return PlayerRoom{
		Id:                 room.Id,
		Name:               room.Name,
		Players:            room.Players,
		Host:               room.Host,
		CurrentRound:       room.CurrentRound,
		AvailableQuestions: room.AvailableQuestions,
		CurrentPlayer:      room.CurrentPlayer,
		CurrentQuestion:    currentQuestion,
		AnsweringPlayer:    room.AnsweringPlayer,
		AllowedToAnswer:    room.AllowedToAnswer,
		FinalRoundState: hiddenFinalRoundState{
			IsActive:           room.FinalRoundState.IsActive,
			AvailableQuestions: room.FinalRoundState.AvailableQuestions,
			Question:           finalQuestion,
		},
		DeadlineAt:  room.DeadlineAt,
		PausedState: room.PausedState,
	}
}
