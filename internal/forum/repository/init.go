package repository

import (
	"database/sql"
	_ "github.com/jackc/pgx"
	"github.com/jmoiron/sqlx"
)

func getPostgres() *sql.DB {
	db, err := sql.Open("pgx", "dbname=docker user=docker password=docker host=127.0.0.1 port=5432 sslmode=disable")
	//db, err := sql.Open("pgx", "dbname=postgres host=127.0.0.1 port=5432 sslmode=disable")
	if err != nil {
		panic("cant parse config" + err.Error())
	}

	err = db.Ping()
	if err != nil {
		panic("can`t ping db" + err.Error())
	}

	db.SetMaxOpenConns(100)

	return db
}

func Init() *sqlx.DB {
	return sqlx.NewDb(getPostgres(), "psx")
}
