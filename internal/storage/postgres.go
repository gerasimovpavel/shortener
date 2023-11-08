package storage

import (
	"context"
	"errors"
	"fmt"
	urlgen "github.com/gerasimovpavel/shortener.git/internal/urlgenerator"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"strconv"
	"strings"
)

type PgWorker struct {
	conn *pgx.Conn
	tx   pgx.Tx
}

func NewPgStorage(ps string) (*PgWorker, error) {
	conn, err := pgx.Connect(context.Background(), ps)
	if err != nil {
		return nil, err
	}
	_, err = conn.Exec(context.Background(),
		`	CREATE TABLE IF NOT EXISTS urls
				(
					uuid character(3) COLLATE pg_catalog."default",
					"shortURL" character(10) COLLATE pg_catalog."default",
					"originalURL" character(1000) COLLATE pg_catalog."default",
				 CONSTRAINT "urls_originalURL_key" UNIQUE ("originalURL")
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
	data.ShortURL = strings.Trim(data.ShortURL, " ")
	return data, nil
}

func (pgw *PgWorker) FindByOriginalURL(originalURL string) (*URLData, error) {
	data := &URLData{}
	err := pgw.conn.QueryRow(context.Background(), `SELECT uuid, "shortURL", "originalURL" FROM public.urls where "originalURL"=$1`, originalURL).Scan(&data.UUID, &data.ShortURL, &data.OriginalURL)
	if err != nil && err != pgx.ErrNoRows {
		return data, err
	}
	data.ShortURL = strings.Trim(data.ShortURL, " ")
	return data, nil
}

func (pgw *PgWorker) PostBatch(urls []*URLData) error {
	var err, errConf error
	ctx := context.Background()

	pgw.tx, err = pgw.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("ошибка tx create: %v", err)
	}

	for _, data := range urls {
		err = pgw.Post(data)
		if err != nil && !errors.Is(err, ErrDataConflict) {
			err2 := pgw.tx.Rollback(ctx)
			if err2 != nil {
				return fmt.Errorf("ошибка rollback: %v", err2)
			}
			return err
		}
		if errors.Is(err, ErrDataConflict) {
			errConf = errors.Join(errConf, err)
		}

	}
	err = pgw.tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("ошибка commit: %v", err)
	}
	return errors.Join(nil, errConf)
}

func (pgw *PgWorker) Post(data *URLData) error {
	var errConf error
	if data.ShortURL == "" {
		data.ShortURL = urlgen.GenShort()
	}

	uuid, err := pgw.rowsCount()
	if err != nil {
		return err
	}
	data.UUID = strconv.Itoa(uuid + 1)
	if pgw.tx != nil {
		_, err = pgw.tx.Exec(context.Background(), `INSERT INTO public.urls (uuid, "shortURL", "originalURL") VALUES ($1,$2,$3)`, data.UUID, data.ShortURL, data.OriginalURL)
	} else {
		_, err = pgw.conn.Exec(context.Background(), `INSERT INTO public.urls (uuid, "shortURL", "originalURL") VALUES ($1,$2,$3)`, data.UUID, data.ShortURL, data.OriginalURL)
	}

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code != pgerrcode.UniqueViolation {
				return err
			}
			if pgErr.Code == pgerrcode.UniqueViolation {
				data2, err := pgw.FindByOriginalURL(data.OriginalURL)
				if err != nil {
					return err
				}
				data.UUID = data2.UUID
				data.OriginalURL = data2.OriginalURL
				data.ShortURL = data2.ShortURL
				data.CorrID = data2.CorrID
				errConf = errors.Join(errConf, ErrDataConflict)
			}
		}

	}
	return errors.Join(nil, errConf)
}

func (pgw *PgWorker) Ping() error {
	return pgw.conn.Ping(context.Background())
}

func (pgw *PgWorker) Close() error {
	return pgw.conn.Close(context.Background())
}
