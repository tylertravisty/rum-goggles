package models

import (
	"database/sql"
	"fmt"
	"strings"
)

const (
	channelColumns = "id, account_id, cid, name, profile_image, api_key"
	channelTable   = "channel"
)

type Channel struct {
	ID           *int64  `json:"id"`
	AccountID    *int64  `json:"account_id"`
	CID          *string `json:"cid"`
	Name         *string `json:"name"`
	ProfileImage *string `json:"profile_image"`
	ApiKey       *string `json:"api_key"`
}

func (c *Channel) Id() *int64 {
	return c.ID
}

func (c *Channel) KeyUrl() *string {
	return c.ApiKey
}

func (c *Channel) LoggedIn() bool {
	return false
}

func (c *Channel) String() *string {
	if c.Name == nil {
		return nil
	}

	s := "/c/" + strings.ReplaceAll(*c.Name, " ", "")
	return &s
}

func (c *Channel) Title() *string {
	return c.Name
}

func (c *Channel) Type() string {
	return "Channel"
}

func (c *Channel) values() []any {
	return []any{c.ID, c.AccountID, c.CID, c.Name, c.ProfileImage, c.ApiKey}
}

func (c *Channel) valuesNoID() []any {
	return c.values()[1:]
}

func (c *Channel) valuesEndID() []any {
	vals := c.values()
	return append(vals[1:], vals[0])
}

type sqlChannel struct {
	id           sql.NullInt64
	accountID    sql.NullInt64
	cid          sql.NullString
	name         sql.NullString
	profileImage sql.NullString
	apiKey       sql.NullString
}

func (sc *sqlChannel) scan(r Row) error {
	return r.Scan(&sc.id, &sc.accountID, &sc.cid, &sc.name, &sc.profileImage, &sc.apiKey)
}

func (sc sqlChannel) toChannel() *Channel {
	var c Channel
	c.ID = toInt64(sc.id)
	c.AccountID = toInt64(sc.accountID)
	c.CID = toString(sc.cid)
	c.Name = toString(sc.name)
	c.ProfileImage = toString(sc.profileImage)
	c.ApiKey = toString(sc.apiKey)

	return &c
}

type ChannelService interface {
	AutoMigrate() error
	ByAccount(a *Account) ([]Channel, error)
	ByID(id int64) (*Channel, error)
	ByName(name string) (*Channel, error)
	Create(c *Channel) error
	Delete(c *Channel) error
	DestructiveReset() error
	Update(c *Channel) error
}

func NewChannelService(db *sql.DB) ChannelService {
	return &channelService{
		Database: db,
	}
}

var _ ChannelService = &channelService{}

type channelService struct {
	Database *sql.DB
}

func (cs *channelService) AutoMigrate() error {
	err := cs.createChannelTable()
	if err != nil {
		return pkgErr(fmt.Sprintf("error creating %s table", channelTable), err)
	}

	return nil
}

