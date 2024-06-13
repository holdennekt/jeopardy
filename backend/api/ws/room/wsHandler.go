package wsRoom

import (
	"context"
	"slices"
	"time"

	"github.com/holdennekt/sgame/api"
	"github.com/holdennekt/sgame/api/ws"
	wsLobby "github.com/holdennekt/sgame/api/ws/lobby"
	lobbyEvents "github.com/holdennekt/sgame/api/ws/lobby/events"
	roomEvents "github.com/holdennekt/sgame/api/ws/room/events"
	"github.com/holdennekt/sgame/entities"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const UPDATE_ROOM_RETRIES = 3

func handleWsMessage(mdb *mongo.Database, rds *redis.Client, wsConn *ws.WsConn, pubSubConn *ws.PubSubConn, pack *entities.Pack, roomId primitive.ObjectID, msg ws.InternalMessage) {
	switch msg.Event {
	case lobbyEvents.CHAT:
		lobbyEvents.HandleWsChatMessage(pubSubConn, msg)
	case roomEvents.QUESTION:
		roomEvents.HandleRdsQuestionMessage(rds, wsConn, pubSubConn, pack, roomId, msg)
	case roomEvents.ANSWER:
		roomEvents.HandleRdsAnswerMessage(rds, wsConn, pubSubConn, roomId, msg)
	case roomEvents.VALIDATION:
		roomEvents.HandleRdsValidationMessage(rds, wsConn, pubSubConn, pack, roomId, msg)
	}
}

func handleWsClosure(rds *redis.Client, pubSubConn *ws.PubSubConn, userId primitive.ObjectID, roomId primitive.ObjectID) {
	roomKey := entities.GetRoomRedisKey(roomId.Hex())
	room, _ := entities.GetRoomByKey(rds, roomKey)

	err := api.TryUpdateRoom(rds, room.Id, func(tx *redis.Tx) error {
		room, httpErr := entities.GetRoomByKey(rds, roomKey)
		if httpErr != nil {
			return httpErr
		}

		isGameStarted := room.CurrentRound == nil && !room.FinalRoundState.IsActive
		if isGameStarted {
			if room.IsUserHost(userId) {
				room.Host = nil
			} else {
				room.Players = slices.DeleteFunc(room.Players, func(p entities.Player) bool {
					return userId == p.Id
				})
			}
		} else {
			if room.IsUserHost(userId) {
				room.Host.IsConnected = false
			} else {
				i := slices.IndexFunc(room.Players, func(p entities.Player) bool {
					return userId == p.Id
				})
				room.Players[i].IsConnected = false
			}
		}

		connectedPlayerIndex := slices.IndexFunc(room.Players, func(p entities.Player) bool {
			return p.IsConnected
		})
		isAnyConnectedUser := room.Host.IsConnected || connectedPlayerIndex != -1

		_, err := tx.Pipelined(context.TODO(), func(p redis.Pipeliner) error {
			if room.IsUserHost(userId) {
				p.JSONSet(context.TODO(), roomKey, "$.host", room.Host)
			} else {
				p.JSONSet(context.TODO(), roomKey, "$.players", room.Players)
			}
			if !isAnyConnectedUser {
				p.Expire(context.TODO(), roomKey, 5*time.Minute)
			}
			return nil
		})
		return err
	}, UPDATE_ROOM_RETRIES)
	if err != nil {
		return
	}

	roomMessage := roomEvents.RoomInternalMessage()
	pubSubConn.Publish(roomMessage)

	lobbyRoomMessage := lobbyEvents.NewLobbyRoomInternalMessage(room)
	ws.PublishRdsMessage(rds, wsLobby.LOBBY, lobbyRoomMessage)

	pubSubConn.Conn.Close()
}
