package user_repository

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIntegration_MongoDB_MongoDBConnect_Ok(t *testing.T) {
	client, err := MongoDBConnect(context.Background(), "mongodb://localhost:27017")
	assert.NoError(t, err)

	err = client.Ping(context.Background(), nil)
	assert.NoError(t, err)
}
