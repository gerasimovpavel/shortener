package storage

import (
	"context"
	"errors"
	urlgen "github.com/gerasimovpavel/shortener.git/internal/urlgenerator"
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
	_, err = conn.Exec(context.Background(),
		`	CREATE TABLE IF NOT EXISTS urls
				(
					uuid character(3) COLLATE pg_catalog."default",
					"shortURL" character(10) COLLATE pg_catalog."default",
					"originalURL" character(1000) COLLATE pg_catalog."default"
				)`,
	)
	if err != nil {
		return nil, err
	}
	return &PgWorker{conn: conn}, nil
}

func (pgw *PgWorker) rowsCount() (int, error) {
	var cnt int
	err := pgw.conn.QueryRow(context.Background(), `SELECT COUNT(uuid) FROM public.urls`).Scan(&cnt)
	if err != nil && err != pgx.ErrNoRows {
		return -1, err
	}
	return cnt, nil
}

func (pgw *PgWorker) Get(shortURL string) (*URLData, error) {
	data := &URLData{}
	err := pgw.conn.QueryRow(context.Background(), `SELECT uuid, "originalURL", "shortURL" FROM public.urls WHERE "shortURL"=$1`, shortURL).Scan(&data.UUID, &data.OriginalURL, &data.ShortURL)
	if err != nil && err != pgx.ErrNoRows {
		return data, err
	}
	return data, nil
}

func (pgw *PgWorker) FindByOriginalURL(originalURL string) (*URLData, error) {
	data := &URLData{}
	err := pgw.conn.QueryRow(context.Background(), `SELECT uuid, "shortURL", "originalURL" FROM public.urls where "originalURL"=$1`, originalURL).Scan(&data.UUID, &data.ShortURL, &data.OriginalURL)
	if err != nil && err != pgx.ErrNoRows {
		return data, err
	}
	return data, nil
}

func (pgw *PgWorker) PostBatch(data []*URLData) error {
	ctx := context.Background()

	tx, err := pgw.conn.Begin(ctx)
	if err != nil {
		return err
	}

	for _, url := range data {
		u, err := pgw.FindByOriginalURL(url.OriginalURL)
		if err != nil {
			return err
		}
		switch u.ShortURL {
		case "":
			{
				url.ShortURL = urlgen.GenShort()
				uuid, _ := pgw.rowsCount()
				url.UUID = strconv.Itoa(uuid + 1)
			}
		default:
			{
				url.UUID = u.UUID
				url.ShortURL = u.ShortURL
			}
		}

		_, err = tx.Exec(ctx, `INSERT INTO public.urls (uuid, "shortURL", "originalURL") VALUES ($1,$2,$3) ON CONFLICT ("originalURL") DO NOTHING`, url.UUID, url.ShortURL, url.OriginalURL)
		if err != nil {
			tx.Rollback(ctx)
		}
	}
	err = tx.Commit(ctx)
	if err != nil {
		return err
	}
	return nil
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
	_, err = pgw.conn.Exec(context.Background(), `INSERT INTO public.urls (uuid, "shortURL", "originalURL") VALUES ($1,$2,$3)`, data.UUID, data.ShortURL, data.OriginalURL)
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
