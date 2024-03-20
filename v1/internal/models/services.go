package models

import (
	"database/sql"
	"fmt"
)

type migrationFunc func() error

type table struct {
	name             string
	automigrate      migrationFunc
	destructivereset migrationFunc
}

type Services struct {
	AccountS        AccountService
	AccountChannelS AccountChannelService
	ChannelS        ChannelService
	Database        *sql.DB
	tables          []table
}

func (s *Services) AutoMigrate() error {
	for _, table := range s.tables {
		if table.automigrate != nil {
			err := table.automigrate()
			if err != nil {
				return pkgErr(fmt.Sprintf("error auto-migrating %s table", table.name), err)
			}
		}
	}

	return nil
}

func (s *Services) Close() error {
	err := s.Database.Close()
	if err != nil {
		return pkgErr("error closing database", err)
	}

	return nil
}

func (s *Services) DestructiveReset() error {
	for _, table := range s.tables {
		if table.destructivereset != nil {
			err := table.destructivereset()
			if err != nil {
				return pkgErr(fmt.Sprintf("error destructive-resetting %s table", table.name), err)
			}
		}
	}

	return nil
}

type ServicesInit func(*Services) error

func NewServices(inits ...ServicesInit) (*Services, error) {
	var s Services
	for _, init := range inits {
		err := init(&s)
		if err != nil {
			return nil, err
		}
	}
	return &s, nil
}

func WithDatabase(file string) ServicesInit {
	return func(s *Services) error {
		db, err := sql.Open("sqlite3", "file:"+file+"?_foreign_keys=ON")
		if err != nil {
			return pkgErr("error opening database", err)
		}

		s.Database = db
		return nil
	}
}

func WithAccountService() ServicesInit {
	return func(s *Services) error {
		s.AccountS = NewAccountService(s.Database)
		s.tables = append(s.tables, table{accountTable, s.AccountS.AutoMigrate, s.AccountS.DestructiveReset})

		return nil
	}
}

func WithAccountChannelService() ServicesInit {
	return func(s *Services) error {
		s.AccountChannelS = NewAccountChannelService(s.Database)

		return nil
	}
}

func WithChannelService() ServicesInit {
	return func(s *Services) error {
		s.ChannelS = NewChannelService(s.Database)
		s.tables = append(s.tables, table{channelTable, s.ChannelS.AutoMigrate, s.ChannelS.DestructiveReset})

		return nil
	}
}
