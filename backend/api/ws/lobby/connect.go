package wsLobby

import (
	"encoding/json"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/holdennekt/sgame/api"
	"github.com/holdennekt/sgame/api/ws"
	"github.com/holdennekt/sgame/custErrors"
	"github.com/holdennekt/sgame/entities"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const LOBBY = "lobby"

func ConnectHandler(mdb *mongo.Database, rds *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.MustGet(api.USER_ID_CONTEXT_KEY).(primitive.ObjectID)

		user, httpErr := entities.GetUser(mdb, userId)
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

		pubSubConn := ws.ConnectUserToPubSub(rds, userId, LOBBY)

		for {
			select {
			case msg, ok := <-wsConn.Messages:
				if !ok {
					pubSubConn.Conn.Close()
					return
				}

				log.Printf("User \"%s\" has sent ws message with event \"%s\": %v\n", userId, msg.Event, string(msg.Payload))
				handleWsMessage(pubSubConn, msg)

			case rdsMsg, ok := <-pubSubConn.Messages:
				if !ok {
					wsConn.Conn.Close()
					return
				}
				var msg ws.InternalMessage
				json.Unmarshal([]byte(rdsMsg.Payload), &msg)

				log.Printf("User \"%s\" has recieved pubSub message from %s with event \"%s\": %v\n", userId, msg.From.Id, msg.Event, string(msg.Payload))
				handleRdsMessage(wsConn, msg)
			}
		}
	}
}
