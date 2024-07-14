package db_repository

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm/clause"
)

func (r *Repository) UserSelectById(ctx context.Context, tgId int64) (User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("tg_id = ?", tgId).Limit(1).Find(&user).Error
	if err != nil {
		return User{}, errors.Wrap(err, "failed to get user by tg_id")
	}

	return user, nil
}

func (r *Repository) UserSelectByUsername(ctx context.Context, tgUsername string) (User, error) {
	var user User

	err := r.db.WithContext(ctx).Where("tg_username = ?", tgUsername).First(&user).Error
	if err != nil {
		return User{}, errors.Wrap(err, "failed to get user by tg username")
	}

	return user, nil
}

func (r *Repository) UsersSelectByStatusInChat(
	ctx context.Context,
	tgChatId int64,
	memberStatuses []MemberStatus,
) (UsersInChat, error) {
	var usersInChat UsersInChat

	err := r.db.WithContext(ctx).
		Raw(`
select u.tg_id, u.tg_username, u.tg_name, u.is_bot, m.status 
	from "user" u 
join member m on u.tg_id = m.tg_user_id
	where m.tg_chat_id = ? and m.status in (?)
`, tgChatId, memberStatuses).
		Scan(&usersInChat).
		Error
	if err != nil {
		return nil, errors.Wrap(err, "failed to get users")
	}

	return usersInChat, nil
}

func (r *Repository) UserUpsert(ctx context.Context, user *User) error {
	user.UpdatedInDBAt = time.Now().UTC()

	err := r.db.WithContext(ctx).Clauses(
		clause.OnConflict{
			Columns: []clause.Column{{Name: "tg_username"}, {Name: "tg_id"}},
			DoUpdates: clause.AssignmentColumns(
				[]string{"tg_id", "tg_username", "tg_name", "is_bot", "updated_in_db_at"},
			),
		}).
		Create(&user).Error
	if err != nil {
		return errors.Wrap(err, "failed to upsert chat")
	}

	return nil

	//user.UpdatedAt = time.Now().UTC()
	//
	//filter := bson.M{"tg_id": user.TgId}
	//update := bson.M{"$set": bson.M{
	//	"tg_id":       user.TgId,
	//	"tg_username": user.TgUsername,
	//	"tg_name":     user.TgName,
	//	"updated_at":  user.UpdatedAt,
	//	"created_at":  user.CreatedAt,
	//	"is_bot":      user.IsBot,
	//}}
	//opts := options.Update().SetUpsert(true)
	//
	//_, err := r.userCollection.UpdateOne(ctx, filter, update, opts)
	//if err != nil {
	//	return errors.WithStack(err)
	//}

	return nil
}
