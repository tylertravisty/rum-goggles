package models

import (
	"database/sql"
	"fmt"
)

const (
	chatbotColumns = "id, name, url"
	chatbotTable   = "chatbot"
)

type Chatbot struct {
	ID   *int64  `json:"id"`
	Name *string `json:"name"`
	Url  *string `json:"url"`
}

func (c *Chatbot) values() []any {
	return []any{c.ID, c.Name, c.Url}
}

func (c *Chatbot) valuesNoID() []any {
	return c.values()[1:]
}

func (c *Chatbot) valuesEndID() []any {
	vals := c.values()
	return append(vals[1:], vals[0])
}

type sqlChatbot struct {
	id   sql.NullInt64
	name sql.NullString
	url  sql.NullString
}

func (sc *sqlChatbot) scan(r Row) error {
	return r.Scan(&sc.id, &sc.name, &sc.url)
}

func (sc sqlChatbot) toChatbot() *Chatbot {
	var c Chatbot
	c.ID = toInt64(sc.id)
	c.Name = toString(sc.name)
	c.Url = toString(sc.url)

	return &c
}

type ChatbotService interface {
	All() ([]Chatbot, error)
	AutoMigrate() error
	ByID(id int64) (*Chatbot, error)
	ByName(name string) (*Chatbot, error)
	Create(c *Chatbot) (int64, error)
	Delete(c *Chatbot) error
	DestructiveReset() error
	Update(c *Chatbot) error
}

func NewChatbotService(db *sql.DB) ChatbotService {
	return &chatbotService{
		Database: db,
	}
}

var _ ChatbotService = &chatbotService{}

type chatbotService struct {
	Database *sql.DB
}

func (cs *chatbotService) All() ([]Chatbot, error) {
	selectQ := fmt.Sprintf(`
		SELECT %s
		FROM "%s"
	`, chatbotColumns, chatbotTable)

	rows, err := cs.Database.Query(selectQ)
	if err != nil {
		return nil, pkgErr("error executing select query", err)
	}
	defer rows.Close()

	chatbots := []Chatbot{}
	for rows.Next() {
		sc := &sqlChatbot{}

		err = sc.scan(rows)
		if err != nil {
			return nil, pkgErr("error scanning row", err)
		}

		chatbots = append(chatbots, *sc.toChatbot())
	}
	err = rows.Err()
	if err != nil && err != sql.ErrNoRows {
		return nil, pkgErr("error iterating over rows", err)
	}

	return chatbots, nil
}

func (cs *chatbotService) AutoMigrate() error {
	err := cs.createChatbotTable()
	if err != nil {
		return pkgErr(fmt.Sprintf("error creating %s table", chatbotTable), err)
	}

	return nil
}

func (cs *chatbotService) createChatbotTable() error {
	createQ := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS "%s" (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			url TEXT
		)
	`, chatbotTable)

	_, err := cs.Database.Exec(createQ)
	if err != nil {
		return fmt.Errorf("error executing create query: %v", err)
	}

	return nil
}

func (cs *chatbotService) ByID(id int64) (*Chatbot, error) {
	err := runChatbotValFuncs(
		&Chatbot{ID: &id},
		chatbotRequireID,
	)
	if err != nil {
		return nil, pkgErr("", err)
	}

	selectQ := fmt.Sprintf(`
		SELECT %s
		FROM "%s"
		WHERE id=?
	`, chatbotColumns, chatbotTable)

	var sc sqlChatbot
	row := cs.Database.QueryRow(selectQ, id)
	err = sc.scan(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, pkgErr("error executing select query", err)
	}

	return sc.toChatbot(), nil
}

func (cs *chatbotService) ByName(name string) (*Chatbot, error) {
	err := runChatbotValFuncs(
		&Chatbot{Name: &name},
		chatbotRequireName,
	)
	if err != nil {
		return nil, pkgErr("", err)
	}

	selectQ := fmt.Sprintf(`
		SELECT %s
		FROM "%s"
		WHERE name=?
	`, chatbotColumns, chatbotTable)

	var sc sqlChatbot
	row := cs.Database.QueryRow(selectQ, name)
	err = sc.scan(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, pkgErr("error executing select query", err)
	}

	return sc.toChatbot(), nil
}

func (cs *chatbotService) Create(c *Chatbot) (int64, error) {
	err := runChatbotValFuncs(
		c,
		chatbotRequireName,
	)
	if err != nil {
		return -1, pkgErr("invalid chatbot", err)
	}

	columns := columnsNoID(chatbotColumns)
	insertQ := fmt.Sprintf(`
		INSERT INTO "%s" (%s)
		VALUES (%s)
		RETURNING id
	`, chatbotTable, columns, values(columns))

	var id int64
	row := cs.Database.QueryRow(insertQ, c.valuesNoID()...)
	err = row.Scan(&id)
	if err != nil {
		return -1, pkgErr("error executing insert query", err)
	}

	return id, nil
}

func (cs *chatbotService) Delete(c *Chatbot) error {
	err := runChatbotValFuncs(
		c,
		chatbotRequireID,
	)
	if err != nil {
		return pkgErr("invalid chatbot", err)
	}

	deleteQ := fmt.Sprintf(`
		DELETE FROM "%s"
		WHERE id=?
	`, chatbotTable)

	_, err = cs.Database.Exec(deleteQ, c.ID)
	if err != nil {
		return pkgErr("error executing delete query", err)
	}

	return nil
}

func (cs *chatbotService) DestructiveReset() error {
	err := cs.dropChatbotTable()
	if err != nil {
		return pkgErr(fmt.Sprintf("error dropping %s table", chatbotTable), err)
	}

	return nil
}

func (cs *chatbotService) dropChatbotTable() error {
	dropQ := fmt.Sprintf(`
		DROP TABLE IF EXISTS "%s"
	`, chatbotTable)

	_, err := cs.Database.Exec(dropQ)
	if err != nil {
		return fmt.Errorf("error executing drop query: %v", err)
	}

	return nil
}

func (cs *chatbotService) Update(c *Chatbot) error {
	err := runChatbotValFuncs(
		c,
		chatbotRequireID,
		chatbotRequireName,
	)
	if err != nil {
		return pkgErr("invalid chatbot", err)
	}

	columns := columnsNoID(chatbotColumns)
	updateQ := fmt.Sprintf(`
		UPDATE "%s"
		SET %s
		WHERE id=?
	`, chatbotTable, set(columns))

	_, err = cs.Database.Exec(updateQ, c.valuesEndID()...)
	if err != nil {
		return pkgErr("error executing update query", err)
	}

	return nil
}

type chatbotValFunc func(*Chatbot) error

func runChatbotValFuncs(c *Chatbot, fns ...chatbotValFunc) error {
	if c == nil {
		return fmt.Errorf("chatbot is nil")
	}

	for _, fn := range fns {
		err := fn(c)
		if err != nil {
			return err
		}
	}

	return nil
}

func chatbotRequireID(c *Chatbot) error {
	if c.ID == nil || *c.ID < 1 {
		return ErrChatbotInvalidID
	}

	return nil
}

func chatbotRequireName(c *Chatbot) error {
	if c.Name == nil || *c.Name == "" {
		return ErrChatbotInvalidName
	}

	return nil
}
