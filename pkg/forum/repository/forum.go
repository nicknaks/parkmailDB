package repository

import (
	"forum/pkg/models"
	"github.com/jmoiron/sqlx"
	"log"
)

type ForumRepositoryInterface interface {
	CreateForum(forum models.Forum) error
	GetForumInfo(slug string) (models.Forum, bool)
	FindUsers(slug string, params models.ParamsForSearch) ([]models.User, bool)
}

type ForumRepository struct {
	DB *sqlx.DB
}

func (r ForumRepository) FindUsers(slug string, params models.ParamsForSearch) ([]models.User, bool) {
	var rows *sqlx.Rows
	var err error

	if params.Since == "" {
		if params.Desc {
			rows, err = r.DB.Queryx(`SELECT U.nickname, U.fullname, U.about, U.email FROM parkmaildb."Users_by_Forum" users INNER JOIN parkmaildb."User" U on U.nickname = users."user" AND users.forum = $1 ORDER BY users."user" DESC LIMIT $2`,
				slug, params.Limit)
		} else {
			rows, err = r.DB.Queryx(`SELECT U.nickname, U.fullname, U.about, U.email FROM parkmaildb."Users_by_Forum" users INNER JOIN parkmaildb."User" U on U.nickname = users."user" AND users.forum = $1 ORDER BY users."user" LIMIT $2`,
				slug, params.Limit)
		}
	} else {
		if params.Desc {
			rows, err = r.DB.Queryx(`SELECT U.nickname, U.fullname, U.about, U.email FROM parkmaildb."Users_by_Forum" users INNER JOIN parkmaildb."User" U on U.nickname = users."user" AND users.forum = $1 AND U.nickname <= $2 ORDER BY users."user" DESC LIMIT $3`,
				slug, params.Since, params.Limit)
		} else {
			rows, err = r.DB.Queryx(`SELECT U.nickname, U.fullname, U.about, U.email FROM parkmaildb."Users_by_Forum" users INNER JOIN parkmaildb."User" U on U.nickname = users."user" AND users.forum = $1 AND U.nickname <= $2 ORDER BY users."user" LIMIT $3`,
				slug, params.Since, params.Limit)
		}
	}

	if err != nil {
		log.Println(err)
		return nil, false
	}

	var users []models.User

	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.Nickname, &user.Fullname, &user.About, &user.Email)
		if err != nil {
			log.Println(err)
			return nil, false
		}
		users = append(users, user)
	}

	return users, true
}

func (r ForumRepository) CreateForum(forum models.Forum) error {
	_, err := r.DB.Exec(`INSERT INTO parkmaildb."Forum" (title, "user", slug, posts, threads) VALUES ($1,$2,$3,0,0)`,
		forum.Title, forum.User, forum.Slug)
	return err
}

func (r ForumRepository) GetForumInfo(slug string) (models.Forum, bool) {
	var forum models.Forum = models.Forum{Slug: slug}
	err := r.DB.QueryRowx(`SELECT f.title, f."user", f.posts, f.threads from parkmaildb."Forum" f WHERE slug = $1`, forum.Slug).
		Scan(&forum.Title, &forum.User, &forum.Posts, &forum.Threads)

	if err != nil {
		log.Println(err)
		return models.Forum{}, false
	}

	return forum, true
}
