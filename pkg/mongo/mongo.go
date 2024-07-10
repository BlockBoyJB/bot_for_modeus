package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

const defaultDisconnectTimeout = time.Second * 5

type mongoPool interface {
	InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error)
	FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult
	//Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (cur *mongo.Cursor, err error)

	UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
	DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error)
}

type Mongo struct {
	Pool   mongoPool
	client *mongo.Client
}

// Пока что черновой вариант взаимодействия с бд. Да и да текущем этапе используется только одна коллекция

func NewMongo(ctx context.Context, url, database, collection string) (*Mongo, error) {
	o := options.Client().ApplyURI(url)
	client, err := mongo.Connect(ctx, o)
	if err != nil {
		return nil, err
	}
	return &Mongo{
		Pool:   client.Database(database).Collection(collection),
		client: client,
	}, nil
}

func (m *Mongo) Disconnect() {
	ctx, cancel := context.WithTimeout(context.Background(), defaultDisconnectTimeout)
	defer cancel()
	_ = m.client.Disconnect(ctx)
}
