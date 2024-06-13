package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type HostRoom struct {
	Id                 primitive.ObjectID   `json:"id"`
	Name               string               `json:"name"`
	PackPreview        PackPreview          `json:"packPreview"`
	Host               *Host                `json:"host"`
	Players            []Player             `json:"players"`
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

func NewHostRoom(room *Room) HostRoom {
	return HostRoom{
		Id:                 room.Id,
		Name:               room.Name,
		PackPreview:        room.PackPreview,
		Host:               room.Host,
		Players:            room.Players,
		CurrentRound:       room.CurrentRound,
		AvailableQuestions: room.AvailableQuestions,
		CurrentPlayer:      room.CurrentPlayer,
		CurrentQuestion:    room.CurrentQuestion,
		AnsweringPlayer:    room.AnsweringPlayer,
		AllowedToAnswer:    room.AllowedToAnswer,
		FinalRoundState:    room.FinalRoundState,
		DeadlineAt:         room.DeadlineAt,
		PausedState:        room.PausedState,
	}
}
