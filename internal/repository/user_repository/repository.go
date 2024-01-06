package user_repository

import "go.mongodb.org/mongo-driver/mongo"

type Repository struct {
	client               *mongo.Client
	userCollection       *mongo.Collection
	userInChatCollection *mongo.Collection
	database             *mongo.Database
}

const databaseName = "db_main"

func New(client *mongo.Client) (*Repository, error) {
	r := Repository{
		client:   client,
		database: client.Database(databaseName),
	}

	r.userCollection = r.database.Collection(userCollectionName)
	r.userInChatCollection = r.database.Collection(userInChatCollectionName)

	return &r, nil
}
