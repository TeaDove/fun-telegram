package db_repository

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/guregu/null/v5"
)

type Channel struct {
	WithId
	WithCreatedAt

	TgId       int64  `sql:"tg_id"       parquet:"name=tg_id, type=INT64"`
	TgTitle    string `sql:"tg_title"    parquet:"name=tg_title, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	TgUsername string `sql:"tg_username" parquet:"name=tg_username, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`

	ParticipantCount   int64       `sql:"participant_count"   parquet:"name=participant_count, type=INT64"`
	RecommendationsIds []int64     `sql:"recommendations_ids"`
	IsLeaf             bool        `sql:"is_leaf" parquet:"name=is_leaf, type=BOOLEAN"`
	TgAbout            null.String `sql:"tg_about" parquet:"name=tg_about, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
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
	WithId
	WithCreatedAt

	TgIdIn  int64 `csv:"tg_id_in"  ch:"tg_id_in"  parquet:"name=tg_id_in, type=INT64"`
	TgIdOut int64 `csv:"tg_id_out" ch:"tg_id_out" parquet:"name=tg_id_out, type=INT64"`
	Order   int64 `csv:"order"     ch:"order"     parquet:"name=order, type=INT64"`
}

type ChannelsEdges []ChannelEdge

func (r ChannelsEdges) ToOutIds() []int64 {
	ids := make([]int64, 0, len(r))
	for _, channelsEdge := range r {
		ids = append(ids, channelsEdge.TgIdOut)
	}

	return ids
}

func (r ChannelsEdges) ToIds() []int64 {
	ids := mapset.NewSet[int64]()
	for _, channelsEdge := range r {
		ids.Append(channelsEdge.TgIdOut, channelsEdge.TgIdIn)
	}

	return ids.ToSlice()
}
