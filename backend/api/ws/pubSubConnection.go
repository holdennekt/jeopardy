package ws

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PubSubConn struct {
	userId   primitive.ObjectID
	Conn     *redis.PubSub
	Messages <-chan *redis.Message
	Publish  func(message InternalMessage) error
}

func ConnectUserToPubSub(rds *redis.Client, userId primitive.ObjectID, chanName string) *PubSubConn {
	pubSub := rds.Subscribe(context.TODO(), chanName)
	return &PubSubConn{
		userId:   userId,
		Conn:     pubSub,
		Messages: pubSub.Channel(),
		Publish: func(message InternalMessage) error {
			return PublishRdsMessage(rds, chanName, message)
		},
	}
}

func PublishRdsMessage(rds *redis.Client, chanName string, message InternalMessage) error {
	marshaled, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return rds.Publish(context.TODO(), chanName, marshaled).Err()
}
