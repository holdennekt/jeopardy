package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/holdennekt/sgame/custErrors"
	"github.com/holdennekt/sgame/entities"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func LoginHandler(mdb *mongo.Database, rds *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {

		var dbUserDTO entities.DbUserDTO
		if err := c.ShouldBind(&dbUserDTO); err != nil {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				gin.H{"error": strings.Join(custErrors.ParseValidationErrors(err), ", ")},
			)
			return
		}

		dbUser, httpErr := entities.GetDbUserByLogin(mdb, dbUserDTO.Login)
		if httpErr != nil {
			custErrors.AbortWithError(c, httpErr)
			return
		}

		userIdStr := dbUser.Id.Hex()

		err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(dbUserDTO.Password))
		if err != nil {
			if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
				c.AbortWithStatusJSON(
					http.StatusUnauthorized,
					gin.H{"error": "wrong password"},
				)
				return
			}
			custErrors.AbortWithInternalError(c, err)
			return
		}

		sessions, err := rds.HGetAll(context.TODO(), SESSIONS_KEY).Result()
		if err != nil {
			custErrors.AbortWithInternalError(c, err)
			return
		}
		for sessionId, userId := range sessions {
			if userIdStr == userId {
				c.SetCookie(SESSION_ID_COOKIE_NAME, sessionId, 0, "", "", false, true)
				c.JSON(http.StatusOK, gin.H{"id": userIdStr})
				return
			}
		}

		sessionId, err := SetSession(rds, userIdStr)
		if err != nil {
			custErrors.AbortWithInternalError(c, err)
		}
		c.SetCookie(SESSION_ID_COOKIE_NAME, sessionId, 0, "", "", false, true)
		c.JSON(http.StatusOK, gin.H{"id": userIdStr})
	}
}

func RegisterHandler(mdb *mongo.Database, rds *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var dbUserDTO entities.DbUserDTO
		if err := c.ShouldBind(&dbUserDTO); err != nil {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				gin.H{"error": strings.Join(custErrors.ParseValidationErrors(err), ", ")},
			)
			return
		}

		hashed, err := bcrypt.GenerateFromPassword([]byte(dbUserDTO.Password), bcrypt.DefaultCost)
		if err != nil {
			custErrors.AbortWithInternalError(c, err)
			return
		}
		dbUserDTO.Password = string(hashed)

		user := entities.User{
			Name: dbUserDTO.Login,
		}
		dbUser := &entities.DbUser{
			User:      user,
			DbUserDTO: dbUserDTO,
		}

		res, err := mdb.Collection(entities.USERS_COLLECTION).InsertOne(context.TODO(), dbUser)
		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				c.AbortWithStatusJSON(
					http.StatusConflict,
					gin.H{"error": "such login already exists"},
				)
				return
			}
			custErrors.AbortWithInternalError(c, err)
			return
		}
		userIdStr := res.InsertedID.(primitive.ObjectID).Hex()

		sessionId, err := SetSession(rds, userIdStr)
		if err != nil {
			custErrors.AbortWithInternalError(c, err)
		}
		c.SetCookie(SESSION_ID_COOKIE_NAME, sessionId, 0, "", "", false, true)
		c.JSON(http.StatusCreated, gin.H{"id": userIdStr})
	}
}

func SetSession(rds *redis.Client, userId string) (string, error) {
	sessionId, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}

	err = rds.HSet(context.TODO(), SESSIONS_KEY, sessionId.String(), userId).Err()
	if err != nil {
		return "", err
	}
	return sessionId.String(), nil
}

func GetUser(mdb *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.MustGet(USER_ID_CONTEXT_KEY).(primitive.ObjectID)

		user, httpErr := entities.GetUser(mdb, userId)
		if httpErr != nil {
			custErrors.AbortWithError(c, httpErr)
			return
		}

		c.JSON(http.StatusOK, user)
	}
}
