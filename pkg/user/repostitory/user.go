package repostitory

import (
	"forum/pkg/models"
	_ "github.com/jackc/pgx"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"log"
)

type UserRepositoryInterface interface {
	AddUser(user models.User) ([]models.User, bool)
	GetUser(nickname string) (models.User, error)
	ChangeUser(user models.User) (models.User, error)
}

type UserRepository struct {
	DB *sqlx.DB
}

func (u *UserRepository) AddUser(user models.User) ([]models.User, bool) {
	_, err := u.DB.Exec(`INSERT INTO parkmaildb."User" (nickname, fullname, about, email) VALUES ($1, $2, $3, $4)`,
		user.Nickname, user.Fullname, user.About, user.Email)
	if err == nil {
		return []models.User{user}, true
	}

	log.Println(err)
	rows, err := u.DB.Queryx(`SELECT nickname, fullname, about, email FROM parkmaildb."User" WHERE lower(nickname) = lower($1) OR lower(email) = lower($2)`,
		user.Nickname, user.Email)
	if err != nil {
		log.Println(err)
	}

	var users []models.User

	for rows.Next() {
		newUser := models.User{}
		_ = rows.Scan(&newUser.Nickname, &newUser.Fullname, &newUser.About, &newUser.Email)
		users = append(users, newUser)
	}

	return users, false
}

func (u *UserRepository) ChangeUser(user models.User) (models.User, error) {
	var newUser models.User

	err := u.DB.QueryRowx(`UPDATE parkmaildb."User" 
					SET fullname = COALESCE(NULLIF($1, ''), fullname), about = COALESCE(NULLIF($2, ''), about), email = COALESCE(NULLIF($3, ''), email)
					WHERE lower(nickname) = lower($4) 
					RETURNING nickname, fullname, about, email`,
		user.Fullname, user.About, user.Email, user.Nickname).
		Scan(&newUser.Nickname, &newUser.Fullname, &newUser.About, &newUser.Email)

	if err != nil {
		return models.User{}, err
	}

	return newUser, nil
}

func (u UserRepository) GetUser(nickname string) (models.User, error) {
	var user models.User = models.User{Nickname: nickname}

	err := u.DB.QueryRow(`SELECT u.nickname, u.fullname, u.about, u.email FROM parkmaildb."User" u WHERE lower(u.nickname) = lower($1)`,
		nickname).Scan(&user.Nickname, &user.Fullname, &user.About, &user.Email)
	if err != nil {
		log.Println(err)
		return models.User{}, err
	}

	return user, nil
}
