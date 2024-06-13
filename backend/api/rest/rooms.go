package rest

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/holdennekt/sgame/custErrors"
	"github.com/holdennekt/sgame/entities"
	"github.com/redis/go-redis/v9"
)

func GetRoomsHandler(rds *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		lobbyRooms := make([]entities.LobbyRoom, 0)

		pattern := fmt.Sprintf("%s*", entities.ROOM_PREFIX)
		iter := rds.Scan(context.TODO(), uint64(rds.Options().DB), pattern, 0).Iterator()
		for iter.Next(context.TODO()) {
			room, httpErr := entities.GetRoomByKey(rds, iter.Val())
			if httpErr != nil {
				custErrors.AbortWithError(c, httpErr)
				return
			}

			lobbyRooms = append(lobbyRooms, entities.NewLobbyRoom(room))
		}

		c.JSON(http.StatusOK, lobbyRooms)
	}
}
