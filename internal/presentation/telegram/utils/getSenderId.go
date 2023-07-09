package utils

import (
	"errors"
	"github.com/gotd/td/tg"
)

func GetSenderId(m *tg.Message) (int64, error) {
	peer, ok := m.GetFromID()
	if !ok {
		peer = m.PeerID
	}

	switch t := peer.(type) {
	case *tg.PeerUser:
		return t.UserID, nil
	default:
		return 0, errors.New("invalid peer")
	}
}
