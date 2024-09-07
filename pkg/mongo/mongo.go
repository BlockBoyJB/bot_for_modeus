package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

const defaultDisconnectTimeout = time.Second * 5

type Pool interface {
	InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error)
	FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult
	//Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (cur *mongo.Cursor, err error)

	UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
	DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error)
}

type Mongo struct {
	client   *mongo.Client
	database *mongo.Database
}

func NewMongo(ctx context.Context, uri, database string) (*Mongo, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	return &Mongo{
		client:   client,
		database: client.Database(database),
	}, nil
}

func (m *Mongo) Disconnect() {
	ctx, cancel := context.WithTimeout(context.Background(), defaultDisconnectTimeout)
	defer cancel()
	_ = m.client.Disconnect(ctx)
}

func (m *Mongo) Drop(ctx context.Context) error {
	return m.database.Drop(ctx)
}

func (m *Mongo) Collection(name string) Pool {
	return m.database.Collection(name)
}
