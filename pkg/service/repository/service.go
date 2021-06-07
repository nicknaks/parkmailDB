package repository

import (
	"forum/pkg/models"
	"github.com/jackc/pgx"
	"log"
)

const (
	CleanDB      = `TRUNCATE parkmaildb."Thread", parkmaildb."Forum", parkmaildb."User", parkmaildb."Vote", parkmaildb."Post", parkmaildb."Users_by_Forum"`
	StatusPost   = `SELECT COUNT(*) FROM parkmaildb."Post"`
	StatusUser   = `SELECT COUNT(*) FROM parkmaildb."User"`
	StatusForum  = `SELECT COUNT(*) FROM parkmaildb."Forum"`
	StatusThread = `SELECT COUNT(*) FROM parkmaildb."Thread"`
)

type ServiceRepositoryInterface interface {
	CleanDB() bool
}

type ServiceRepository struct {
	Status *models.Status
	DB     *pgx.ConnPool
}

func (r ServiceRepository) CleanDb() bool {
	_, err := r.DB.Exec("CleanDB")
	if err != nil {
		log.Println(err)
		return false
	}

	return true
}

func (r ServiceRepository) GetStatus() models.Status {
	status := models.Status{}

	err := r.DB.QueryRow("StatusPost").Scan(&status.Post)
	err = r.DB.QueryRow("StatusUser").Scan(&status.User)
	err = r.DB.QueryRow("StatusForum").Scan(&status.Forum)
	err = r.DB.QueryRow("StatusThread").Scan(&status.Thread)
	if err != nil {
		log.Println(err)
	}
	return status
}
