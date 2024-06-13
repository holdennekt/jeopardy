package entities

import "go.mongodb.org/mongo-driver/bson/primitive"

type LobbyRoom struct {
	Id          primitive.ObjectID `json:"id"`
	Name        string             `json:"name"`
	PackPreview PackPreview        `json:"packPreview"`
	Host        *Host              `json:"host"`
	Players     []Player           `json:"players"`
	MaxPlayers  int                `json:"maxPlayers"`
	Type        PrivacyType        `json:"type"`
	Status      string             `json:"status"`
}

func NewLobbyRoom(room *Room) LobbyRoom {
	var status string
	if room.CurrentRound != nil || room.FinalRoundState.IsActive {
		status = "Playing"
	} else {
		status = "Idle"
	}
	lr := LobbyRoom{
		Id:          room.Id,
		Name:        room.Name,
		PackPreview: room.PackPreview,
		Host:        room.Host,
		Players:     room.Players,
		MaxPlayers:  room.Options.MaxPlayers,
		Type:        room.Options.Type,
		Status:      status,
	}

	return lr
}
