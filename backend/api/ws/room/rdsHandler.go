package wsRoom

import (
	"encoding/json"

	"github.com/holdennekt/sgame/api/ws"
	"github.com/holdennekt/sgame/api/ws/room/events"
	"github.com/holdennekt/sgame/entities"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func handleRdsMessage(mdb *mongo.Database, rds *redis.Client, wsConn *ws.WsConn, pubSubConn *ws.PubSubConn, pack *entities.Pack, room *entities.Room, userId primitive.ObjectID, msg ws.InternalMessage) {
	switch msg.Event {
	case events.ROOM:
		handleRdsRoomMessage(rds, wsConn, room.Id, userId)
	}
}

func handleRdsRoomMessage(rds *redis.Client, wsConn *ws.WsConn, roomId primitive.ObjectID, userId primitive.ObjectID) {
	room, _ := entities.GetRoomById(rds, roomId)
	payload, _ := json.Marshal(room.GetProjection(userId))
	wsConn.Publish(ws.Message{Event: events.ROOM, Payload: payload})
}
