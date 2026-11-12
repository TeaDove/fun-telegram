package tgdtos

//go:generate go run github.com/abice/go-enum@v0.9.5 --file $GOFILE --nocomments

// ENUM(Message=MESSAGE, Author=AUTHOR, HistoryHeader=HISTORY_HEADER).
type TGType string
