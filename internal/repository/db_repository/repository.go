package db_repository

import (
	"github.com/kamva/mgm/v3"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type Repository struct {
	messageCollection *mgm.Collection
	userCollection    *mgm.Collection
}

func New(mongoUrl string) (*Repository, error) {
	const databaseName = "db_main"

	r := Repository{}

	err := mgm.SetDefaultConfig(
		&mgm.Config{CtxTimeout: 12 * time.Second},
		databaseName,
		options.Client().ApplyURI(mongoUrl),
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	r.messageCollection = mgm.Coll(&Message{})
	r.userCollection = mgm.Coll(&User{})

	return &r, nil
}
