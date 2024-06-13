package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/holdennekt/sgame/custErrors"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const SESSION_ID_COOKIE_NAME = "sessionId"
const SESSIONS_KEY = "sessions"
const USER_ID_CONTEXT_KEY = "userId"

func AuthorizeConnection(rds *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionId, err := c.Cookie(SESSION_ID_COOKIE_NAME)
		if err != nil {
			custErrors.AbortWithError(c, custErrors.NewHttpError(
				http.StatusUnauthorized,
				gin.H{"error": "missing sessionId cookie"},
			))
			return
		}

		userId, err := rds.HGet(context.TODO(), SESSIONS_KEY, sessionId).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				custErrors.AbortWithError(c, custErrors.NewHttpError(
					http.StatusUnauthorized,
					gin.H{"error": "invalid sessionId"},
				))
				return
			}
			custErrors.AbortWithInternalError(c, err)
			return
		}
		userObjectId, err := primitive.ObjectIDFromHex(userId)
		if err != nil {
			custErrors.AbortWithInternalError(c, err)
			return
		}

		c.Set(USER_ID_CONTEXT_KEY, userObjectId)
	}
}
