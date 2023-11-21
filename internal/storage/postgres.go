package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/georgysavva/scany/v2/pgxscan"
	urlgen "github.com/gerasimovpavel/shortener.git/internal/urlgenerator"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"strconv"
	"strings"
)

type PgWorker struct {
	conn *pgx.Conn
	tx   pgx.Tx
}

func NewPostgreWorker(ps string) (*PgWorker, error) {
	conn, err := pgx.Connect(context.Background(), ps)
	if err != nil {
		return nil, err
	}
	_, err = conn.Exec(context.Background(),
		`CREATE TABLE IF NOT EXISTS public.urls
(
    uuid text COLLATE pg_catalog."default",
    "shortURL" text COLLATE pg_catalog."default",
    "originalURL" text COLLATE pg_catalog."default",
    status text COLLATE pg_catalog."default" NOT NULL DEFAULT ''::bpchar,
    "userID" text COLLATE pg_catalog."default",
    is_deleted boolean NOT NULL DEFAULT false,
    CONSTRAINT "urls_originalURL_userID_key" UNIQUE ("originalURL", "userID")
)`,
	)
	if err != nil {
		return nil, err
	}
	return &PgWorker{conn: conn}, nil
}

func (pgw *PgWorker) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if pgw.tx != nil {
		if args == nil {
			return pgw.tx.Exec(ctx, sql)
		}
		return pgw.tx.Exec(ctx, sql, args...)
	}
	if args == nil {
		return pgw.conn.Exec(ctx, sql)
	}
	return pgw.conn.Exec(ctx, sql, args...)
}

func (pgw *PgWorker) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if pgw.tx != nil {
		if args == nil {
			return pgw.tx.Query(ctx, sql)
		}
		return pgw.tx.Query(ctx, sql, args...)
	}
	if args == nil {
		return pgw.conn.Query(ctx, sql)
	}
	return pgw.conn.Query(ctx, sql, args...)
}

func (pgw *PgWorker) Select(ctx context.Context, dst interface{}, sql string, args ...any) error {
	if pgw.tx != nil {
		return pgxscan.Select(ctx, pgw.tx, dst, sql, args...)
	}
	return pgxscan.Select(ctx, pgw.conn, dst, sql, args...)
}

func (pgw *PgWorker) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if pgw.tx != nil {
		if args == nil {
			return pgw.tx.QueryRow(ctx, sql)
		}
		return pgw.tx.QueryRow(ctx, sql, args...)
	}
	if args == nil {
		return pgw.conn.QueryRow(ctx, sql)
	}
	return pgw.conn.QueryRow(ctx, sql, args...)
}

func (pgw *PgWorker) rowsCount() (int, error) {
	var cnt int

	err := pgw.QueryRow(context.Background(), `SELECT COUNT(uuid) FROM public.urls`).Scan(&cnt)
	if err != nil && err != pgx.ErrNoRows {
		return -1, err
	}
	return cnt, nil

}

func (pgw *PgWorker) Get(shortURL string) (*URLData, error) {
	data := &URLData{}
	err := pgw.QueryRow(context.Background(), `SELECT uuid, "originalURL", "shortURL" FROM public.urls WHERE "shortURL"=$1`, shortURL).Scan(&data.UUID, &data.OriginalURL, &data.ShortURL)
	if err != nil && err != pgx.ErrNoRows {
		return data, err
	}
	data.ShortURL = strings.Trim(data.ShortURL, " ")
	return data, nil
}

func (pgw *PgWorker) FindByOriginalURL(originalURL string) (*URLData, error) {
	data := URLData{}
	row := pgw.QueryRow(context.Background(), `SELECT uuid, "shortURL", "originalURL"FROM urls where "originalURL"=$1`, originalURL)

	err := row.Scan(&data.UUID, &data.ShortURL, &data.OriginalURL, &data.UserID)
	if err != nil && err != pgx.ErrNoRows {
		return &data, err
	}
	data.ShortURL = strings.Trim(data.ShortURL, " ")
	return &data, nil
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
	pgw.tx = nil

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

	//_, err = pgw.Exec(context.Background(), `INSERT INTO urls (uuid, "shortURL", "originalURL") VALUES ($1,$2,$3)`, data.UUID, data.ShortURL, data.OriginalURL)

	err = pgw.QueryRow(context.Background(),
		`INSERT INTO urls (uuid, "shortURL", "originalURL", "userID") 
				VALUES ($1,$2,$3,$4) 
				ON CONFLICT ("originalURL","userID") DO UPDATE SET status='conflict' RETURNING "shortURL", "originalURL", status`,
		data.UUID,
		data.ShortURL,
		data.OriginalURL,
		data.UserID,
	).Scan(&data.ShortURL, &data.OriginalURL, &data.UUID)

	if err != nil {
		return err
	}
	if strings.EqualFold(strings.Trim(data.UUID, " "), "conflict") {
		errConf = errors.Join(errConf, ErrDataConflict)
	}

	return errors.Join(nil, errConf)
}

func (pgw *PgWorker) Ping() error {
	return pgw.conn.Ping(context.Background())
}

func (pgw *PgWorker) Close() error {
	return pgw.conn.Close(context.Background())
}

func (pgw *PgWorker) GetUserURL(userID string) ([]*URLData, error) {
	urls := []*URLData{}
	err := pgw.Select(context.Background(), &urls, `SELECT "originalURL", "shortURL" FROM urls WHERE "userID"=$1`, userID)
	if err != nil {
		return urls, err
	}
	return urls, nil
}

func (pgw *PgWorker) DeleteUserURL(urls []*URLData) error {
	ctx := context.Background()

	batch := &pgx.Batch{}

	for _, data := range urls {
		batch.Queue(`UPDATE urls SET is_deleted=true WHERE "userID"=$1 AND "shortURL"=$2`, data.UserID, data.ShortURL)
	}
	var br pgx.BatchResults
	var err error

	pgw.tx, err = pgw.conn.Begin(ctx)
	if err != nil {
		return err
	}

	br = pgw.tx.SendBatch(ctx, batch)

	_, err = br.Exec()

	if err != nil {
		pgw.tx.Rollback(ctx)
		br.Close()
		return err
	}
	br.Close()
	pgw.tx.Commit(ctx)
	return nil
}
