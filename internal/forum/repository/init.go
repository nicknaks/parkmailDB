package repository

import (
	"github.com/jackc/pgx"
)

type Postgres struct {
	DB *pgx.ConnPool
}

func NewPostgres() (*Postgres, error) {
	conf := pgx.ConnConfig{
		User:                 "docker",
		Database:             "docker",
		Password:             "docker",
		PreferSimpleProtocol: false,
	}

	poolConf := pgx.ConnPoolConfig{
		ConnConfig:     conf,
		MaxConnections: 100,
		AfterConnect:   nil,
		AcquireTimeout: 0,
	}
	db, err := pgx.NewConnPool(poolConf)
	if err != nil {
		return nil, err
	}
	return &Postgres{
		DB: db,
	}, nil
}

func (p *Postgres) GetPostgres() *pgx.ConnPool {
	return p.DB
}

func (p *Postgres) Close() error {
	p.DB.Close()
	return nil
}
