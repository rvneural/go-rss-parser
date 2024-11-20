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
	service := rssService.New("Реальное время", "Сводка новостей", "http://realnoevremya.ru/")
	rais_feed, err_rais := service.Parse("http://www.kremlin.ru/events/all/feed")
	mvd_feed, err_mvd := service.Parse("https://16.xn--b1aew.xn--p1ai/news/rss")

	if err_rais == nil && err_mvd == nil {
		sum_feed := service.Merge(false, rais_feed, mvd_feed)
		c.XML(200, sum_feed)
	} else if err_rais != nil && err_mvd == nil {
		log.Print(err_rais.Error())
		c.XML(200, service.Merge(false, mvd_feed))
	} else if err_mvd != nil && err_rais == nil {
		log.Print(err_mvd.Error())
		c.XML(200, service.Merge(false, rais_feed))
	} else {
		c.XML(200, gin.H{"error": err_mvd.Error()})
	}
}
