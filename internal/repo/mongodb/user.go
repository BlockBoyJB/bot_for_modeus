package mongodb

import (
	"bot_for_modeus/internal/model/dbmodel"
	"bot_for_modeus/internal/repo/mongoerrs"
	"bot_for_modeus/pkg/mongo"
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	mgo "go.mongodb.org/mongo-driver/mongo"
)

type UserRepo struct {
	*mongo.Mongo
}

func NewUserRepo(mongo *mongo.Mongo) *UserRepo {
	return &UserRepo{mongo}
}

func (r *UserRepo) CreateUser(ctx context.Context, u dbmodel.User) error {
	if _, err := r.Pool.InsertOne(ctx, u); err != nil {
		return err
	}
	return nil
}

func (r *UserRepo) GetUserById(ctx context.Context, userId int64) (dbmodel.User, error) {
	filter := bson.D{{"user_id", userId}}

	var user dbmodel.User

	err := r.Pool.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mgo.ErrNoDocuments) {
			return dbmodel.User{}, mongoerrs.ErrNotFound
		}

		return dbmodel.User{}, err
	}
	return user, nil
}

func (r *UserRepo) DeleteUser(ctx context.Context, userId int64) error {
	filter := bson.D{{"user_id", userId}}
	_, err := r.Pool.DeleteOne(ctx, filter)
	if err != nil {
		if errors.Is(err, mgo.ErrNoDocuments) {
			return mongoerrs.ErrNotFound
		}

		return err
	}
	return nil
}

func (r *UserRepo) UpdateData(ctx context.Context, userId int64, data bson.D) error {
	filter := bson.D{{"user_id", userId}}
	if _, err := r.Pool.UpdateOne(ctx, filter, data); err != nil {
		if errors.Is(err, mgo.ErrNoDocuments) {
			return mongoerrs.ErrNotFound
		}
		return err
	}
	return nil
}
