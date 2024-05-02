package models

import (
	"database/sql"
	"fmt"
)

const (
	accountColumns = "id, uid, username, cookies, profile_image, api_key"
	accountTable   = "account"
)

type Account struct {
	ID           *int64  `json:"id"`
	UID          *string `json:"uid"`
	Username     *string `json:"username"`
	Cookies      *string `json:"cookies"`
	ProfileImage *string `json:"profile_image"`
	ApiKey       *string `json:"api_key"`
}

func (a *Account) Id() *int64 {
	return a.ID
}

func (a *Account) KeyUrl() *string {
	return a.ApiKey
}

func (a *Account) LoggedIn() bool {
	return a.Cookies != nil
}

func (a *Account) String() *string {
	if a.Username == nil {
		return nil
	}

	s := "/user/" + *a.Username
	return &s
}

func (a *Account) Title() *string {
	return a.Username
}

func (a *Account) Type() string {
	return "Account"
}

func (a *Account) values() []any {
	return []any{a.ID, a.UID, a.Username, a.Cookies, a.ProfileImage, a.ApiKey}
}

func (a *Account) valuesNoID() []any {
	return a.values()[1:]
}

func (a *Account) valuesEndID() []any {
	vals := a.values()
	return append(vals[1:], vals[0])
}

type sqlAccount struct {
	id           sql.NullInt64
	uid          sql.NullString
	username     sql.NullString
	cookies      sql.NullString
	profileImage sql.NullString
	apiKey       sql.NullString
}

func (sa *sqlAccount) scan(r Row) error {
	return r.Scan(&sa.id, &sa.uid, &sa.username, &sa.cookies, &sa.profileImage, &sa.apiKey)
}

func (sa sqlAccount) toAccount() *Account {
	var a Account
	a.ID = toInt64(sa.id)
	a.UID = toString(sa.uid)
	a.Username = toString(sa.username)
	a.Cookies = toString(sa.cookies)
	a.ProfileImage = toString(sa.profileImage)
	a.ApiKey = toString(sa.apiKey)

	return &a
}

type AccountService interface {
	All() ([]Account, error)
	AutoMigrate() error
	ByID(id int64) (*Account, error)
	ByUsername(username string) (*Account, error)
	Create(a *Account) (int64, error)
	Delete(a *Account) error
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

func (as *accountService) All() ([]Account, error) {
	selectQ := fmt.Sprintf(`
		SELECT %s
		FROM "%s"
	`, accountColumns, accountTable)

	rows, err := as.Database.Query(selectQ)
	if err != nil {
		return nil, pkgErr("error executing select query", err)
	}
	defer rows.Close()

	accounts := []Account{}
	for rows.Next() {
		sa := &sqlAccount{}

		err = sa.scan(rows)
		if err != nil {
			return nil, pkgErr("error scanning row", err)
		}

		accounts = append(accounts, *sa.toAccount())
	}
	err = rows.Err()
	if err != nil && err != sql.ErrNoRows {
		return nil, pkgErr("error iterating over rows", err)
	}

	return accounts, nil
}

func (as *accountService) AutoMigrate() error {
	err := as.createAccountTable()
	if err != nil {
		return pkgErr(fmt.Sprintf("error creating %s table", accountTable), err)
	}

	return nil
}

func (as *accountService) createAccountTable() error {
	createQ := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS "%s" (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			uid TEXT UNIQUE,
			username TEXT UNIQUE NOT NULL,
			cookies TEXT,
			profile_image TEXT,
			api_key TEXT
		)
	`, accountTable)

	_, err := as.Database.Exec(createQ)
	if err != nil {
		return fmt.Errorf("error executing create query: %v", err)
	}

	return nil
}

func (as *accountService) ByID(id int64) (*Account, error) {
	err := runAccountValFuncs(
		&Account{ID: &id},
		accountRequireID,
	)
	if err != nil {
		return nil, pkgErr("", err)
	}

	selectQ := fmt.Sprintf(`
		SELECT %s
		FROM "%s"
		WHERE id=?
	`, accountColumns, accountTable)

	var sa sqlAccount
	row := as.Database.QueryRow(selectQ, id)
	err = sa.scan(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, pkgErr("error executing select query", err)
	}

	return sa.toAccount(), nil
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
		return nil, pkgErr("error executing select query", err)
	}

	return sa.toAccount(), nil
}

func (as *accountService) Create(a *Account) (int64, error) {
	err := runAccountValFuncs(
		a,
		accountRequireUsername,
	)
	if err != nil {
		return -1, pkgErr("invalid account", err)
	}

	columns := columnsNoID(accountColumns)
	insertQ := fmt.Sprintf(`
		INSERT INTO "%s" (%s)
		VALUES (%s)
		RETURNING id
	`, accountTable, columns, values(columns))

	// _, err = as.Database.Exec(insertQ, a.valuesNoID()...)
	var id int64
	row := as.Database.QueryRow(insertQ, a.valuesNoID()...)
	err = row.Scan(&id)
	if err != nil {
		return -1, pkgErr("error executing insert query", err)
	}

	return id, nil
}

func (as *accountService) Delete(a *Account) error {
	err := runAccountValFuncs(
		a,
		accountRequireID,
	)
	if err != nil {
		return pkgErr("invalid account", err)
	}

	deleteQ := fmt.Sprintf(`
		DELETE FROM "%s"
		WHERE id=?
	`, accountTable)

	_, err = as.Database.Exec(deleteQ, a.ID)
	if err != nil {
		return pkgErr("error executing delete query", err)
	}

	return nil
}

func (as *accountService) DestructiveReset() error {
	err := as.dropAccountTable()
	if err != nil {
		return pkgErr(fmt.Sprintf("error dropping %s table", accountTable), err)
	}

	return nil
}

func (as *accountService) dropAccountTable() error {
	dropQ := fmt.Sprintf(`
		DROP TABLE IF EXISTS "%s"
	`, accountTable)

	_, err := as.Database.Exec(dropQ)
	if err != nil {
		return fmt.Errorf("error executing drop query: %v", err)
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

	_, err = as.Database.Exec(updateQ, a.valuesEndID()...)
	if err != nil {
		return pkgErr("error executing update query", err)
	}

	return nil
}

type accountValFunc func(*Account) error

func runAccountValFuncs(a *Account, fns ...accountValFunc) error {
	if a == nil {
		return fmt.Errorf("account is nil")
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
