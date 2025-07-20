package analitics

import (
	"math/rand/v2"
	"sync"
	"testing"
	"time"

	"github.com/teadove/teasutils/utils/random_utils"
	"github.com/teadove/teasutils/utils/test_utils"

	"github.com/stretchr/testify/assert"
)

func generateMessage() Message {
	return Message{
		CreatedAt: time.Now().UTC(),
		TgChatID:  rand.Int64N(100000),
		TgId:      rand.IntN(100000),
		TgUserId:  rand.Int64N(100000),
		Text:      random_utils.Text(),
	}
}

func TestIntegration_AnaliticsService_InsertNewMessage_Ok(t *testing.T) {
	r := getService(t)
	message := generateMessage()

	err := r.MessageInsert(test_utils.GetLoggedContext(), &message)
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

			err := r.MessageInsert(test_utils.GetLoggedContext(), &message)
			assert.NoError(t, err)
		}()
	}
	wg.Wait()
}
