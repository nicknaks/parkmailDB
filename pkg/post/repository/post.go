package repository

import (
	"database/sql"
	"forum/pkg/models"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"log"
)

type PostRepositoryInterface interface {
	AddPosts(posts []models.Post, threadId int, forumName string) ([]models.Post, error)
	ChangePost(updateMessage models.PostUpdate, id int) (models.Post, bool)
	GetAllInfo(params models.FullPostParams, id int) (models.FullPost, bool)
	GetAllPostByThread(id int, limit int, since int, desc bool) ([]models.Post, bool)
	GetPostsTree(id int, limit int, since int, desc bool) ([]models.Post, bool)
	GetPostsParentTree(id int, limit int, since int, desc bool) ([]models.Post, bool)
}

type PostRepository struct {
	DB *sqlx.DB
}

func (p PostRepository) ParseRowsToPost(rows *sqlx.Rows) ([]models.Post, bool) {
	var posts []models.Post
	var post models.Post
	for rows.Next() {
		err := rows.Scan(&post.Id, &post.Parent, &post.Author, &post.Message, &post.IsEdited, &post.Forum, &post.Thread, &post.Created)
		if err != nil {
			return nil, false
		}
		posts = append(posts, post)
	}

	return posts, true
}

func (p PostRepository) GetPostsTree(id int, limit int, since int, desc bool) ([]models.Post, bool) {
	var rows *sqlx.Rows
	var err error

	if since == 0 {
		if desc {
			rows, err = p.DB.Queryx(`SELECT id, parent, author, message, isedited, forum, thread, created FROM parkmaildb."Post" WHERE thread = $1 ORDER BY path DESC, id DESC LIMIT $2`, id, limit)
		} else {
			rows, err = p.DB.Queryx(`SELECT id, parent, author, message, isedited, forum, thread, created FROM parkmaildb."Post" WHERE thread = $1 ORDER BY path, id LIMIT $2`, id, limit)
		}
	} else {
		if desc {
			rows, err = p.DB.Queryx(`SELECT id, parent, author, message, isedited, forum, thread, created FROM parkmaildb."Post" WHERE thread = $1 AND path < (SELECT path FROM parkmaildb."Post" where id = $2) ORDER BY path DESC , id DESC LIMIT $3`, id, since, limit)
		} else {
			rows, err = p.DB.Queryx(`SELECT id, parent, author, message, isedited, forum, thread, created FROM parkmaildb."Post" WHERE thread = $1 AND path > (SELECT path FROM parkmaildb."Post" where id = $2) ORDER BY path, id LIMIT $3`, id, since, limit)
		}
	}

	if err != nil {
		return nil, false
	}

	return p.ParseRowsToPost(rows)
}

func (p PostRepository) GetPostsParentTree(id int, limit int, since int, desc bool) ([]models.Post, bool) {
	var rows *sqlx.Rows
	var err error

	if since == 0 {
		if desc {
			rows, err = p.DB.Queryx(`SELECT id, parent, author, message, isedited, forum, thread, created FROM parkmaildb."Post" WHERE path[1] IN (SELECT id FROM parkmaildb."Post" WHERE thread = $1 AND parent IS NULL ORDER BY id DESC LIMIT $2) ORDER BY path[1] DESC, path, id`, id, limit)
		} else {
			rows, err = p.DB.Queryx(`SELECT id, parent, author, message, isedited, forum, thread, created FROM parkmaildb."Post" WHERE path[1] IN (SELECT id FROM parkmaildb."Post" WHERE thread = $1 AND parent IS NULL ORDER BY id DESC LIMIT $2) ORDER BY path, id`, id, limit)
		}
	} else {
		if desc {
			rows, err = p.DB.Queryx(`SELECT id, parent, author, message, isedited, forum, thread, created FROM parkmaildb."Post" WHERE path[1] IN (SELECT id FROM parkmaildb."Post" WHERE thread = $1 AND parent IS NULL AND path[1] < (SELECT path[1] FROM parkmaildb."Post" WHERE id = $2) ORDER BY id DESC LIMIT $3) ORDER BY path[1] DESC, path, id`, id, since, limit)
		} else {
			rows, err = p.DB.Queryx(`SELECT id, parent, author, message, isedited, forum, thread, created FROM parkmaildb."Post" WHERE path[1] IN (SELECT id FROM parkmaildb."Post" WHERE thread = $1 AND parent IS NULL AND path[1] > (SELECT path[1] FROM parkmaildb."Post" WHERE id = $2) ORDER BY id DESC LIMIT $3) ORDER BY path, id`, id, since, limit)
		}
	}

	if err != nil {
		return nil, false
	}

	return p.ParseRowsToPost(rows)
}

