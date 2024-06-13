package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/holdennekt/sgame/api"
	"github.com/holdennekt/sgame/api/rest"
	wsLobby "github.com/holdennekt/sgame/api/ws/lobby"
	wsRoom "github.com/holdennekt/sgame/api/ws/room"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	godotenv.Load()

	rds := redis.NewClient(&redis.Options{
		Addr:     getEnvVar("REDIS_HOST") + ":" + getEnvVar("REDIS_PORT"),
		Username: getEnvVar("REDIS_USERNAME"),
		Password: getEnvVar("REDIS_PASSWORD"),
		DB:       getEnvVarInt("REDIS_DB"),
	})
	defer rds.Close()

	opts := options.Client().ApplyURI(getEnvVar("MONGO_CONN"))
	conn, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Disconnect(context.TODO())

	mdb := conn.Database(getEnvVar("MONGO_DB_NAME"))
	InitDB(context.TODO(), mdb)

	engine := gin.Default()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{getEnvVar("MY_CLIENT_ORIGIN")}
	corsConfig.AllowCredentials = true
	engine.Use(cors.New(corsConfig))

	engine.Handle(http.MethodPost, "/login", api.LoginHandler(mdb, rds))
	engine.Handle(http.MethodPost, "/register", api.RegisterHandler(mdb, rds))
	engine.Handle(http.MethodGet, "/user", api.AuthorizeConnection(rds), api.GetUser(mdb))

	restGroup := engine.Group("/rest", api.AuthorizeConnection(rds))

	restGroup.Handle(http.MethodPost, "/room", rest.CreateRoomHandler(mdb, rds))
	restGroup.Handle(http.MethodGet, "/rooms", rest.GetRoomsHandler(rds))
	restGroup.Handle(http.MethodGet, "/room/:id", rest.GetRoomHandler(rds))
	restGroup.Handle(http.MethodPatch, "/room/:id", rest.EnterRoomHandler(mdb, rds))

	restGroup.Handle(http.MethodPost, "/pack", rest.CreatePackHandler(mdb))
	restGroup.Handle(http.MethodGet, "/packsPreview", rest.GetPacksPreviewHandler(mdb))
	restGroup.Handle(http.MethodGet, "/packs", rest.GetHiddenPacksHandler(mdb))
	restGroup.Handle(http.MethodGet, "/pack/:id", rest.GetPackHandler(mdb))
	restGroup.Handle(http.MethodPut, "/pack/:id", rest.UpdatePackHandler(mdb))
	restGroup.Handle(http.MethodDelete, "/pack/:id", rest.DeletePackHandler(mdb))

	wsGroup := engine.Group("/ws", api.AuthorizeConnection(rds))
	wsGroup.Handle(http.MethodGet, "/lobby", wsLobby.ConnectHandler(mdb, rds))
	wsGroup.Handle(http.MethodGet, "/room/:id", wsRoom.ConnectHandler(mdb, rds))

	servAddres := getEnvVar("HOST") + ":" + getEnvVar("PORT")
	log.Fatal(engine.Run(servAddres))
}
