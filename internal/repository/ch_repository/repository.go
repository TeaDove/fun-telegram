package ch_repository

import (
	"context"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/goteleout/internal/shared"
	"github.com/teadove/goteleout/internal/utils"
	"time"
)

type Repository struct {
	conn driver.Conn
}

func New(ctx context.Context) (*Repository, error) {
	r := Repository{}
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{shared.AppSettings.Storage.ClickhouseUtl},
		Auth: clickhouse.Auth{
			Database: "default",
			Username: "default",
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60 * 3,
		},
		Protocol: clickhouse.Native,
		Debug:    false,
		Debugf: func(format string, v ...any) {
			zerolog.Ctx(ctx).Debug().Str("status", "ch.log").Str("log", fmt.Sprintf(format, v...)).Send()
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		DialTimeout:          time.Second * 30,
		MaxOpenConns:         5,
		MaxIdleConns:         5,
		ConnMaxLifetime:      time.Duration(20) * time.Minute,
		ConnOpenStrategy:     clickhouse.ConnOpenInOrder,
		BlockBufferSize:      10,
		MaxCompressionBuffer: 10240,
		ClientInfo: clickhouse.ClientInfo{
			Products: []struct {
				Name    string
				Version string
			}{
				{Name: "FunTelegram", Version: utils.Undefined},
			},
		},
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	r.conn = conn

	return &r, nil
}

func (r *Repository) Ping(ctx context.Context) error {
	err := r.conn.Ping(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
