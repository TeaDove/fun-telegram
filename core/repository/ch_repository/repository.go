package ch_repository

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/goteleout/core/shared"
)

type Repository struct {
	conn         driver.Conn
	databaseName string
}

func New(ctx context.Context) (*Repository, error) {
	r := Repository{databaseName: "default"}
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{shared.AppSettings.Storage.ClickhouseUtl},
		Auth: clickhouse.Auth{
			Database: r.databaseName,
			Username: "default",
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60 * 3,
		},
		Protocol: clickhouse.Native,
		//Debug:    true,
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
				{Name: "FunTelegram", Version: shared.Undefined},
			},
		},
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	r.conn = conn
	go r.init(ctx)

	return &r, nil
}

func (r *Repository) Ping(ctx context.Context) error {
	err := r.conn.Ping(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (r *Repository) init(ctx context.Context) {
	for _, sql := range initSQL {
		err := r.conn.Exec(ctx, sql)
		if err != nil {
			zerolog.Ctx(ctx).Error().Stack().Err(err).Str("status", "failed.to.run.init.sql").Send()
		}
	}
}
