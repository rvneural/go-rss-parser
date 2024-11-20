package rest

import (
	"rvneural/rss/internal/models/rss"
	dbService "rvneural/rss/internal/services/db"
	rssService "rvneural/rss/internal/services/rss"

	"github.com/gin-gonic/gin"
)

type RSS struct {
	db *dbService.Service
}

func New() *RSS {
	return &RSS{
		db: dbService.New(),
	}
}

type res struct {
	URL   string
	Title string
}

func (r *RSS) Get(c *gin.Context) {

	var feeds []*rss.RSS
	db_feeds, err := r.db.GetFeeds()
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	service := rssService.New("Реальное время", "Сводка новостей", "http://realnoevremya.ru/")
	for _, feed := range db_feeds {
		feed, err := service.Parse(feed.URL, feed.Title)
		if err != nil {
			continue
		}
		feeds = append(feeds, feed)
	}
	c.XML(200, service.Merge(true, feeds...))
}
