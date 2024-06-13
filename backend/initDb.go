package main

import (
	"context"
	"time"

	"github.com/holdennekt/sgame/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const CODE_NAMESPACE_EXISTS = 48

func handleError(err error) {
	if err != nil {
		mongoErr, ok := err.(mongo.CommandError)
		if !ok {
			panic(err)
		}
		if mongoErr.Code != CODE_NAMESPACE_EXISTS {
			panic(err)
		}
	}
}

func InitDB(parent context.Context, mdb *mongo.Database) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	err := mdb.CreateCollection(ctx, entities.USERS_COLLECTION)
	if err != nil {
		handleError(err)
	}
	_, err = mdb.Collection(entities.USERS_COLLECTION).Indexes().CreateOne(
		context.TODO(),
		mongo.IndexModel{
			Keys:    bson.D{{Key: "login", Value: 1}},
			Options: options.Index().SetName("login_unique").SetUnique(true),
		},
	)
	if err != nil {
		handleError(err)
	}

	err = mdb.CreateCollection(ctx, entities.PACKS_COLLECTION)
	if err != nil {
		handleError(err)
	}

	_, err = mdb.Collection(entities.PACKS_COLLECTION).Indexes().CreateOne(
		context.TODO(),
		mongo.IndexModel{
			Keys:    bson.D{{Key: "content", Value: "text"}},
			Options: options.Index().SetName("content_text"),
		},
	)
	if err != nil {
		handleError(err)
	}
}
