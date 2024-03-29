package main

import (
	"context"
	"github.com/go-redis/redis"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/service/grpc"
	"github.com/micro/go-micro/util/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	UserService "github.com/nayanmakasare/UserService/proto"
	"time"
)

const (
	//defaultHost = "mongodb://nayan:tlwn722n@cluster0-shard-00-00-8aov2.mongodb.net:27017,cluster0-shard-00-01-8aov2.mongodb.net:27017,cluster0-shard-00-02-8aov2.mongodb.net:27017/test?ssl=true&replicaSet=Cluster0-shard-0&authSource=admin&retryWrites=true&w=majority"
	//defaultHost = "mongodb://192.168.1.9:27017"
	defaultHost = "mongodb://192.168.1.143:27017"
)

func main(){
	service := grpc.NewService(
		micro.Name("UserService"),
		micro.Address(":50052"),
		micro.Version("1.0"),
	)
	service.Init()
	uri := defaultHost
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Debug(err)
	}
	client := GetRedisClient()
	h := UserServiceHandler{MongoCollection:mongoClient.Database("cloudwalker").Collection("users"), RedisConnection:client}
	err = UserService.RegisterUserServiceHandler(service.Server(), &h)
	if err != nil {
		log.Fatal(err)
	}
	// Run service
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}

func GetRedisClient() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	_, err := client.Ping().Result()
	if err != nil {
		log.Fatalf("Could not connect to redis %v", err)
	}
	return client
}
