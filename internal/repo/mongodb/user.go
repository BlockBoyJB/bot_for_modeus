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
	pool mongo.Pool
}

func NewUserRepo(mongo *mongo.Mongo) *UserRepo {
	return &UserRepo{mongo.Collection("user")}
}

func (r *UserRepo) Create(ctx context.Context, u dbmodel.User) error {
	if _, err := r.pool.InsertOne(ctx, u); err != nil {
		return err
	}
	return nil
}

func (r *UserRepo) FindById(ctx context.Context, id int64) (dbmodel.User, error) {
	var user dbmodel.User

	if err := r.pool.FindOne(ctx, bson.D{{"user_id", id}}).Decode(&user); err != nil {
		if errors.Is(err, mgo.ErrNoDocuments) {
			return dbmodel.User{}, mongoerrs.ErrNotFound
		}
		return dbmodel.User{}, err
	}
	return user, nil
}

func (r *UserRepo) Update(ctx context.Context, id int64, data bson.D) error {
	c, err := r.pool.UpdateOne(ctx, bson.D{{"user_id", id}}, data)
	if err != nil {
		return err
	}
	if c.MatchedCount == 0 && c.ModifiedCount == 0 {
		return mongoerrs.ErrNotFound
	}
	return nil
}

func (r *UserRepo) Delete(ctx context.Context, id int64) error {
	c, err := r.pool.DeleteOne(ctx, bson.D{{"user_id", id}})
	if err != nil {
		return err
	}
	if c.DeletedCount == 0 {
		return mongoerrs.ErrNotFound
	}
	return nil
}
