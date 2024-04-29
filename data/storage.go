package data

import (
	"context"
	"github.com/jackc/pgx/v5"
	"time"
)

type Storage interface {
	PressButton(userId string) (int64, error)
	GetLastPress() (*Press, error)
	GetLastPressByUser(userId string) (*Press, error)
	GetNumberOfPresses() (int64, error)
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
		    			time_pressed BIGINT NOT NULL
		)
	`)
	return err
}

func (s *PostgresStorage) Close() {
	_ = s.conn.Close(context.Background())
}

func (s *PostgresStorage) PressButton(userId string) (int64, error) {
	now := time.Now().Unix() * 1000
	_, err := s.conn.Exec(context.Background(), `
		INSERT INTO presses (user_id, time_pressed) VALUES ($1, $2)
	`, userId, now)
	return now, err
}

func (s *PostgresStorage) GetLastPress() (*Press, error) {
	row := s.conn.QueryRow(context.Background(), `
		SELECT user_id, time_pressed FROM presses ORDER BY time_pressed DESC LIMIT 1
	`)
	var userId string
	var t int64
	err := row.Scan(&userId, &t)
	if err != nil {
		return nil, err
	}
	return &Press{UserId: userId, Time: t}, nil
}

func (s *PostgresStorage) GetLastPressByUser(userId string) (*Press, error) {
	row := s.conn.QueryRow(context.Background(), `
		SELECT user_id, time_pressed FROM presses WHERE user_id = $1 ORDER BY time_pressed DESC LIMIT 1
	`, userId)
	var u string
	var t int64
	err := row.Scan(&u, &t)
	if err != nil {
		return nil, err
	}
	return &Press{UserId: u, Time: t}, nil
}

func (s *PostgresStorage) GetNumberOfPresses() (int64, error) {
	var count int64
	row := s.conn.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM presses
	`)
	err := row.Scan(&count)
	return count, err
}
