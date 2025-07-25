package pg

import (
	"time"

	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"github.com/pkg/errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewClientFromSettings() (*gorm.DB, error) {
	pgConfig := postgres.Config{
		DSN:                  "postgresql://main:main@localhost:5432/main",
		PreferSimpleProtocol: true,
	}

	db, err := gorm.Open(postgres.New(pgConfig), &gorm.Config{
		SkipDefaultTransaction: true,
		TranslateError:         true,
		NowFunc: func() time.Time {
			ti, _ := time.LoadLocation("utc")
			return time.Now().In(ti)
		},
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},

		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to database")
	}

	sqlDb, err := db.DB()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get std db")
	}

	sqlDb.SetMaxIdleConns(10)
	sqlDb.SetMaxOpenConns(30)

	return db, nil
}
