package repository

import (
	"forum/pkg/models"
	"github.com/jmoiron/sqlx"
	"log"
)

type ServiceRepositoryInterface interface {
	CleanDB() bool
}

type ServiceRepository struct {
	Status *models.Status
	DB     *sqlx.DB
}

func (r ServiceRepository) CleanDb() bool {
	_, err := r.DB.Exec(`TRUNCATE parkmaildb."Thread", parkmaildb."Forum", parkmaildb."User", parkmaildb."Vote", parkmaildb."Post", parkmaildb."Users_by_Forum"`)
	if err != nil {
		log.Println(err)
		return false
	}

	return true
}

func (r ServiceRepository) GetStatus() models.Status {
	status := models.Status{}

	err := r.DB.QueryRowx(`SELECT COUNT(*) FROM parkmaildb."Post"`).Scan(&status.Post)
	err = r.DB.QueryRowx(`SELECT COUNT(*) FROM parkmaildb."User"`).Scan(&status.User)
	err = r.DB.QueryRowx(`SELECT COUNT(*) FROM parkmaildb."Forum"`).Scan(&status.Forum)
	err = r.DB.QueryRowx(`SELECT COUNT(*) FROM parkmaildb."Thread"`).Scan(&status.Thread)
	if err != nil {
		log.Println(err)
	}
	return status
}