func (p PostRepository) GetAllPostByThread(id int, limit int, since int, desc bool) ([]models.Post, bool) {
	var rows *sqlx.Rows
	var err error

	if since == 0 {
		if desc {
			rows, err = p.DB.Queryx(`SELECT id, parent, author, message, isedited, forum, thread, created FROM parkmaildb."Post" WHERE thread = $1 ORDER BY id DESC LIMIT $2`, id, limit)
		} else {
			rows, err = p.DB.Queryx(`SELECT id, parent, author, message, isedited, forum, thread, created FROM parkmaildb."Post" WHERE thread = $1 ORDER BY id LIMIT $2`, id, limit)
		}
	} else {
		if desc {
			rows, err = p.DB.Queryx(`SELECT id, parent, author, message, isedited, forum, thread, created FROM parkmaildb."Post" WHERE thread = $1 AND id < $2 ORDER BY id DESC LIMIT $3`, id, since, limit)
		} else {
			rows, err = p.DB.Queryx(`SELECT id, parent, author, message, isedited, forum, thread, created FROM parkmaildb."Post" WHERE thread = $1 AND id > $2 ORDER BY id LIMIT $3`, id, since, limit)
		}
	}

	if err != nil {
		return nil, false
	}

	return p.ParseRowsToPost(rows)
}

func (p PostRepository) GetAllInfo(params models.FullPostParams, id int) (models.FullPost, bool) {
	var info models.FullPost

	post := models.Post{}
	user := models.User{}
	forum := models.Forum{}
	thread := models.Thread{}

	err := p.DB.QueryRowx(`SELECT id, parent, author, message, isedited, forum, thread, created FROM parkmaildb."Post" WHERE id = $1`, id).
		Scan(&post.Id, &post.Parent, &post.Author, &post.Message, &post.IsEdited, &post.Forum, &post.Thread, &post.Created)
	if err != nil {
		log.Println(err)
		return models.FullPost{}, false
	}
	info.Post = &post

	if params.User {
		err = p.DB.QueryRowx(`SELECT nickname, fullname, about, email FROM parkmaildb."User" WHERE nickname = $1`, post.Author).
			Scan(&user.Nickname, &user.Fullname, &user.About, &user.Email)
		if err != nil {
			return models.FullPost{}, false
		}
		info.Author = &user
	}

	if params.Thread {
		err = p.DB.QueryRowx(`SELECT id, title, author, forum, message, votes, slug, created FROM parkmaildb."Thread" WHERE id = $1`, post.Thread).
			Scan(&thread.Id, &thread.Title, &thread.Author, &thread.Forum, &thread.Message, &thread.Votes, &thread.Slug, &thread.Created)
		if err != nil {
			return models.FullPost{}, false
		}
		info.Thread = &thread
	}

	if params.Forum {
		err = p.DB.QueryRowx(`SELECT title, "user", slug, posts, threads FROM parkmaildb."Forum" WHERE slug = $1`, post.Forum).
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
	err := p.DB.QueryRowx(`UPDATE parkmaildb."Post" SET message = $1, isedited = TRUE WHERE id = $2 RETURNING id, parent, author, message, isedited, forum, thread, created`, updateMessage.Message, id).
		Scan(&post.Id, &post.Parent, &post.Author, &post.Message, &post.IsEdited, &post.Forum, &post.Thread, &post.Created)

	if err != nil {
		log.Println(err)
		return models.Post{}, false
	}

	return post, true
}

func (p PostRepository) AddPosts(posts []models.Post, threadId int, forumName string) ([]models.Post, error) {
	var insertedPosts []models.Post
	begin, err := p.DB.Beginx()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	for _, post := range posts {
		if post.Parent != 0 {
			id := -1
			err = p.DB.QueryRowx(`SELECT id FROM parkmaildb."Post" WHERE thread = $1 AND id = $2`, threadId, post.Parent).Scan(&id)
			if err == sql.ErrNoRows {
				begin.Rollback()
				return nil, errors.New("Cant Find Parent")
			}
		}

		var insertedPost models.Post

		err = begin.QueryRowx(`INSERT INTO parkmaildb."Post" (PARENT, AUTHOR, MESSAGE, ISEDITED, FORUM, THREAD, CREATED) VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id, parent, author, message, isedited, forum, thread, created`,
			post.Parent, post.Author, post.Message, false, forumName, threadId, post.Created).
			Scan(&insertedPost.Id, &insertedPost.Parent, &insertedPost.Author, &insertedPost.Message, &insertedPost.IsEdited, &insertedPost.Forum, &insertedPost.Thread, &insertedPost.Created)

		if err != nil {
			log.Println(err)
			begin.Rollback()
			return nil, err
		}

		insertedPosts = append(insertedPosts, insertedPost)
	}
	err = begin.Commit()
	if err != nil {
		log.Println(err)
		begin.Rollback()
		return nil, err
	}

	return insertedPosts, nil
}
