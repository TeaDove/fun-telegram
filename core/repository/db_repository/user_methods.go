package db_repository

import "context"

func (r *Repository) UserGetById(ctx context.Context, userId int64) (User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("tg_id = ?", userId).Limit(1).Find(&user).Error
	if err != nil {
		return User{}, err
	}

	return user, nil
}
