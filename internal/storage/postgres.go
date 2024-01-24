package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/georgysavva/scany/v2/pgxscan"
	urlgen "github.com/gerasimovpavel/shortener.git/internal/urlgenerator"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"strconv"
	"strings"
)

type PgWorker struct {
	//conn *pgx.Conn
	//tx   pgx.Tx
	pool *pgxpool.Pool
}

func NewPostgreWorker(ps string) (*PgWorker, error) {
	config, err := pgxpool.ParseConfig(ps)
	if err != nil {
		return nil, err
	}
	config.MaxConns = 50
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	//conn, err := pgx.Connect(context.Background(), ps)
	if err != nil {
		return nil, err
	}

	_, err = pool.Exec(context.Background(),
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
	return &PgWorker{pool: pool}, nil
}

func (pgw *PgWorker) rowsCount() (int, error) {
	var cnt int
	err := pgw.pool.QueryRow(context.Background(), `SELECT COUNT(uuid) FROM public.urls`).Scan(&cnt)
	if err != nil && err != pgx.ErrNoRows {
		return -1, err
	}
	return cnt, nil

}

func (pgw *PgWorker) Get(shortURL string) (*URLData, error) {
	urls := []URLData{}
	data := &URLData{}
	err := pgxscan.Select(context.Background(), pgw.pool, &urls, `SELECT uuid, "originalURL", "shortURL", is_deleted, "userID" FROM public.urls WHERE "shortURL"=$1`, shortURL)
	if err != nil && err != pgx.ErrNoRows {
		return data, err
	}
	if len(urls) > 0 {
		data = &urls[0]
	}
	data.ShortURL = strings.Trim(data.ShortURL, " ")
	return data, nil
}

func (pgw *PgWorker) FindByOriginalURL(originalURL string) (*URLData, error) {
	data := URLData{}
	row := pgw.pool.QueryRow(context.Background(), `SELECT uuid, "shortURL", "originalURL"FROM urls where "originalURL"=$1`, originalURL)

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

	tx, err := pgw.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("ошибка tx create: %v", err)
	}

	for _, data := range urls {
		err = pgw.Post(data)
		if err != nil && !errors.Is(err, ErrDataConflict) {
			err2 := tx.Rollback(ctx)
			if err2 != nil {
				return fmt.Errorf("ошибка rollback: %v", err2)
			}
			return err
		}
		if errors.Is(err, ErrDataConflict) {
			errConf = errors.Join(errConf, err)
		}

	}
	err = tx.Commit(ctx)

	if err != nil {
		return fmt.Errorf("ошибка commit: %v", err)
	}
	return errors.Join(nil, errConf)
}

func (pgw *PgWorker) Post(data *URLData) error {
	var errConf error
	if data.ShortURL == "" {
		data.ShortURL = urlgen.GenShort_Optimized()
	}

	uuid, err := pgw.rowsCount()
	if err != nil {
		return err
	}
	data.UUID = strconv.Itoa(uuid + 1)

	//_, err = pgw.Exec(context.Background(), `INSERT INTO urls (uuid, "shortURL", "originalURL") VALUES ($1,$2,$3)`, data.UUID, data.ShortURL, data.OriginalURL)

	err = pgw.pool.QueryRow(context.Background(),
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
	return pgw.pool.Ping(context.Background())
}

func (pgw *PgWorker) Close() error {
	pgw.pool.Close()
	return nil
}

func (pgw *PgWorker) GetUserURL(userID string) ([]*URLData, error) {
	urls := []*URLData{}
	err := pgxscan.Select(context.Background(), pgw.pool, &urls, `SELECT "originalURL", "shortURL" FROM urls WHERE "userID"=$1`, userID)
	if err != nil {
		return urls, err
	}
	return urls, nil
}

func (pgw *PgWorker) DeleteUserURL(urls []*URLData) error {

	valueStrings := make([]string, 0, len(urls))
	valueArgs := make([]interface{}, 0, len(urls)*2)
	i := 0
	for _, url := range urls {
		valueStrings = append(valueStrings, fmt.Sprintf(`($%d, $%d)`, i*2+1, i*2+2))
		valueArgs = append(valueArgs, url.ShortURL)
		valueArgs = append(valueArgs, url.UserID)
		i++
	}

	stmt := fmt.Sprintf(`
					UPDATE urls AS u SET is_deleted=true  
					FROM (VALUES
							%s
						 ) AS x ("shortURL", "userID")
					WHERE x."shortURL"=u."shortURL" AND x."userID"=u."userID"`,
		strings.Join(valueStrings, ","))

	_, err := pgw.pool.Exec(context.Background(), stmt, valueArgs...)
	return err
}
