package utils

import (
	mtp_errors "github.com/celestix/gotgproto/errors"
	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/storage"
	"github.com/gotd/td/tg"
	"github.com/pkg/errors"
)

func GetUser(ctx *ext.Context, userId int64) (*tg.User, error) {
	peer := ctx.PeerStorage.GetPeerById(userId)
	if peer.ID == 0 {
		return nil, errors.WithStack(mtp_errors.ErrPeerNotFound)
	}
	if peer.Type == storage.TypeUser.GetInt() {
		users, err := ctx.Raw.UsersGetUsers(ctx, []tg.InputUserClass{&tg.InputUser{
			UserID:     peer.ID,
			AccessHash: peer.AccessHash,
		}})
		if err != nil {
			return nil, errors.WithStack(err)
		}
		if len(users) != 1 {
			return nil, errors.WithStack(mtp_errors.ErrPeerNotFound)
		}
		user, ok := users[0].(*tg.User)
		if !ok {
			return nil, errors.WithStack(mtp_errors.ErrNotUser)
		}

		return user, nil
	} else {
		return nil, errors.WithStack(mtp_errors.ErrNotUser)
	}
}

func GetUserFromPeer(ctx *ext.Context, peer tg.InputPeerClass) (*tg.User, error) {
	var peerId, peerAccessHash int64
	peerUser, ok := peer.(*tg.InputPeerUser)
	if ok {
		peerId = peerUser.UserID
		peerAccessHash = peerUser.AccessHash
	} else {
		peerUserFromMsg, ok := peer.(*tg.InputPeerUserFromMessage)
		if !ok {
			return nil, errors.New("wrong peer type")
		}

		peerUser, ok = peerUserFromMsg.Peer.(*tg.InputPeerUser)
		if !ok {
			return nil, errors.New("wrong peer type")
		}
		peerId = peerUser.UserID
		peerAccessHash = peerUser.AccessHash
	}

	users, err := ctx.Raw.UsersGetUsers(ctx, []tg.InputUserClass{&tg.InputUser{
		UserID:     peerId,
		AccessHash: peerAccessHash,
	}})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if len(users) != 1 {
		return nil, errors.WithStack(mtp_errors.ErrPeerNotFound)
	}
	user, ok := users[0].(*tg.User)
	if !ok {
		return nil, errors.WithStack(mtp_errors.ErrNotUser)
	}

	return user, nil
}
