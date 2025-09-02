package message_service

import (
	"time"

	"github.com/guregu/null/v5"
)

type Message struct {
	ID        uint
	CreatedAt time.Time

	TgChatID int64
	TgID     int

	TgUserID        int64
	Text            string
	WordsCount      uint64
	ToxicWordsCount uint64

	ReplyToTgMsgID  null.Int64
	ReplyToTgUserID null.Int64
}

type Messages []Message

type UserInChat struct {
	TgID       int64
	TgUsername string
	TgName     string
	IsBot      bool
	Status     MemberStatus
}

type UsersInChat []UserInChat

type MemberStatus string

const (
	Plain MemberStatus = "PLAIN"
	// Creator is status for chat/channel creator.
	Creator MemberStatus = "CREATOR"
	// Admin is status for chat/channel admin.
	Admin MemberStatus = "ADMIN"
	// Banned is status for banned user.
	Banned MemberStatus = "BANNED"
	// Left is status for user that left chat/channel.
	Left MemberStatus = "LEFT"

	Unknown MemberStatus = "UNKNOWN"
)

type Storage struct {
	Messages Messages
	Users    UsersInChat
}
