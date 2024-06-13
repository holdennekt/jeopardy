package rest

import (
	"context"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/holdennekt/sgame/api"
	"github.com/holdennekt/sgame/api/ws"
	wsLobby "github.com/holdennekt/sgame/api/ws/lobby"
	lobbyEvents "github.com/holdennekt/sgame/api/ws/lobby/events"
	roomEvents "github.com/holdennekt/sgame/api/ws/room/events"
	"github.com/holdennekt/sgame/custErrors"
	"github.com/holdennekt/sgame/entities"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const PASSWORD_QUERY_PARAM = "password"

func CreateRoomHandler(mdb *mongo.Database, rds *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.MustGet(api.USER_ID_CONTEXT_KEY).(primitive.ObjectID)

		user, httpErr := entities.GetUser(mdb, userId)
		if httpErr != nil {
			custErrors.AbortWithError(c, httpErr)
			return
		}

		var roomDTO entities.RoomDTO
		if err := c.ShouldBindJSON(&roomDTO); err != nil {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				gin.H{"error": strings.Join(custErrors.ParseValidationErrors(err), ", ")},
			)
			return
		}

		pack, httpErr := entities.GetPack(mdb, roomDTO.PackId)
		if httpErr != nil {
			custErrors.AbortWithError(c, httpErr)
			return
		}

		room := &entities.Room{
			Id:      primitive.NewObjectID(),
			RoomDTO: roomDTO,
			PackPreview: entities.PackPreview{
				Id:   pack.Id,
				Name: pack.Name,
			},
			Players:   make([]entities.Player, 0),
			CreatedBy: userId,
			Host:      &entities.Host{User: *user},
		}

		key := entities.GetRoomRedisKey(room.Id.Hex())
		if err := rds.JSONSet(context.TODO(), key, "$", room).Err(); err != nil {
			custErrors.AbortWithInternalError(c, err)
			return
		}

		roomMessage := lobbyEvents.NewLobbyRoomInternalMessage(room)
		if err := ws.PublishRdsMessage(rds, wsLobby.LOBBY, roomMessage); err != nil {
			log.Println(err)
		}

		c.JSON(http.StatusCreated, gin.H{"id": room.Id})
	}
}

func GetRoomHandler(rds *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.MustGet(api.USER_ID_CONTEXT_KEY).(primitive.ObjectID)

		roomId := entities.GetRoomRedisKey(c.Param("id"))
		room, httpErr := entities.GetRoomByKey(rds, roomId)
		if httpErr != nil {
			custErrors.AbortWithError(c, httpErr)
			return
		}

		password := c.Query(PASSWORD_QUERY_PARAM)
		if room.Options.Type == entities.Private && password != *room.Options.Password {
			custErrors.AbortWithError(c, custErrors.NewHttpError(
				http.StatusForbidden,
				gin.H{"error": "wrong password"},
			))
			return
		}

		c.JSON(http.StatusOK, room.GetProjection(userId))
	}
}

func EnterRoomHandler(mdb *mongo.Database, rds *redis.Client) gin.HandlerFunc {
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

		if room.IsUserIn(userId) {
			c.JSON(http.StatusOK, room.GetProjection(userId))
			return
		}

		if room.CurrentRound == nil && !room.FinalRoundState.IsActive {
			custErrors.AbortWithError(c, custErrors.NewHttpError(
				http.StatusForbidden,
				gin.H{"error": "game already started"},
			))
			return
		}

		password := c.Query(PASSWORD_QUERY_PARAM)
		if room.Options.Type == entities.Private && password != *room.Options.Password {
			custErrors.AbortWithError(c, custErrors.NewHttpError(
				http.StatusForbidden,
				gin.H{"error": "wrong password"},
			))
			return
		}

		err := api.TryUpdateRoom(rds, room.Id, func(tx *redis.Tx) error {
			room, httpErr := entities.GetRoomByKey(rds, roomKey)
			if httpErr != nil {
				return httpErr
			}

			isFull := len(room.Players) >= room.Options.MaxPlayers
			canBeHost := user.Id == room.CreatedBy && room.Host == nil

			if isFull && !canBeHost {
				return custErrors.NewHttpError(
					http.StatusConflict,
					gin.H{"error": "the room is already full"},
				)
			}

			_, err := tx.TxPipelined(context.TODO(), func(p redis.Pipeliner) error {
				if canBeHost {
					room.Host = &entities.Host{User: *user}
					p.JSONSet(context.TODO(), roomKey, "$.host", room.Host)
				} else {
					room.Players = append(room.Players, entities.Player{User: *user})
					p.JSONSet(context.TODO(), roomKey, "$.players", room.Players)
				}
				connectedPlayerIndex := slices.IndexFunc(room.Players, func(p entities.Player) bool {
					return p.IsConnected
				})
				isAnyConnectedUser := room.Host.IsConnected || connectedPlayerIndex != -1
				if !isAnyConnectedUser {
					p.Expire(context.TODO(), roomKey, 5*time.Minute)
				}
				return nil
			})

			return err
		}, 3)
		if err != nil {
			switch val := err.(type) {
			case custErrors.HttpError:
				custErrors.AbortWithError(c, val)
				return
			default:
				custErrors.AbortWithInternalError(c, val)
				return
			}
		}

		roomMessage := roomEvents.RoomInternalMessage()
		if err := ws.PublishRdsMessage(rds, roomKey, roomMessage); err != nil {
			log.Println(err)
		}

		lobbyMessage := lobbyEvents.NewLobbyRoomInternalMessage(room)
		if err := ws.PublishRdsMessage(rds, wsLobby.LOBBY, lobbyMessage); err != nil {
			log.Println(err)
		}

		c.JSON(http.StatusOK, room.GetProjection(userId))
	}
}