func (cs *channelService) createChannelTable() error {
	createQ := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS "%s" (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			account_id INTEGER NOT NULL,
			cid TEXT UNIQUE NOT NULL,
			name TEXT UNIQUE NOT NULL,
			profile_image TEXT,
			api_key TEXT NOT NULL,
			FOREIGN KEY (account_id) REFERENCES "%s" (id)
		)
	`, channelTable, accountTable)

	_, err := cs.Database.Exec(createQ)
	if err != nil {
		return fmt.Errorf("error executing create query: %v", err)
	}

	return nil
}

func (cs *channelService) ByAccount(a *Account) ([]Channel, error) {
	err := runAccountValFuncs(
		a,
		accountRequireID,
	)
	if err != nil {
		return nil, pkgErr("", err)
	}

	selectQ := fmt.Sprintf(`
		SELECT %s
		FROM "%s"
		WHERE account_id=?
	`, channelColumns, channelTable)

	rows, err := cs.Database.Query(selectQ, a.ID)
	if err != nil {
		return nil, pkgErr("error executing select query", err)
	}
	defer rows.Close()

	channels := []Channel{}
	for rows.Next() {
		sc := &sqlChannel{}

		err = sc.scan(rows)
		if err != nil {
			return nil, pkgErr("error scanning row", err)
		}

		channels = append(channels, *sc.toChannel())
	}
	err = rows.Err()
	if err != nil && err != sql.ErrNoRows {
		return nil, pkgErr("error iterating over rows", err)
	}

	return channels, nil
}

func (cs *channelService) ByID(id int64) (*Channel, error) {
	err := runChannelValFuncs(
		&Channel{ID: &id},
		channelRequireID,
	)
	if err != nil {
		return nil, pkgErr("", err)
	}

	selectQ := fmt.Sprintf(`
		SELECT %s
		FROM "%s"
		WHERE id=?
	`, channelColumns, channelTable)

	var sc sqlChannel
	row := cs.Database.QueryRow(selectQ, id)
	err = sc.scan(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, pkgErr("error executing select query", err)
	}

	return sc.toChannel(), nil
}

func (cs *channelService) ByName(name string) (*Channel, error) {
	err := runChannelValFuncs(
		&Channel{Name: &name},
		channelRequireName,
	)
	if err != nil {
		return nil, pkgErr("", err)
	}

	selectQ := fmt.Sprintf(`
		SELECT %s
		FROM "%s"
		WHERE name=?
	`, channelColumns, channelTable)

	var sc sqlChannel
	row := cs.Database.QueryRow(selectQ, name)
	err = sc.scan(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, pkgErr("error executing select query", err)
	}

	return sc.toChannel(), nil
}

func (cs *channelService) Create(c *Channel) error {
	err := runChannelValFuncs(
		c,
		channelRequireAccountID,
		channelRequireApiKey,
		channelRequireCID,
		channelRequireName,
	)
	if err != nil {
		return pkgErr("invalid channel", err)
	}

	columns := columnsNoID(channelColumns)
	insertQ := fmt.Sprintf(`
		INSERT INTO "%s" (%s)
		VALUES (%s)
	`, channelTable, columns, values(columns))

	_, err = cs.Database.Exec(insertQ, c.valuesNoID()...)
	if err != nil {
		return pkgErr("error executing insert query", err)
	}

	return nil
}

func (cs *channelService) Delete(c *Channel) error {
	err := runChannelValFuncs(
		c,
		channelRequireID,
	)
	if err != nil {
		return pkgErr("invalid channel", err)
	}

	deleteQ := fmt.Sprintf(`
		DELETE FROM "%s"
		WHERE id=?
	`, channelTable)

	_, err = cs.Database.Exec(deleteQ, c.ID)
	if err != nil {
		return pkgErr("error executing delete query", err)
	}

	return nil
}

func (cs *channelService) DestructiveReset() error {
	err := cs.dropChannelTable()
	if err != nil {
		return pkgErr(fmt.Sprintf("error dropping %s table", channelTable), err)
	}

	return nil
}

func (cs *channelService) dropChannelTable() error {
	dropQ := fmt.Sprintf(`
		DROP TABLE IF EXISTS "%s"
	`, channelTable)

	_, err := cs.Database.Exec(dropQ)
	if err != nil {
		return fmt.Errorf("error executing drop query: %v", err)
	}

	return nil
}

func (cs *channelService) Update(c *Channel) error {
	err := runChannelValFuncs(
		c,
		channelRequireAccountID,
		channelRequireApiKey,
		channelRequireCID,
		channelRequireName,
	)
	if err != nil {
		return pkgErr("invalid channel", err)
	}

	columns := columnsNoID(channelColumns)
	updateQ := fmt.Sprintf(`
		UPDATE "%s"
		SET %s
		WHERE id=?
	`, channelTable, set(columns))

	_, err = cs.Database.Exec(updateQ, c.valuesEndID()...)
	if err != nil {
		return pkgErr("error executing update query", err)
	}

	return nil
}

type channelValFunc func(*Channel) error

func runChannelValFuncs(c *Channel, fns ...channelValFunc) error {
	if c == nil {
		return fmt.Errorf("channel is nil")
	}

	for _, fn := range fns {
		err := fn(c)
		if err != nil {
			return err
		}
	}

	return nil
}

func channelRequireAccountID(c *Channel) error {
	if c.AccountID == nil || *c.AccountID <= 0 {
		return ErrChannelInvalidAccountID
	}

	return nil
}

func channelRequireApiKey(c *Channel) error {
	if c.ApiKey == nil || *c.ApiKey == "" {
		return ErrChannelInvalidApiKey
	}

	return nil
}

func channelRequireID(c *Channel) error {
	if c.ID == nil || *c.ID < 1 {
		return ErrChannelInvalidID
	}

	return nil
}

func channelRequireCID(c *Channel) error {
	if c.CID == nil || *c.CID == "" {
		return ErrChannelInvalidCID
	}

	return nil
}

func channelRequireName(c *Channel) error {
	if c.Name == nil || *c.Name == "" {
		return ErrChannelInvalidName
	}

	return nil
}
