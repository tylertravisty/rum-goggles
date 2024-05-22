package models

import (
	"database/sql"
	"fmt"
)

const (
	chatbotRuleColumns = "id, chatbot_id, parameters"
	chatbotRuleTable   = "chatbot_rule"
)

type ChatbotRule struct {
	ID         *int64  `json:"id"`
	ChatbotID  *int64  `json:"chatbot_id"`
	Parameters *string `json:"parameters"`
}

func (c *ChatbotRule) values() []any {
	return []any{c.ID, c.ChatbotID, c.Parameters}
}

func (c *ChatbotRule) valuesNoID() []any {
	return c.values()[1:]
}

func (c *ChatbotRule) valuesEndID() []any {
	vals := c.values()
	return append(vals[1:], vals[0])
}

type sqlChatbotRule struct {
	id         sql.NullInt64
	chatbotID  sql.NullInt64
	parameters sql.NullString
}

func (sc *sqlChatbotRule) scan(r Row) error {
	return r.Scan(&sc.id, &sc.chatbotID, &sc.parameters)
}

func (sc sqlChatbotRule) toChatbotRule() *ChatbotRule {
	var c ChatbotRule
	c.ID = toInt64(sc.id)
	c.ChatbotID = toInt64(sc.chatbotID)
	c.Parameters = toString(sc.parameters)

	return &c
}

type ChatbotRuleService interface {
	AutoMigrate() error
	ByChatbotID(cid int64) ([]ChatbotRule, error)
	Create(c *ChatbotRule) (int64, error)
	Delete(c *ChatbotRule) error
	DestructiveReset() error
	Update(c *ChatbotRule) error
}

func NewChatbotRuleService(db *sql.DB) ChatbotRuleService {
	return &chatbotRuleService{
		Database: db,
	}
}

var _ ChatbotRuleService = &chatbotRuleService{}

type chatbotRuleService struct {
	Database *sql.DB
}

func (cs *chatbotRuleService) AutoMigrate() error {
	err := cs.createChatbotRuleTable()
	if err != nil {
		return pkgErr(fmt.Sprintf("error creating %s table", chatbotRuleTable), err)
	}

	return nil
}

func (cs *chatbotRuleService) createChatbotRuleTable() error {
	createQ := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS "%s" (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			chatbot_id INTEGER NOT NULL,
			parameters TEXT NOT NULL,
			FOREIGN KEY (chatbot_id) REFERENCES "%s" (id)
		)
	`, chatbotRuleTable, chatbotTable)

	_, err := cs.Database.Exec(createQ)
	if err != nil {
		return fmt.Errorf("error executing create query: %v", err)
	}

	return nil
}

func (cs *chatbotRuleService) ByChatbotID(cid int64) ([]ChatbotRule, error) {
	selectQ := fmt.Sprintf(`
		SELECT %s
		FROM "%s"
		WHERE chatbot_id=?
	`, chatbotRuleColumns, chatbotRuleTable)

	rows, err := cs.Database.Query(selectQ, cid)
	if err != nil {
		return nil, pkgErr("error executing select query", err)
	}
	defer rows.Close()

	rules := []ChatbotRule{}
	for rows.Next() {
		scr := &sqlChatbotRule{}

		err = scr.scan(rows)
		if err != nil {
			return nil, pkgErr("error scanning row", err)
		}

		rules = append(rules, *scr.toChatbotRule())
	}
	err = rows.Err()
	if err != nil && err != sql.ErrNoRows {
		return nil, pkgErr("error iterating over rows", err)
	}

	return rules, nil
}

func (cs *chatbotRuleService) Create(c *ChatbotRule) (int64, error) {
	err := runChatbotRuleValFuncs(
		c,
		chatbotRuleRequireParameters,
	)
	if err != nil {
		return -1, pkgErr("invalid chat rule", err)
	}

	columns := columnsNoID(chatbotRuleColumns)
	insertQ := fmt.Sprintf(`
		INSERT INTO "%s" (%s)
		VALUES (%s)
		RETURNING id
	`, chatbotRuleTable, columns, values(columns))

	var id int64
	row := cs.Database.QueryRow(insertQ, c.valuesNoID()...)
	err = row.Scan(&id)
	if err != nil {
		return -1, pkgErr("error executing insert query", err)
	}

	return id, nil
}

func (cs *chatbotRuleService) Delete(c *ChatbotRule) error {
	err := runChatbotRuleValFuncs(
		c,
		chatbotRuleRequireID,
	)
	if err != nil {
		return pkgErr("invalid chat rule", err)
	}

	deleteQ := fmt.Sprintf(`
		DELETE FROM "%s"
		WHERE id=?
	`, chatbotRuleTable)

	_, err = cs.Database.Exec(deleteQ, c.ID)
	if err != nil {
		return pkgErr("error executing delete query", err)
	}

	return nil
}

func (cs *chatbotRuleService) DestructiveReset() error {
	err := cs.dropChatbotRuleTable()
	if err != nil {
		return pkgErr(fmt.Sprintf("error dropping %s table", chatbotRuleTable), err)
	}

	return nil
}

func (cs *chatbotRuleService) dropChatbotRuleTable() error {
	dropQ := fmt.Sprintf(`
		DROP TABLE IF EXISTS "%s"
	`, chatbotRuleTable)

	_, err := cs.Database.Exec(dropQ)
	if err != nil {
		return fmt.Errorf("error executing drop query: %v", err)
	}

	return nil
}

func (cs *chatbotRuleService) Update(c *ChatbotRule) error {
	err := runChatbotRuleValFuncs(
		c,
		chatbotRuleRequireID,
		chatbotRuleRequireParameters,
	)
	if err != nil {
		return pkgErr("invalid chat rule", err)
	}

	columns := columnsNoID(chatbotRuleColumns)
	updateQ := fmt.Sprintf(`
		UPDATE "%s"
		SET %s
		WHERE id=?
	`, chatbotRuleTable, set(columns))

	_, err = cs.Database.Exec(updateQ, c.valuesEndID()...)
	if err != nil {
		return pkgErr("error executing update query", err)
	}

	return nil
}

type chatbotRuleValFunc func(*ChatbotRule) error

func runChatbotRuleValFuncs(c *ChatbotRule, fns ...chatbotRuleValFunc) error {
	if c == nil {
		return fmt.Errorf("chat rule is nil")
	}

	for _, fn := range fns {
		err := fn(c)
		if err != nil {
			return err
		}
	}

	return nil
}

func chatbotRuleRequireID(c *ChatbotRule) error {
	if c.ID == nil || *c.ID < 1 {
		return ErrChatbotRuleInvalidID
	}

	return nil
}

func chatbotRuleRequireParameters(c *ChatbotRule) error {
	if c.Parameters == nil || *c.Parameters == "" {
		return ErrChatbotRuleInvalidParameters
	}

	return nil
}
