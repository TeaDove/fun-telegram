package analitics

import (
	"github.com/stretchr/testify/assert"
	"github.com/teadove/goteleout/internal/shared"
	"math/rand/v2"
	"sync"
	"testing"
	"time"
)

func generateMessage() Message {
	return Message{
		CreatedAt: time.Now().UTC(),
		TgChatID:  rand.Int64N(100000),
		TgId:      rand.IntN(100000),
		TgUserId:  rand.Int64N(100000),
		Text:      shared.RandomString(),
	}
}

func TestIntegration_AnaliticsService_InsertNewMessage_Ok(t *testing.T) {
	r := getService(t)
	message := generateMessage()

	err := r.InsertNewMessage(shared.GetCtx(), &message)
	assert.NoError(t, err)
}

func TestIntegration_AnaliticsService_InsertManyMessage_Ok(t *testing.T) {
	r := getService(t)
	var wg sync.WaitGroup

	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			message := generateMessage()

			err := r.InsertNewMessage(shared.GetCtx(), &message)
			assert.NoError(t, err)
		}()
	}
	wg.Wait()
}
