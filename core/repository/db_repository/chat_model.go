package db_repository

type Chat struct {
	WithId
	WithCreatedAt
	WithUpdatedAt

	TgId  int64  `sql:"tg_id" gorm:"index:,unique"`
	Title string `            gorm:"index"`
}

type UserInChat struct {
	TgId       int64
	TgUsername string
	TgName     string
	IsBot      bool
	Status     MemberStatus
}

type UsersInChat []UserInChat

func (r UsersInChat) ToMap() map[int64]UserInChat {
	map_ := make(map[int64]UserInChat, len(r))
	for _, user := range r {
		map_[user.TgId] = user
	}

	return map_
}

func (r UsersInChat) ToIds() []int64 {
	slice := make([]int64, len(r))
	for _, user := range r {
		slice = append(slice, user.TgId)
	}

	return slice
}
