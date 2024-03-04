package ch_repository

import "time"

type Channel struct {
	TgId       int64     `json:"tg_id" ch:"tg_id"`
	TgTitle    string    `json:"tg_title" ch:"tg_title"`
	TgUsername string    `json:"tg_username" ch:"tg_username"`
	UploadedAt time.Time `json:"uploaded_at" ch:"uploaded_at"`

	ParticipantCount   int64   `json:"participant_count" ch:"participant_count"`
	RecommendationsIds []int64 `json:"recommendations_ids" ch:"recommendations_ids"`
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
	TgIdIn  int64 `json:"tg_id_in" ch:"tg_id_in"`
	TgIdOut int64 `json:"tg_id_out" ch:"tg_id_out"`
	Order   int64 `json:"order" ch:"order"`
}
