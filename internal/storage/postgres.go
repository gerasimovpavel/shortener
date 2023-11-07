package storage

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"strconv"
)

type PgWorker struct {
	conn *pgx.Conn
}

func NewPgStorage(ps string) (*PgWorker, error) {
	//ps := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
	//	`localhost`, `shortener`, `shortener`, `shortener`)
	conn, err := pgx.Connect(context.Background(), ps)
	if err != nil {
		return nil, err
	}
	return &PgWorker{conn: conn}, nil
}

func (pgw *PgWorker) rowsCount() (int, error) {
	var cnt int
	err := pgw.conn.QueryRow(context.Background(), `SELECT COUNT(uuid) FROM urls`).Scan(&cnt)
	if err != nil {
		return -1, err
	}
	return cnt, nil
}

func (pgw *PgWorker) Get(shortURL string) (*URLData, error) {
	data := &URLData{}
	err := pgw.conn.QueryRow(context.Background(), `SELECT uuid, originalURL, shortURL FROM urls WHERE shortURL=?`, shortURL).Scan(&data.UUID, &data.OriginalURL, &data.ShortURL)
	if err != nil {
		return data, err
	}
	if data.ShortURL == "" {
		return data, errors.New("ничего не найдено")
	}
	return data, nil
}

func (pgw *PgWorker) FindByOriginalURL(originalURL string) (*URLData, error) {
	data := &URLData{}
	err := pgw.conn.QueryRow(context.Background(), `SELECT uuid, originalURL, shortURL FROM urls WHERE originalURL=?`, originalURL).Scan(&data.UUID, &data.OriginalURL, &data.ShortURL)
	if err != nil {
		return data, err
	}
	if data.ShortURL == "" {
		return data, errors.New("ничего не найдено")
	}
	return data, nil
}
func (pgw *PgWorker) Post(data *URLData) error {
	item, err := pgw.FindByOriginalURL(data.OriginalURL)
	if err != nil {
		return err
	}
	if item.ShortURL != "" {
		return errors.New("ссылка уже существует")
	}
	item, err = pgw.Get(data.ShortURL)
	if err != nil {
		return err
	}
	if item.ShortURL != "" {
		return errors.New("ссылка уже существует")
	}
	uuid, err := pgw.rowsCount()
	if err != nil {
		return err
	}
	data.UUID = strconv.Itoa(uuid + 1)
	_, err = pgw.conn.Exec(context.Background(), `INSERT INTO urls (uuid, shortURL, originalURL) VALUES (?,?,?)`, data.UUID, data.ShortURL, data.OriginalURL)
	if err != nil {
		return err
	}
	return nil
}

func (pgw *PgWorker) Ping() error {
	return pgw.conn.Ping(context.Background())
}

func (pgw *PgWorker) Close() error {
	return pgw.conn.Close(context.Background())
}
