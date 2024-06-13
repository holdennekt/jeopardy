package wsRoom

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/holdennekt/sgame/api"
	"github.com/holdennekt/sgame/api/ws"
	"github.com/holdennekt/sgame/api/ws/room/events"
	"github.com/holdennekt/sgame/custErrors"
	"github.com/holdennekt/sgame/entities"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func ConnectHandler(mdb *mongo.Database, rds *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.MustGet(api.USER_ID_CONTEXT_KEY).(primitive.ObjectID)

		user, httpErr := entities.GetUser(mdb, userId)
		if httpErr != nil {
			custErrors.AbortWithError(c, httpErr)
			return
		}

		roomKey := entities.GetRoomRedisKey(c.Param("id"))
		room, httpErr := entities.GetRoomByKey(rds, roomKey)
		if httpErr != nil {
			custErrors.AbortWithError(c, httpErr)
			return
		}

		if !room.IsUserIn(userId) {
			custErrors.AbortWithError(c, custErrors.NewHttpError(
				http.StatusForbidden,
				gin.H{"error": "you are not in the room"},
			))
			return
		}

		pack, httpErr := entities.GetPack(mdb, room.PackId)
		if httpErr != nil {
			custErrors.AbortWithError(c, httpErr)
			return
		}

		wsConn, err := ws.ConnectUserToWs(c, *user)
		if err != nil {
			custErrors.AbortWithInternalError(c, err)
			return
		}
		log.Printf("User \"%s\" has connected to ws\n", userId)

		pubSubConn := ws.ConnectUserToPubSub(rds, userId, roomKey)

		err = api.TryUpdateRoom(rds, room.Id, func(tx *redis.Tx) error {
			room, httpErr := entities.GetRoomByKey(rds, roomKey)
			if httpErr != nil {
				return httpErr
			}

			_, err := tx.Pipelined(context.TODO(), func(p redis.Pipeliner) error {
				if room.IsUserHost(userId) {
					room.Host.IsConnected = true
					p.JSONSet(context.TODO(), roomKey, "$.host", room.Host)
				} else {
					i := slices.IndexFunc(room.Players, func(p entities.Player) bool {
						return userId == p.Id
					})
					room.Players[i].IsConnected = true
					p.JSONSet(context.TODO(), roomKey, "$.players", room.Players)
				}
				p.Persist(context.TODO(), roomKey)
				return nil
			})
			return err
		}, UPDATE_ROOM_RETRIES)
		if err != nil {
			wsConn.PublishError(err)
			wsConn.Conn.Close()
			pubSubConn.Conn.Close()
			return
		}

		roomMessage := events.RoomInternalMessage()
		if err := pubSubConn.Publish(roomMessage); err != nil {
			wsConn.PublishError(err)
			wsConn.Conn.Close()
			pubSubConn.Conn.Close()
			return
		}

		for {
			select {
			case msg, ok := <-wsConn.Messages:
				if !ok {
					handleWsClosure(rds, pubSubConn, userId, room.Id)
					return
				}

				log.Printf("User \"%s\" has sent ws message with event \"%s\": %v\n", userId, msg.Event, string(msg.Payload))
				handleWsMessage(mdb, rds, wsConn, pubSubConn, pack, room.Id, msg)

			case rdsMsg, ok := <-pubSubConn.Messages:
				if !ok {
					wsConn.Conn.Close()
					return
				}
				var msg ws.InternalMessage
				json.Unmarshal([]byte(rdsMsg.Payload), &msg)

				log.Printf("User \"%s\" has recieved pubSub message from %s with event \"%s\": %v\n", userId, msg.From.Id, msg.Event, string(msg.Payload))
				handleRdsMessage(mdb, rds, wsConn, pubSubConn, pack, room, userId, msg)
			}
		}
	}
}
