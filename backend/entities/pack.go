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

const PACKS_COLLECTION = "packs"

type Pack struct {
	Id             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Author         User               `json:"author"`
	RoundsCheckSum []byte             `json:"-" bson:"roundsCheckSum"`
	Content        string             `json:"-" bson:"content"`
	PackDTO        `bson:"inline"`
}

type PackDTO struct {
	Name       string      `json:"name" binding:"max=50"`
	Type       PrivacyType `json:"type" binding:"oneof=public private"`
	Rounds     []Round     `json:"rounds" binding:"max=10,unique=Name"`
	FinalRound FinalRound  `json:"finalRound"`
}

type PackPreview struct {
	Id   primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name string             `json:"name"`
}

type Round struct {
	Name       string     `json:"name" binding:"min=1,max=50"`
	Categories []Category `json:"categories" binding:"min=1,max=10,unique=Name"`
}

type Category struct {
	Name      string     `json:"name" binding:"min=1,max=25"`
	Questions []Question `json:"questions" binding:"min=1,max=10"`
}

type MediaType string

const (
	Image MediaType = "image"
	Audio MediaType = "audio"
	Video MediaType = "video"
)

type Attachment struct {
	MediaType  MediaType `json:"mediaType" binding:"oneof=image audio video"`
	ContentUrl string    `json:"contentUrl" binding:"url,max=2000"`
}

type Question struct {
	HiddenQuestion
	Answers []string `json:"answers" binding:"min=1,max=10,dive,min=1,max=50"`
	Comment *string  `json:"comment" binding:"omitnil,max=200"`
}

type FinalRound struct {
	Categories []FinalCategory `json:"categories" binding:"min=1,max=10"`
}

type FinalCategory struct {
	HiddenFinalCategory
	Question FinalQuestion `json:"question"`
}

type FinalQuestion struct {
	HiddenFinalQuestion
	Answers []string `json:"answers" binding:"min=1,max=10,dive,min=1,max=50"`
	Comment *string  `json:"comment" binding:"omitnil,min=10,max=100"`
}

type HiddenPack struct {
	Id         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Author     User               `json:"author"`
	Name       string             `json:"name" binding:"max=50"`
	Rounds     []hiddenRound      `json:"rounds" binding:"max=10,unique=Name"`
	FinalRound hiddenFinalRound   `json:"finalRound"`
}

type hiddenRound struct {
	Name       string           `json:"name" binding:"max=50"`
	Categories []hiddenCategory `json:"categories" binding:"max=10,unique=Name"`
}

type hiddenCategory struct {
	Name string `json:"name" binding:"max=25"`
}

type HiddenQuestion struct {
	Index      int         `json:"index" binding:"min=0,max=9"`
	Value      int         `json:"value" binding:"max=10000"`
	Text       string      `json:"text" binding:"required,max=200"`
	Attachment *Attachment `json:"attachment" binding:"omitnil"`
}

type hiddenFinalRound struct {
	Categories []HiddenFinalCategory `json:"categories"`
}

type HiddenFinalCategory struct {
	Name string `json:"name" binding:"max=25"`
}

type HiddenFinalQuestion struct {
	Text       string      `json:"text" binding:"required,max=200"`
	Attachment *Attachment `json:"attachment" binding:"omitnil"`
}

func NewHiddenPack(pack Pack) HiddenPack {
	hiddenRounds := make([]hiddenRound, len(pack.Rounds))
	for i, round := range pack.Rounds {
		hiddenCategories := make([]hiddenCategory, len(round.Categories))
		for j, category := range round.Categories {
			hiddenCategories[j] = hiddenCategory{Name: category.Name}
		}
		hiddenRounds[i] = hiddenRound{Name: round.Name, Categories: hiddenCategories}
	}
	hiddenFinalCategories := make([]HiddenFinalCategory, len(pack.FinalRound.Categories))
	for i, finalCategory := range pack.FinalRound.Categories {
		hiddenFinalCategories[i] = HiddenFinalCategory{Name: finalCategory.Name}
	}
	return HiddenPack{
		Id:     pack.Id,
		Author: pack.Author,
		Name:   pack.Name,
		Rounds: hiddenRounds,
		FinalRound: hiddenFinalRound{
			Categories: hiddenFinalCategories,
		},
	}
}

func GetPack(mdb *mongo.Database, id primitive.ObjectID) (*Pack, custErrors.HttpError) {
	var pack Pack

	err := mdb.Collection(PACKS_COLLECTION).FindOne(
		context.TODO(),
		bson.D{{Key: "_id", Value: id}},
	).Decode(&pack)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, custErrors.NewHttpError(
				http.StatusNotFound,
				gin.H{"error": fmt.Sprintf("there is no pack with id \"%s\"", id)},
			)
		}
		return nil, custErrors.NewInternalError(err)
	}

	return &pack, nil
}
