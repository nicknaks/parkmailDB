package usecase

import (
	"encoding/json"
	"forum/internal/utils/utils"
	"forum/pkg/models"
	"forum/pkg/thread/repository"
	"github.com/pkg/errors"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"
)

type ThreadUsecaseInterface interface {
	CreateThread(thread models.Thread) (models.Thread, int, error)

	ParseJsonToThread(body io.ReadCloser) (models.Thread, error)
	GetThreadByRequest(body io.ReadCloser, vars map[string]string) (models.Thread, bool)
	FindThreadsByParams(slug string, params models.ParamsForSearch) ([]models.Thread, bool)
	ParseJsonToUpdateThread(body io.ReadCloser) (models.ThreadUpdate, error)
	UpdateThread(update models.ThreadUpdate, slugOrId string) (models.Thread, bool)
}

type ThreadUsecase struct {
	DB repository.ThreadRepositoryInterface
}

func (u ThreadUsecase) FindThreadsByParams(slug string, params models.ParamsForSearch) ([]models.Thread, bool) {
	threads, ok := u.DB.FindThreads(slug, params)
	if !ok {
		return nil, false
	}

	return threads, true
}

func (u ThreadUsecase) CreateThread(thread models.Thread) (models.Thread, int, error) {
	insertedThread, err := u.DB.CreateThread(thread)
	if err == nil {
		return insertedThread, http.StatusCreated, nil
	}

	log.Println(err)
	code := utils.PgxErrorCode(err)
	if code == "23503" { //ошибка отсутствия c юзером/форумом
		return models.Thread{}, http.StatusNotFound, errors.New(models.ErrUserUnknown)
	}

	thread, ok := u.DB.GetThreadInfo(thread.Title, thread.Forum)
	if !ok {
		return models.Thread{}, http.StatusNotFound, errors.New(models.ErrUserUnknown)
	}

	return thread, http.StatusConflict, nil
}

func (u ThreadUsecase) ParseJsonToThread(body io.ReadCloser) (models.Thread, error) {
	defer body.Close()
	var thread models.Thread

	decoder := json.NewDecoder(body)
	err := decoder.Decode(&thread)

	if err != nil {
		log.Println(err)
	}

	thread.Votes = 0
	thread.Created = time.Now()
	return thread, err
}

func (u ThreadUsecase) GetThreadByRequest(body io.ReadCloser, vars map[string]string) (models.Thread, bool) {
	thread, err := u.ParseJsonToThread(body)
	if err != nil {
		return models.Thread{}, false
	}

	var ok bool
	thread.Forum, ok = utils.GetDataFromPath("slug", vars)
	if !ok {
		return models.Thread{}, false
	}

	return thread, true
}

func (u ThreadUsecase) GetThreadInfo(slugOrId string) (models.Thread, bool) {
	id, err := strconv.Atoi(slugOrId)
	if err != nil {
		return u.DB.GetThreadInfoBySlug(slugOrId)
	}

	return u.DB.GetThreadInfoById(id)
}

func (u ThreadUsecase) UpdateThread(update models.ThreadUpdate, slugOrId string) (models.Thread, bool) {
	return u.DB.UpdateThread(update, slugOrId)
}

func (u ThreadUsecase) ParseJsonToUpdateThread(body io.ReadCloser) (models.ThreadUpdate, error) {
	defer body.Close()
	var thread models.ThreadUpdate

	decoder := json.NewDecoder(body)
	err := decoder.Decode(&thread)

	if err != nil {
		log.Println(err)
	}

	return thread, err
}

func (u ThreadUsecase) ParseJsonToVote(body io.ReadCloser) (models.Vote, error) {
	defer body.Close()
	var vote models.Vote

	decoder := json.NewDecoder(body)
	err := decoder.Decode(&vote)

	if err != nil {
		log.Println(err)
	}

	if math.Abs(float64(vote.Voice)) != 1 {
		err = errors.New("Voice Can be only -1 or 1")
		log.Println(err)
		return vote, err
	}

	return vote, err
}

func (u ThreadUsecase) SetVote(vote models.Vote, slugOrId string) (models.Thread, bool) {
	id, err := strconv.Atoi(slugOrId)
	if err != nil {
		var ok bool
		id, ok = u.DB.GetThreadIdBySlug(slugOrId)
		if !ok {
			return models.Thread{}, false
		}
	}

	ok := u.DB.SetVote(vote, id)
	if !ok {
		return models.Thread{}, false
	}

	thread, ok := u.DB.GetThreadInfoById(id)
	if !ok {
		return models.Thread{}, false
	}

	return thread, true
}
