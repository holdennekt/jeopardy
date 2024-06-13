package rest

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/holdennekt/sgame/api"
	"github.com/holdennekt/sgame/custErrors"
	"github.com/holdennekt/sgame/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const FILTER_QUERY_PARAM = "filter"
const PAGE_QUERY_PARAM = "page"
const LIMIT_QUERY_PARAM = "limit"

const DEFAULT_PAGE = "1"
const DEFAULT_LIMIT = "50"

func GetPacksPreviewHandler(mdb *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.MustGet(api.USER_ID_CONTEXT_KEY).(primitive.ObjectID)

		searchFilter := strings.TrimSpace(c.Query(FILTER_QUERY_PARAM))
		limit, err := strconv.ParseInt(c.DefaultQuery(LIMIT_QUERY_PARAM, DEFAULT_LIMIT), 10, 64)
		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				gin.H{"error": "limit must be a number"},
			)
			return
		}

		packs := make([]entities.PackPreview, 0)
		res, err := mdb.Collection(entities.PACKS_COLLECTION).Find(
			context.TODO(),
			bson.M{
				"name": primitive.Regex{
					Pattern: searchFilter,
					Options: "i",
				},
				"$or": []bson.M{
					{"type": "public"},
					{"author": userId},
				},
			},
			options.Find().SetLimit(limit).SetProjection(bson.M{"_id": 1, "name": 1}),
		)
		if err != nil {
			custErrors.AbortWithInternalError(c, err)
			return
		}
		if err := res.All(context.TODO(), &packs); err != nil {
			custErrors.AbortWithInternalError(c, err)
			return
		}

		c.JSON(http.StatusOK, packs)
	}
}

func GetHiddenPacksHandler(mdb *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.MustGet(api.USER_ID_CONTEXT_KEY).(primitive.ObjectID)

		searchFilter := strings.TrimSpace(c.Query(FILTER_QUERY_PARAM))
		page, err := strconv.ParseInt(c.DefaultQuery(PAGE_QUERY_PARAM, DEFAULT_PAGE), 10, 64)
		if err != nil || page < 1 {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				gin.H{"error": "page must be an integer number greater than 0"},
			)
			return
		}
		limit, err := strconv.ParseInt(c.DefaultQuery(LIMIT_QUERY_PARAM, DEFAULT_LIMIT), 10, 64)
		if err != nil || limit < 1 {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				gin.H{"error": "limit must be an integer number greater than 0"},
			)
			return
		}

		packs := make([]entities.Pack, 0)
		res, err := mdb.Collection(entities.PACKS_COLLECTION).Find(
			context.TODO(),
			bson.M{
				"name": primitive.Regex{
					Pattern: searchFilter,
					Options: "i",
				},
				"$or": []bson.M{
					{"type": "public"},
					{"author": userId},
				},
			},
			options.
				Find().
				SetSort(bson.D{{Key: "_id", Value: 1}}).
				SetSkip((page-1)*limit).
				SetLimit(limit),
		)
		if err != nil {
			custErrors.AbortWithInternalError(c, err)
			return
		}
		if err := res.All(context.TODO(), &packs); err != nil {
			custErrors.AbortWithInternalError(c, err)
			return
		}

		hiddenPacks := make([]entities.HiddenPack, len(packs))
		for i, pack := range packs {
			hiddenPacks[i] = entities.NewHiddenPack(pack)
		}

		count, err := mdb.Collection(entities.PACKS_COLLECTION).CountDocuments(
			context.TODO(),
			bson.M{
				"name": primitive.Regex{
					Pattern: searchFilter,
					Options: "i",
				},
				"$or": []bson.M{
					{"type": "public"},
					{"author": userId},
				},
			},
		)
		if err != nil {
			custErrors.AbortWithInternalError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"count":      count,
			"portionRes": hiddenPacks,
		})
	}
}
