package tgdtos

//go:generate go tool go-enum --file $GOFILE --marshal --forceupper --nocomments

// ENUM(message=1, author, history_header).
type TGType uint8
