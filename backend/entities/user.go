package entities

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/holdennekt/sgame/custErrors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const USERS_COLLECTION = "users"

var SYSTEM User = User{Id: primitive.NilObjectID}

type User struct {
	Id     primitive.ObjectID `json:"id" bson:"_id,omitempty" binding:"required"`
	Name   string             `json:"name" bson:"name" binding:"min=1,max=20"`
	Avatar *string            `json:"avatar" bson:"avatar" binding:"omitnil,url"`
}

type DbUserDTO struct {
	Login    string `json:"login" form:"login" bson:"login" binding:"min=4,max=20"`
	Password string `json:"password" form:"password" bson:"password" binding:"min=8,max=40"`
}

type DbUser struct {
	User      `bson:"inline"`
	DbUserDTO `bson:"inline"`
}

type Host struct {
	User
	IsConnected bool `json:"isConnected"`
}

type Player struct {
	User
	Score       int  `json:"score"`
	IsConnected bool `json:"isConnected"`
}

func GetDbUser(mdb *mongo.Database, userId primitive.ObjectID) (*DbUser, custErrors.HttpError) {
	var dbUser DbUser
	err := mdb.Collection(USERS_COLLECTION).FindOne(
		context.TODO(),
		bson.D{{Key: "_id", Value: userId}},
	).Decode(&dbUser)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, custErrors.NewHttpError(
				http.StatusNotFound,
				gin.H{"error": fmt.Sprintf("there is no user with id \"%s\"", userId)},
			)
		}
		return nil, custErrors.NewInternalError(err)
	}

	return &dbUser, nil
}

func GetDbUserByLogin(mdb *mongo.Database, login string) (*DbUser, custErrors.HttpError) {
	var dbUser DbUser
	err := mdb.Collection(USERS_COLLECTION).FindOne(
		context.TODO(),
		bson.D{{Key: "login", Value: login}},
	).Decode(&dbUser)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, custErrors.NewHttpError(
				http.StatusNotFound,
				gin.H{"error": fmt.Sprintf("there is no user with login \"%s\"", login)},
			)
		}
		return nil, custErrors.NewInternalError(err)
	}

	return &dbUser, nil
}

func GetUser(mdb *mongo.Database, userId primitive.ObjectID) (*User, custErrors.HttpError) {
	var user User
	err := mdb.Collection(USERS_COLLECTION).FindOne(
		context.TODO(),
		bson.D{{Key: "_id", Value: userId}},
	).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, custErrors.NewHttpError(
				http.StatusNotFound,
				gin.H{"error": fmt.Sprintf("there is no user with id \"%s\"", userId)},
			)
		}
		return nil, custErrors.NewInternalError(err)
	}

	return &user, nil
}
