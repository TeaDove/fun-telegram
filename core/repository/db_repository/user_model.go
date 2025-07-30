package db_repository

type User struct {
	WithID
	WithCreatedAt
	WithUpdatedInDBAt

	TgID       int64  `sql:"tg_id"       gorm:"index:_,unique"`
	TgUsername string `sql:"tg_username" gorm:"index:_,unique"`
	TgName     string `sql:"tg_name"     gorm:"index"`
	IsBot      bool   `sql:"is_bot"`
}

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

var MemberStatusesActive = []MemberStatus{Plain, Creator, Admin} //nolint: gochecknoglobals // FIXME

type Member struct {
	WithID
	WithCreatedInDBAt
	WithUpdatedInDBAt

	TgUserID int64 `sql:"tg_user_id" gorm:"index:tg_user_id_tg_chat_id_idx,unique"`
	TgChatID int64 `sql:"tg_chat_id" gorm:"index:tg_user_id_tg_chat_id_idx,unique"`
	Status   MemberStatus
}
