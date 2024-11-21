package db

import (
	"fmt"
	"os"
	model "rvneural/rss/internal/models/db"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Service struct {
	host     string
	port     string
	username string
	password string
	db_name  string
	table    string
}

func New() *Service {
	return &Service{
		host:     os.Getenv("DB_HOST"),
		port:     os.Getenv("DB_PORT"),
		username: os.Getenv("DB_USERNAME"),
		password: os.Getenv("DB_PASSWORD"),
		db_name:  os.Getenv("DB_NAME"),
		table:    os.Getenv("DB_TABLE"),
	}
}

func (s *Service) connect() (*sqlx.DB, error) {
	if s.host == "" || s.port == "" || s.username == "" || s.password == "" || s.db_name == "" || s.table == "" {
		return nil, fmt.Errorf("empty db credentials")
	}
	connectionData := fmt.Sprintf("user=%s dbname=%s sslmode=disable password=%s host=%s port=%s", s.username, s.db_name, s.password, s.host, s.port)
	return sqlx.Connect("postgres", connectionData)
}

func (s *Service) GetFeeds() (dbResult []model.RSS, err error) {
	dbResult = make([]model.RSS, 0, 35)
	db, err := s.connect()
	if err != nil {
		return dbResult, err
	}
	defer db.Close()

	request := "SELECT * FROM " + s.table

	err = db.Select(&dbResult, request)
	return dbResult, err
}

func (s *Service) AddFeed(feed model.RSS) (err error) {
	db, err := s.connect()
	if err != nil {
		return err
	}
	defer db.Close()

	request := "INSERT INTO " + s.table + " (url, title) VALUES ($1, $2, $3, $4)"
	_, err = db.Exec(request, feed.URL, feed.Title)
	return err
}

func (s *Service) DeleteFeed(id int) (err error) {
	db, err := s.connect()
	if err != nil {
		return err
	}
	defer db.Close()

	request := "DELETE FROM " + s.table + " WHERE id = $1"
	_, err = db.Exec(request, id)
	return err
}
