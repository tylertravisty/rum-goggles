package models

import (
	"database/sql"
	"fmt"
)

const (
	accountColumns = "id, username, cookies"
	accountTable   = "account"
)

type Account struct {
	ID       *int64
	Username *string
	Cookies  *string
}

type sqlAccount struct {
	id       sql.NullInt64
	username sql.NullString
	cookies  sql.NullString
}

func (sa *sqlAccount) scan(r Row) error {
	return r.Scan(&sa.id, &sa.username, &sa.cookies)
}

func (sa sqlAccount) toAccount() *Account {
	var a Account
	a.ID = toInt64(sa.id)
	a.Username = toString(sa.username)
	a.Cookies = toString(sa.cookies)

	return &a
}

type AccountService interface {
	AutoMigrate() error
	ByUsername(username string) (*Account, error)
	Create(a *Account) error
	DestructiveReset() error
	Update(a *Account) error
}

func NewAccountService(db *sql.DB) AccountService {
	return &accountService{
		Database: db,
	}
}

var _ AccountService = &accountService{}

type accountService struct {
	Database *sql.DB
}

func (as *accountService) AutoMigrate() error {
	err := as.createAccountTable()
	if err != nil {
		return err
	}

	return nil
}

func (as *accountService) createAccountTable() error {
	createQ := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS "%s" (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			cookies TEXT
		)
	`, accountTable)

	_, err := as.Database.Exec(createQ)
	if err != nil {
		return fmt.Errorf("error creating table: %v", err)
	}

	return nil
}

func (as *accountService) ByUsername(username string) (*Account, error) {
	err := runAccountValFuncs(
		&Account{Username: &username},
		accountRequireUsername,
	)
	if err != nil {
		return nil, pkgErr("", err)
	}

	selectQ := fmt.Sprintf(`
		SELECT %s
		FROM "%s"
		WHERE username=?
	`, accountColumns, accountTable)

	var sa sqlAccount
	row := as.Database.QueryRow(selectQ, username)
	err = sa.scan(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, pkgErr(fmt.Sprintf("error querying \"%s\" by username", accountTable), err)
	}

	return sa.toAccount(), nil
}

func (as *accountService) Create(a *Account) error {
	err := runAccountValFuncs(
		a,
		accountRequireUsername,
	)
	if err != nil {
		return pkgErr("invalid account", err)
	}

	columns := columnsNoID(accountColumns)
	insertQ := fmt.Sprintf(`
		INSERT INTO "%s" (%s)
		VALUES (%s)
	`, accountTable, columns, values(columns))

	_, err = as.Database.Exec(insertQ, a.Username, a.Cookies)
	if err != nil {
		return pkgErr(fmt.Sprintf("error inserting %s", accountTable), err)
	}

	return nil
}

func (as *accountService) DestructiveReset() error {
	dropQ := fmt.Sprintf(`
		DROP TABLE IF EXISTS "%s"
	`, accountTable)

	_, err := as.Database.Exec(dropQ)
	if err != nil {
		return fmt.Errorf("error dropping table: %v", err)
	}

	return nil
}

func (as *accountService) Update(a *Account) error {
	err := runAccountValFuncs(
		a,
		accountRequireID,
		accountRequireUsername,
	)
	if err != nil {
		return pkgErr("invalid account", err)
	}

	columns := columnsNoID(accountColumns)
	updateQ := fmt.Sprintf(`
		UPDATE "%s"
		SET %s
		WHERE id=?
	`, accountTable, set(columns))

	_, err = as.Database.Exec(updateQ, a.Username, a.Cookies, a.ID)
	if err != nil {
		return pkgErr(fmt.Sprintf("error updating %s", accountTable), err)
	}

	return nil
}

type accountValFunc func(*Account) error

func runAccountValFuncs(a *Account, fns ...accountValFunc) error {
	if a == nil {
		return fmt.Errorf("account cannot be nil")
	}

	for _, fn := range fns {
		err := fn(a)
		if err != nil {
			return err
		}
	}

	return nil
}

func accountRequireID(a *Account) error {
	if a.ID == nil || *a.ID < 1 {
		return ErrAccountInvalidID
	}

	return nil
}

func accountRequireUsername(a *Account) error {
	if a.Username == nil || *a.Username == "" {
		return ErrAccountInvalidUsername
	}

	return nil
}
