package rest

import (
	"os"
	"rvneural/rss/internal/models/db"
	"rvneural/rss/internal/models/rss"
	dbService "rvneural/rss/internal/services/db"
	rssService "rvneural/rss/internal/services/rss"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type RSS struct {
	db       *dbService.Service
	feedList []db.RSS
	lastRead time.Time
	service  *rssService.Parser
}

func New() *RSS {
	service := rssService.New("Реальное время", "Сводка новостей", "http://realnoevremya.ru/")
	return &RSS{
		db:      dbService.New(),
		service: service,
	}
}

type res struct {
	URL   string
	Title string
}

func (r *RSS) GetFeed(c *gin.Context) {
	today := c.Query("today") == "1" || c.Query("today") == "true"
	fulltext := c.Query("text") == "1" || c.Query("text") == "true"
	updateTime, err := strconv.Atoi(os.Getenv("UPDATE_TIME"))
	if err != nil {
		updateTime = 30
	}
	if r.feedList == nil || time.Now().Sub(r.lastRead) > time.Duration(updateTime)*time.Minute {
		var err error
		r.feedList, err = r.db.GetFeeds()
		if err != nil {
			c.AbortWithError(500, err)
			r.feedList = nil
			return
		}
		r.lastRead = time.Now()
	}
	mutex := sync.Mutex{}
	wg := sync.WaitGroup{}

	var feeds []*rss.RSS
	for _, feedElement := range r.feedList {
		wg.Add(1)
		go func(feeds *([]*rss.RSS)) {
			defer wg.Done()
			feed, err := r.service.Parse(feedElement.URL, feedElement.Title, fulltext, today)
			if err != nil {
				return
			}
			mutex.Lock()
			*feeds = append(*feeds, feed)
			mutex.Unlock()
		}(&feeds)
	}
	wg.Wait()
	c.XML(200, r.service.Merge(today, feeds...))
}

func (r *RSS) GetFeedList(c *gin.Context) {
	feeds, err := r.db.GetFeeds()
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	c.JSON(200, gin.H{"feeds": feeds})
}

func (r *RSS) AddFeed(c *gin.Context) {
	var feed db.RSS
	err := c.BindJSON(&feed)
	if err != nil {
		c.AbortWithError(400, err)
		return
	}
	if feed.URL == "" || feed.Title == "" {
		c.AbortWithError(400, err)
		return
	}
	if feed.Type == "" {
		feed.Type = "rss"
	}
	err = r.db.AddFeed(feed)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	c.JSON(200, feed)
}

func (r *RSS) DeleteFeed(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.AbortWithError(400, err)
		return
	}
	err = r.db.DeleteFeed(id)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	c.JSON(200, gin.H{"id": id})
}
