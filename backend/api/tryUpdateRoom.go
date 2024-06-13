package api

import (
	"context"

	"github.com/holdennekt/sgame/custErrors"
	"github.com/holdennekt/sgame/entities"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TryUpdateRoom(rds *redis.Client, roomId primitive.ObjectID, updateFunc func(tx *redis.Tx) error, retries int) error {
	for i := 0; i < retries; i++ {
		err := rds.Watch(context.TODO(), updateFunc, entities.GetRoomRedisKey(roomId.Hex()))

		if err != redis.Nil {
			return err
		}
	}
	return custErrors.ErrInternal
}
