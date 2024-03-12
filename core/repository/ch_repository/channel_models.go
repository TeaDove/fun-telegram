package ch_repository

import (
	"github.com/guregu/null/v5"
	"time"
)

type Channel struct {
	TgId       int64     `csv:"tg_id" ch:"tg_id"`
	TgTitle    string    `csv:"tg_title" ch:"tg_title"`
	TgUsername string    `csv:"tg_username" ch:"tg_username"`
	UploadedAt time.Time `csv:"uploaded_at" ch:"uploaded_at"`

	ParticipantCount   int64       `csv:"participant_count" ch:"participant_count"`
	RecommendationsIds []int64     `csv:"recommendations_ids" ch:"recommendations_ids"`
	IsLeaf             bool        `csv:"is_leaf" ch:"is_leaf"`
	TgAbout            null.String `csv:"tg_about" ch:"tg_about"`
}

type Channels []Channel

func (r Channels) ToMap() map[int64]Channel {
	map_ := make(map[int64]Channel, len(r))
	for _, channel := range r {
		map_[channel.TgId] = channel
	}

	return map_
}

type ChannelEdge struct {
	TgIdIn  int64 `csv:"tg_id_in" ch:"tg_id_in"`
	TgIdOut int64 `csv:"tg_id_out" ch:"tg_id_out"`
	Order   int64 `csv:"order" ch:"order"`
}
