package repo

import (
	"bot_for_modeus/internal/model/dbmodel"
	"bot_for_modeus/internal/repo/mongodb"
	"bot_for_modeus/pkg/mongo"
	"context"
	"go.mongodb.org/mongo-driver/bson"
)

type User interface {
	CreateUser(ctx context.Context, u dbmodel.User) error
	GetUserById(ctx context.Context, userId int64) (dbmodel.User, error)
	DeleteUser(ctx context.Context, userId int64) error
	UpdateData(ctx context.Context, userId int64, data bson.D) error
}

type Repositories struct {
	User
}

func NewRepositories(mongo *mongo.Mongo) *Repositories {
	return &Repositories{
		User: mongodb.NewUserRepo(mongo),
	}
}
