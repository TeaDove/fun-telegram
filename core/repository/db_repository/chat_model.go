package db_repository

type Chat struct {
	WithID
	WithCreatedAt
	WithUpdatedAt

	TgID  int64  `sql:"tg_id" gorm:"index:,unique"`
	Title string `            gorm:"index"`
}

type UserInChat struct {
	TgID       int64
	TgUsername string
	TgName     string
	IsBot      bool
	Status     MemberStatus
}

type UsersInChat []UserInChat

func (r UsersInChat) ToMap() map[int64]UserInChat {
	v := make(map[int64]UserInChat, len(r))
	for _, user := range r {
		v[user.TgID] = user
	}

	return v
}

func (r UsersInChat) ToIDs() []int64 {
	var slice []int64
	for _, user := range r {
		slice = append(slice, user.TgID)
	}

	return slice
}
