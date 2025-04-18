package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

type Postgres struct {
	db  *sqlx.DB
	url string
}

func NewPostgres(url string) (*Postgres, error) {
	db, err := sqlx.Open("postgres", url)
	if err != nil {
		return nil, err
	}

	return &Postgres{
		db:  db,
		url: url,
	}, nil
}

func (d *Postgres) CreateDB(dbname string) error {
	_, err := d.db.Exec("create database " + dbname)
	// TODO: add error handling when already exist
	fmt.Println(err)
	return nil
}

func (d *Postgres) MigrateUp(databaseDir string) error {
	m, err := migrate.New(fmt.Sprintf("file://%s", databaseDir), d.url)
	if err != nil {
		return err
	}
	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

func (d *Postgres) MigrateDown(databaseDir string) error {
	m, err := migrate.New(fmt.Sprintf("file://%s", databaseDir), d.url)
	if err != nil {
		return err
	}
	if err = m.Down(); err != nil {
		return err
	}
	return err
}

func (d *Postgres) Ping() error {
	return d.db.Ping()
}

func (d *Postgres) Exec(query string, args ...interface{}) error {
	_, err := d.db.Exec(query, args...)
	if err != nil {
		return err
	}
	return nil
}

func (d *Postgres) NameExec(query string, args interface{}) error {
	_, err := d.db.NamedExec(query, args)
	if err != nil {
		return err
	}
	return nil
}

// Multi Select
func (d *Postgres) Query(query string, args ...interface{}) (*sql.Rows, error) {
	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s", err.Error())
	}
	return rows, nil
}

// Single Select
func (d *Postgres) QueryRow(query string, args ...interface{}) *sql.Row {
	row := d.db.QueryRow(query, args...)
	return row
}

func (d *Postgres) Begin() (*Tx, error) {
	tx, err := d.db.Begin()
	if err != nil {
		return nil, err
	}
	return &Tx{tx: tx}, nil
}
