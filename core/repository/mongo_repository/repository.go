package mongo_repository

import (
	"time"

	"github.com/kamva/mgm/v3"
	"github.com/pkg/errors"
	"github.com/teadove/fun_telegram/core/shared"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Repository struct {
	messageCollection *mgm.Collection
	userCollection    *mgm.Collection
	memberCollection  *mgm.Collection
	chatCollection    *mgm.Collection

	client                   *mongo.Client
	restartMessageCollection *mgm.Collection
}

const databaseName = "db_main"

func New() (*Repository, error) {
	r := Repository{}

	err := mgm.SetDefaultConfig(
		&mgm.Config{CtxTimeout: 12 * time.Second},
		databaseName,
		options.Client().ApplyURI(shared.AppSettings.Storage.MongoDbUrl),
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	r.messageCollection = mgm.Coll(&Message{})
	r.userCollection = mgm.Coll(&User{})
	r.restartMessageCollection = mgm.Coll(&RestartMessage{})
	r.memberCollection = mgm.Coll(&Member{})
	r.chatCollection = mgm.Coll(&Chat{})

	r.client, err = mgm.NewClient(options.Client().ApplyURI(shared.AppSettings.Storage.MongoDbUrl))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &r, nil
}
