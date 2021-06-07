package repository

import (
	"forum/pkg/forum/repository"
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

func (p *Postgres) ProcedureRequests() error {
	//forum
	if _, err := p.DB.Prepare("SelectUsersByForumDesc", repository.SelectUsersByForumDesc); err != nil {
		return err
	}
	if _, err := p.DB.Prepare("SelectUsersByForumSince", repository.SelectUsersByForumSince); err != nil {
		return err
	}
	if _, err := p.DB.Prepare("SelectUsersByForumSinceDesc", repository.SelectUsersByForumSinceDesc); err != nil {
		return err
	}
	if _, err := p.DB.Prepare("SelectForum", repository.SelectForum); err != nil {
		return err
	}
	if _, err := p.DB.Prepare("SelectUsersByForum", repository.SelectUsersByForum); err != nil {
		return err
	}
	if _, err := p.DB.Prepare("InsertForum", repository.InsertForum); err != nil {
		return err
	}

	return nil
}
