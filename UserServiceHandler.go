package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/micro/protobuf/ptypes"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	UserService "github.com/nayanmakasare/UserService/proto"
	"time"
)

type UserServiceHandler struct {
	MongoCollection *mongo.Collection
	RedisConnection *redis.Client
}

func (h *UserServiceHandler) CreateUser(ctx context.Context, req *UserService.User, res *UserService.CreateReponse) error {
	findResult := h.MongoCollection.FindOne(context.Background(), bson.D{{"googleid", req.GoogleId}})
	if findResult.Err() != nil {
		// means user is not present so add user in DB
		if h.ValidateUserInformation(req) {
			ts, _ := ptypes.TimestampProto(time.Now())
			req.CreatedAt = ts
			_, err := h.MongoCollection.InsertOne(context.Background(), req)
			res.IsCreated = true
			go h.PreparingDataForUser(req)
			return err
		}else {
			return errors.New("User Information is not valid")
		}
	}else {
		return errors.New("User Already Present")
	}
}

func(h *UserServiceHandler) PreparingDataForUser(user *UserService.User) {
	genreKey :=  fmt.Sprintf("user:%s:genre", user.GoogleId)
	for _,v := range user.Genre {
		h.RedisConnection.SAdd(genreKey, v)
	}
	languageKey := fmt.Sprintf("user:%s:languages", user.GoogleId)
	for _,v := range user.Language {
		h.RedisConnection.SAdd(languageKey, v)
	}
	categoriesKey := fmt.Sprintf("user:%s:categories", user.GoogleId)
	for _,v := range user.ContentType {
		h.RedisConnection.SAdd(categoriesKey, v)
	}
}

func (h *UserServiceHandler) GetUser(ctx context.Context, request *UserService.GetRequest, response *UserService.User) error {
	return h.MongoCollection.FindOne(context.Background(), bson.D{{"googleid", request.GoogleId}}).Decode(&response)
}

func (h *UserServiceHandler) UpdateUser(ctx context.Context, req *UserService.User, res *UserService.UpdateResponse) error {
	if !h.ValidateUserInformation(req) {
		return errors.New("User Information Invalid")
	}
	_, err := h.MongoCollection.ReplaceOne(context.Background(), bson.D{{"googleid", req.GoogleId}}, &req)
	if err != nil {
		res.IsUpdated = false
	}else {
		res.IsUpdated = true
	}
	return err
}

func (h *UserServiceHandler) DeleteUser(ctx context.Context, request *UserService.DeleteRequest, response *UserService.DeleteReponse) error {
	_, err := h.MongoCollection.DeleteOne(ctx, bson.D{{"googleid", request.GoogleId}})
	if err != nil {
		response.IsDeleted = false
	}else {
		response.IsDeleted = true
	}
	return err
}

func (h *UserServiceHandler) LinkedTvDevice(ctx context.Context, request *UserService.TvDevice, response *UserService.LinkedDeviceResponse) error {
	var cwUser UserService.User
	err := h.MongoCollection.FindOne(ctx, bson.D{{"googleid", request.GoogleId}}).Decode(&cwUser)
	if err != nil {
		return err
	}
	cwUser.LinkedDevices = append(cwUser.LinkedDevices, request.LinkedDevice)
	_, err = h.MongoCollection.ReplaceOne(ctx, bson.D{{"googleid", request.GoogleId}}, cwUser)
	if err != nil {
		response.IsLinkedDeviceFetched = false
		return err
	}
	response.LinkedDevices = cwUser.LinkedDevices
	response.IsLinkedDeviceFetched = true
	return err
}

func (h *UserServiceHandler) RemoveTvDevice(ctx context.Context, request *UserService.RemoveTvDeviceRequest, response *UserService.RemoveTvDeviceResponse) error {
	var cwUser UserService.User
	err := h.MongoCollection.FindOne(context.Background(), bson.D{{"googleid", request.GoogleId}}).Decode(&cwUser)
	if err != nil {
		return err
	}
	// Find and remove
	for i, v := range cwUser.LinkedDevices {
		if v.TvEmac == request.TvEmac {
			cwUser.LinkedDevices = append(cwUser.LinkedDevices[:i], cwUser.LinkedDevices[i+1:]...)
			_, err = h.MongoCollection.ReplaceOne(ctx, bson.D{{"googleid", request.GoogleId}}, cwUser)
			if err != nil {
				return err
			}
			response.IsTvDeviceRemoved = true
			return nil
		}
	}
	response.IsTvDeviceRemoved = false
	return errors.New("The Tv Device not found")
}

func (h *UserServiceHandler) GetLinkedDevices(ctx context.Context, request *UserService.GetRequest, response *UserService.LinkedDeviceResponse) error {
	var cwUser UserService.User
	err := h.MongoCollection.FindOne(ctx, bson.D{{"googleid", request.GoogleId}}).Decode(&cwUser)
	if err != nil {
		return err
	}
	response.LinkedDevices = cwUser.LinkedDevices
	response.IsLinkedDeviceFetched = true
	return nil
}

func(h *UserServiceHandler) ValidateUserInformation(user *UserService.User) bool {
	if(len(user.GoogleId) > 0 && len(user.Genre) > 0 && len(user.Language) > 0 && len(user.ContentType) > 0 && len(user.Email) > 0 && len(user.PhoneNumber) > 0){
		return true
	}else {
		return false
	}
}


