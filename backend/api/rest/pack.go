package rest

import (
	"context"
	"crypto/sha1"
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/holdennekt/sgame/api"
	"github.com/holdennekt/sgame/custErrors"
	"github.com/holdennekt/sgame/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func validateRounds(packDTO entities.PackDTO) ([]string, custErrors.HttpError) {
	content := []string{packDTO.Name}

	for _, round := range packDTO.Rounds {
		questionsCount := len(round.Categories[0].Questions)
		for _, category := range round.Categories {
			if len(category.Questions) != questionsCount {
				return nil, custErrors.NewHttpError(
					http.StatusBadRequest,
					gin.H{"error": "within the round every category must have equal number of questions"},
				)
			}
			content = append(content, category.Name)
		}
		content = append(content, round.Name)
	}

	return content, nil
}

func validateRoundsCheckSum(mdb *mongo.Database, packDTO entities.PackDTO, ignoreId primitive.ObjectID) ([]byte, custErrors.HttpError) {
	marshaledRounds, _ := json.Marshal(struct {
		Rounds     []entities.Round    `json:"rounds"`
		FinalRound entities.FinalRound `json:"finalRound"`
	}{Rounds: packDTO.Rounds, FinalRound: packDTO.FinalRound})
	hasher := sha1.New()
	_, err := hasher.Write(marshaledRounds)
	if err != nil {
		return nil, custErrors.NewInternalError(err)
	}
	roundsCheckSum, err := hasher.(encoding.BinaryMarshaler).MarshalBinary()
	if err != nil {
		return nil, custErrors.NewInternalError(err)
	}
	var packWithSameRounds entities.Pack
	err = mdb.Collection(entities.PACKS_COLLECTION).FindOne(
		context.TODO(),
		bson.D{
			{Key: "roundsCheckSum", Value: roundsCheckSum},
			{Key: "_id", Value: bson.D{{Key: "$ne", Value: ignoreId}}},
		},
	).Decode(&packWithSameRounds)
	if err == nil {
		return nil, custErrors.NewHttpError(
			http.StatusConflict,
			gin.H{"error": fmt.Sprintf("the pack with such rounds already exists and has id \"%s\"", packWithSameRounds.Id.Hex())},
		)
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, custErrors.NewInternalError(err)
	}

	return roundsCheckSum, nil
}

func CreatePackHandler(mdb *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.MustGet(api.USER_ID_CONTEXT_KEY).(primitive.ObjectID)

		user, httpErr := entities.GetUser(mdb, userId)
		if httpErr != nil {
			custErrors.AbortWithError(c, httpErr)
			return
		}

		var packDTO entities.PackDTO
		if err := c.ShouldBindJSON(&packDTO); err != nil {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				gin.H{"error": strings.Join(custErrors.ParseValidationErrors(err), ", ")},
			)
			return
		}

		content, httpErr := validateRounds(packDTO)
		if httpErr != nil {
			custErrors.AbortWithError(c, httpErr)
			return
		}

		roundsCheckSum, httpErr := validateRoundsCheckSum(mdb, packDTO, primitive.NilObjectID)
		if httpErr != nil {
			custErrors.AbortWithError(c, httpErr)
			return
		}

		pack := &entities.Pack{
			Author:         *user,
			RoundsCheckSum: roundsCheckSum,
			Content:        strings.Join(content, ", "),
			PackDTO:        packDTO,
		}

		res, err := mdb.Collection(entities.PACKS_COLLECTION).InsertOne(context.TODO(), pack)
		if err != nil {
			custErrors.AbortWithInternalError(c, err)
			return
		}

		c.JSON(http.StatusCreated, gin.H{"id": res.InsertedID})
	}
}

func GetPackHandler(mdb *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.MustGet(api.USER_ID_CONTEXT_KEY).(primitive.ObjectID)

		packId := c.Param("id")
		objId, err := primitive.ObjectIDFromHex(packId)
		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				gin.H{"error": "invalid packId"},
			)
			return
		}

		pack, httpErr := entities.GetPack(mdb, objId)
		if httpErr != nil {
			custErrors.AbortWithError(c, httpErr)
			return
		}

		if userId != pack.Author.Id {
			c.AbortWithStatusJSON(
				http.StatusForbidden,
				gin.H{"error": "can not get not yours pack"},
			)
			return
		}

		c.JSON(http.StatusOK, pack)
	}
}

func UpdatePackHandler(mdb *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.MustGet(api.USER_ID_CONTEXT_KEY).(primitive.ObjectID)

		packId := c.Param("id")
		objId, err := primitive.ObjectIDFromHex(packId)
		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				gin.H{"error": "invalid packId"},
			)
			return
		}

		pack, httpErr := entities.GetPack(mdb, objId)
		if httpErr != nil {
			custErrors.AbortWithError(c, httpErr)
			return
		}

		if pack.Author.Id != userId {
			c.AbortWithStatusJSON(
				http.StatusForbidden,
				gin.H{"error": "cannot update not your pack"},
			)
			return
		}

		var packDTO entities.PackDTO
		if err := c.ShouldBindJSON(&packDTO); err != nil {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				gin.H{"error": strings.Join(custErrors.ParseValidationErrors(err), ", ")},
			)
			return
		}

		content, httpErr := validateRounds(packDTO)
		if httpErr != nil {
			custErrors.AbortWithError(c, httpErr)
			return
		}

		newRoundsCheckSum, httpErr := validateRoundsCheckSum(mdb, packDTO, objId)
		if httpErr != nil {
			custErrors.AbortWithError(c, httpErr)
			return
		}

		pack.RoundsCheckSum = newRoundsCheckSum
		pack.Content = strings.Join(content, ", ")
		pack.PackDTO = packDTO

		res, err := mdb.Collection(entities.PACKS_COLLECTION).ReplaceOne(
			context.TODO(),
			bson.D{{Key: "_id", Value: pack.Id}},
			pack,
		)
		if err != nil {
			custErrors.AbortWithInternalError(c, err)
			return
		}
		if res.MatchedCount == 0 {
			c.AbortWithStatusJSON(
				http.StatusNotFound,
				gin.H{"error": "provided pack was not found"},
			)
			return
		}

		c.JSON(http.StatusOK, pack)
	}
}

func DeletePackHandler(mdb *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.MustGet(api.USER_ID_CONTEXT_KEY).(primitive.ObjectID)

		packId := c.Param("id")
		objId, err := primitive.ObjectIDFromHex(packId)
		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				gin.H{"error": "invalid packId"},
			)
			return
		}

		pack, httpErr := entities.GetPack(mdb, objId)
		if httpErr != nil {
			custErrors.AbortWithError(c, httpErr)
			return
		}

		if pack.Author.Id != userId {
			c.AbortWithStatusJSON(
				http.StatusForbidden,
				gin.H{"error": "cannot delete not your pack"},
			)
			return
		}

		_, err = mdb.Collection(entities.PACKS_COLLECTION).DeleteOne(
			context.TODO(),
			bson.D{{Key: "_id", Value: objId}},
		)
		if err != nil {
			custErrors.AbortWithInternalError(c, err)
			return
		}

		c.JSON(http.StatusNoContent, packId)
	}
}
