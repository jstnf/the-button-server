package data

import (
	"context"
	"github.com/jackc/pgx/v5"
)

type Storage interface {
	PressButton(userId string) error
	GetLastPress() (*Press, error)
	GetLastPressByUser(userId string) (*Press, error)
}

type PostgresStorage struct {
	conn *pgx.Conn
}

func NewPostgresStorage(url string) (*PostgresStorage, error) {
	conn, err := pgx.Connect(context.Background(), url)
	if err != nil {
		return nil, err
	}
	return &PostgresStorage{conn: conn}, nil
}

func (s *PostgresStorage) Init() error {
	_, err := s.conn.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS presses (
		    			id SERIAL PRIMARY KEY,
		    			user_id TEXT NOT NULL,
		    			time BIGINT NOT NULL
		)
	`)
	return err
}

func (s *PostgresStorage) Close() {
	_ = s.conn.Close(context.Background())
}

func (s *PostgresStorage) PressButton(userId string) error {
	// TODO
	return nil
}

func (s *PostgresStorage) GetLastPress() (*Press, error) {
	// TODO
	return nil, nil
}

func (s *PostgresStorage) GetLastPressByUser(userId string) (*Press, error) {
	// TODO
	return nil, nil
}
