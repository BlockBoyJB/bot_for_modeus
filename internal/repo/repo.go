package repo

import (
	"bot_for_modeus/internal/model/dbmodel"
	"bot_for_modeus/internal/repo/mongodb"
	"bot_for_modeus/pkg/mongo"
	"context"
	"go.mongodb.org/mongo-driver/bson"
)

type User interface {
	Create(ctx context.Context, u dbmodel.User) error
	FindById(ctx context.Context, id int64) (dbmodel.User, error)
	Update(ctx context.Context, id int64, data bson.D) error
	Delete(ctx context.Context, id int64) error
}

type Repositories struct {
	User
}

func NewRepositories(mongo *mongo.Mongo) *Repositories {
	return &Repositories{
		User: mongodb.NewUserRepo(mongo),
	}
}
