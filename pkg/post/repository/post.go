package repository

import (
	"fmt"
	"forum/pkg/models"
	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"log"
	"strings"
)

type PostRepositoryInterface interface {
	AddPosts(posts models.Posts, threadId int, forumName string) (models.Posts, error)
	ChangePost(updateMessage models.PostUpdate, id int) (models.Post, bool)
	GetAllInfo(params models.FullPostParams, id int) (models.FullPost, bool)
	GetAllPostByThread(id int, limit int, since int, desc bool) ([]models.Post, bool)
	GetPostsTree(id int, limit int, since int, desc bool) ([]models.Post, bool)
	GetPostsParentTree(id int, limit int, since int, desc bool) ([]models.Post, bool)
}

type PostRepository struct {
	DB *pgx.ConnPool
}

func (p PostRepository) ParseRowsToPost(rows *pgx.Rows) ([]models.Post, bool) {
	var posts []models.Post
	for rows.Next() {
		var post models.Post
		err := rows.Scan(&post.Id, &post.Parent, &post.Author, &post.Message, &post.IsEdited, &post.Forum, &post.Thread, &post.Created)
		if err != nil {
			log.Println(err)
			return posts, false
		}
		posts = append(posts, post)
	}

	return posts, true
}

func (p PostRepository) GetPostsTree(id int, limit int, since int, desc bool) ([]models.Post, bool) {
	var rows *pgx.Rows
	var err error

	if since == 0 {
		if desc {
			rows, err = p.DB.Query(`SELECT id, parent, author, message, isedited, forum, thread, created FROM parkmaildb."Post" WHERE thread = $1 ORDER BY path DESC, id DESC LIMIT $2`, id, limit)
		} else {
			rows, err = p.DB.Query(`SELECT id, parent, author, message, isedited, forum, thread, created FROM parkmaildb."Post" WHERE thread = $1 ORDER BY path, id LIMIT $2`, id, limit)
		}
	} else {
		if desc {
			rows, err = p.DB.Query(`SELECT id, parent, author, message, isedited, forum, thread, created FROM parkmaildb."Post" WHERE thread = $1 AND path < (SELECT path FROM parkmaildb."Post" where id = $2) ORDER BY path DESC , id DESC LIMIT $3`, id, since, limit)
		} else {
			rows, err = p.DB.Query(`SELECT id, parent, author, message, isedited, forum, thread, created FROM parkmaildb."Post" WHERE thread = $1 AND path > (SELECT path FROM parkmaildb."Post" where id = $2) ORDER BY path, id LIMIT $3`, id, since, limit)
		}
	}

	if err != nil {
		return []models.Post{}, false
	}

	return p.ParseRowsToPost(rows)
}

func (p PostRepository) GetPostsParentTree(id int, limit int, since int, desc bool) ([]models.Post, bool) {
	var rows *pgx.Rows
	var err error

	if since == 0 {
		if desc {
			rows, err = p.DB.Query(`SELECT id, parent, author, message, isedited, forum, thread, created FROM parkmaildb."Post" WHERE path[1] IN (SELECT id FROM parkmaildb."Post" WHERE thread = $1 AND parent = 0 ORDER BY id DESC LIMIT $2) ORDER BY path[1] DESC, path, id`, id, limit)
		} else {
			rows, err = p.DB.Query(`SELECT id, parent, author, message, isedited, forum, thread, created FROM parkmaildb."Post" WHERE path[1] IN (SELECT id FROM parkmaildb."Post" WHERE thread = $1 AND parent = 0 ORDER BY id LIMIT $2) ORDER BY path, id`, id, limit)
		}
	} else {
		if desc {
			rows, err = p.DB.Query(`SELECT id, parent, author, message, isedited, forum, thread, created FROM parkmaildb."Post" WHERE path[1] IN (SELECT id FROM parkmaildb."Post" WHERE thread = $1 AND parent = 0 AND path[1] < (SELECT path[1] FROM parkmaildb."Post" WHERE id = $2) ORDER BY id DESC LIMIT $3) ORDER BY path[1] DESC, path, id`, id, since, limit)
		} else {
			rows, err = p.DB.Query(`SELECT id, parent, author, message, isedited, forum, thread, created FROM parkmaildb."Post" WHERE path[1] IN (SELECT id FROM parkmaildb."Post" WHERE thread = $1 AND parent = 0 AND path[1] > (SELECT path[1] FROM parkmaildb."Post" WHERE id = $2) ORDER BY id LIMIT $3) ORDER BY path, id`, id, since, limit)
		}
	}

	if err != nil {
		return []models.Post{}, false
	}

	return p.ParseRowsToPost(rows)
}

