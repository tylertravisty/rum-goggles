package models

import (
	"database/sql"
	"fmt"
)

type migrationFunc func() error

type service struct {
	name             string
	automigrate      migrationFunc
	destructivereset migrationFunc
}

type Services struct {
	AccountS AccountService
	Database *sql.DB
	services []service
}

func (s *Services) AutoMigrate() error {
	for _, service := range s.services {
		err := service.automigrate()
		if err != nil {
			return pkgErr(fmt.Sprintf("error auto-migrating %s service", service.name), err)
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
	for _, service := range s.services {
		err := service.destructivereset()
		if err != nil {
			return pkgErr(fmt.Sprintf("error destructive-resetting %s service", service.name), err)
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
		s.services = append(s.services, service{accountTable, s.AccountS.AutoMigrate, s.AccountS.DestructiveReset})

		return nil
	}
}
