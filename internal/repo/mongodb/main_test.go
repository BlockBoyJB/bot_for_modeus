package mongodb

import (
	"bot_for_modeus/pkg/mongo"
	"context"
	"github.com/stretchr/testify/suite"
	"testing"
)

type mongodbTestSuite struct {
	suite.Suite
	ctx   context.Context
	mongo *mongo.Mongo
	user  *UserRepo
}

func (s *mongodbTestSuite) SetupTest() {
	testMongoUrl := "mongodb://localhost:27017"
	ctx := context.Background()
	mongodb, err := mongo.NewMongo(ctx, testMongoUrl, "test")
	if err != nil {
		panic(err)
	}
	s.mongo = mongodb
	s.ctx = ctx

	s.user = NewUserRepo(mongodb)
}

func (s *mongodbTestSuite) TearDownTest() {
	_ = s.mongo.Drop(s.ctx)
	s.mongo.Disconnect()
}

func TestMongoDB(t *testing.T) {
	suite.Run(t, new(mongodbTestSuite))
}
