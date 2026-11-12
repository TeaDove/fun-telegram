package messageservice

import "github.com/TeaDove/fun-telegram/internal/dtos/tgdtos"

type Message struct {
	tgdtos.Message

	WordsCount      uint64
	ToxicWordsCount uint64
}

type Messages []Message

type UsersInChat []tgdtos.Author

type Storage struct {
	Messages Messages

	Users           UsersInChat
	UsersNameGetter NameGetter
}