func (p PostRepository) GetAllPostByThread(id int, limit int, since int, desc bool) ([]models.Post, bool) {
	var rows *pgx.Rows
	var err error

	if since == 0 {
		if desc {
			rows, err = p.DB.Query(`SELECT id, parent, author, message, isedited, forum, thread, created FROM parkmaildb."Post" WHERE thread = $1 ORDER BY id DESC LIMIT $2`, id, limit)
		} else {
			rows, err = p.DB.Query(`SELECT id, parent, author, message, isedited, forum, thread, created FROM parkmaildb."Post" WHERE thread = $1 ORDER BY id LIMIT $2`, id, limit)
		}
	} else {
		if desc {
			rows, err = p.DB.Query(`SELECT id, parent, author, message, isedited, forum, thread, created FROM parkmaildb."Post" WHERE thread = $1 AND id < $2 ORDER BY id DESC LIMIT $3`, id, since, limit)
		} else {
			rows, err = p.DB.Query(`SELECT id, parent, author, message, isedited, forum, thread, created FROM parkmaildb."Post" WHERE thread = $1 AND id > $2 ORDER BY id LIMIT $3`, id, since, limit)
		}
	}

	if err != nil {
		return []models.Post{}, false
	}

	return p.ParseRowsToPost(rows)
}

func (p PostRepository) GetAllInfo(params models.FullPostParams, id int) (models.FullPost, bool) {
	var info models.FullPost

	post := models.Post{}
	user := models.User{}
	forum := models.Forum{}
	thread := models.Thread{}

	err := p.DB.QueryRow(`SELECT id, parent, author, message, isedited, forum, thread, created FROM parkmaildb."Post" WHERE id = $1`, id).
		Scan(&post.Id, &post.Parent, &post.Author, &post.Message, &post.IsEdited, &post.Forum, &post.Thread, &post.Created)
	if err != nil {
		log.Println(err)
		return models.FullPost{}, false
	}
	info.Post = &post

	if params.User {
		err = p.DB.QueryRow(`SELECT nickname, fullname, about, email FROM parkmaildb."User" WHERE nickname = $1`, post.Author).
			Scan(&user.Nickname, &user.Fullname, &user.About, &user.Email)
		if err != nil {
			return models.FullPost{}, false
		}
		info.Author = &user
	}

	if params.Thread {
		err = p.DB.QueryRow(`SELECT id, title, author, forum, message, votes, slug, created FROM parkmaildb."Thread" WHERE id = $1`, post.Thread).
			Scan(&thread.Id, &thread.Title, &thread.Author, &thread.Forum, &thread.Message, &thread.Votes, &thread.Slug, &thread.Created)
		if err != nil {
			return models.FullPost{}, false
		}
		info.Thread = &thread
	}

	if params.Forum {
		err = p.DB.QueryRow(`SELECT title, "user", slug, posts, threads FROM parkmaildb."Forum" WHERE slug = $1`, post.Forum).
			Scan(&forum.Title, &forum.User, &forum.Slug, &forum.Posts, &forum.Threads)
		if err != nil {
			return models.FullPost{}, false
		}
		info.Forum = &forum
	}

	return info, true
}

func (p PostRepository) ChangePost(updateMessage models.PostUpdate, id int) (models.Post, bool) {
	var post models.Post
	err := p.DB.QueryRow(`UPDATE parkmaildb."Post" SET message = COALESCE(NULLIF($1, ''), message), isedited = CASE WHEN $1 = '' OR message=$1 THEN isedited else true end WHERE id = $2 RETURNING id, parent, author, message, isedited, forum, thread, created`, updateMessage.Message, id).
		Scan(&post.Id, &post.Parent, &post.Author, &post.Message, &post.IsEdited, &post.Forum, &post.Thread, &post.Created)

	if err != nil {
		log.Println(err)
		return models.Post{}, false
	}

	return post, true
}

func (p PostRepository) AddPosts(posts models.Posts, threadId int, forumName string) (models.Posts, error) {
	var insertedPosts models.Posts

	var sqlValues []interface{}
	sqlQuery := `INSERT INTO parkmaildb."Post" (PARENT, AUTHOR, MESSAGE, FORUM, THREAD) VALUES `

	if len(posts) == 0 {
		return models.Posts{}, nil
	}

	for i, post := range posts {
		if post.Parent != 0 {
			id := -1
			err := p.DB.QueryRow(`SELECT id FROM parkmaildb."Post" WHERE thread = $1 AND id = $2`, threadId, post.Parent).Scan(&id)
			if err == pgx.ErrNoRows {
				return nil, errors.New("Cant Find Parent")
			}
		}

		sqlValuesString := fmt.Sprintf("($%d, $%d, $%d, $%d, $%d),", i*5+1, i*5+2, i*5+3, i*5+4, i*5+5)

		sqlQuery += sqlValuesString
		sqlValues = append(sqlValues, post.Parent, post.Author, post.Message, forumName, threadId)
	}

	sqlQuery = strings.TrimSuffix(sqlQuery, ",")
	sqlQuery += ` RETURNING id, parent, author, message, isedited, forum, thread, created;`

	rows, err := p.DB.Query(sqlQuery, sqlValues...)

	if err != nil {
		return nil, errors.New("Cant Find Parent")
	}

	for rows.Next() {
		post := models.Post{}

		err := rows.Scan(&post.Id, &post.Parent, &post.Author, &post.Message, &post.IsEdited, &post.Forum, &post.Thread, &post.Created)
		log.Println(1)
		log.Println(post)
		if err != nil || post.Author == "" {
			return nil, err
		}

		insertedPosts = append(insertedPosts, post)
	}

	if len(insertedPosts) == 0 {
		return nil, errors.New(models.ErrUserUnknown)
	}

	return insertedPosts, nil
}
