package models

import (
	"database/sql"
	"fmt"
)

const (
	accountChannelColumns = "a.id, a.uid, a.username, a.cookies, a.profile_image, a.api_key, c.id, c.account_id, c.cid, c.name, c.profile_image, c.api_key"
)

type AccountChannel struct {
	Account
	Channel
}

type sqlAccountChannel struct {
	sqlAccount
	sqlChannel
}

func (sac *sqlAccountChannel) scan(r Row) error {
	return r.Scan(
		&sac.sqlAccount.id,
		&sac.sqlAccount.uid,
		&sac.sqlAccount.username,
		&sac.sqlAccount.cookies,
		&sac.sqlAccount.profileImage,
		&sac.sqlAccount.apiKey,
		&sac.sqlChannel.id,
		&sac.sqlChannel.accountID,
		&sac.sqlChannel.cid,
		&sac.sqlChannel.name,
		&sac.sqlChannel.profileImage,
		&sac.sqlChannel.apiKey,
	)
}

func (sac *sqlAccountChannel) toAccountChannel() *AccountChannel {
	var ac AccountChannel

	ac.Account = *sac.toAccount()
	ac.Channel = *sac.toChannel()

	return &ac
}

type AccountChannelService interface {
	All() ([]AccountChannel, error)
}

func NewAccountChannelService(db *sql.DB) AccountChannelService {
	return &accountChannelService{
		Database: db,
	}
}

var _ AccountChannelService = &accountChannelService{}

type accountChannelService struct {
	Database *sql.DB
}

func (as *accountChannelService) All() ([]AccountChannel, error) {
	selectQ := fmt.Sprintf(`
		SELECT %s 
		FROM "%s" a
		LEFT JOIN "%s" c ON a.id=c.account_id
	`, accountChannelColumns, accountTable, channelTable)

	rows, err := as.Database.Query(selectQ)
	if err != nil {
		return nil, pkgErr("error executing select query", err)
	}
	defer rows.Close()

	accountChannels := []AccountChannel{}
	for rows.Next() {
		sac := &sqlAccountChannel{}

		err = sac.scan(rows)
		if err != nil {
			return nil, pkgErr("error scanning row", err)
		}

		accountChannels = append(accountChannels, *sac.toAccountChannel())
	}
	err = rows.Err()
	if err != nil && err != sql.ErrNoRows {
		return nil, pkgErr("error iterating over rows", err)
	}

	return accountChannels, nil
}
