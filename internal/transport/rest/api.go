package rest

import (
	"log"
	rssService "rvneural/rss/internal/services/rss"

	"github.com/gin-gonic/gin"
)

type RSS struct {
}

func New() *RSS {
	return &RSS{}
}

func (r *RSS) Get(c *gin.Context) {
	url := "https://media.kpfu.ru/news-rss"
	service := rssService.New("Реальное время", "Сводка новостей", "http://realnoevremya.ru/")
	feed, err := service.Parse(url)
	if err != nil {
		log.Println(err)
		c.XML(200, gin.H{"error": err.Error()})
		return
	}
	c.XML(200, service.Merge(true, feed))
}
