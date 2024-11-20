package rss

import (
	"encoding/xml"
	"io"
	"net/http"
	"rvneural/rss/internal/models/rss"
	"time"

	"github.com/mmcdole/gofeed"
)

type Parser struct {
	title       string
	description string
	link        string
	parser      *gofeed.Parser
}

func New(title, description, link string) *Parser {
	return &Parser{
		title:       title,
		description: description,
		link:        link,
		parser:      gofeed.NewParser(),
	}
}

func (p *Parser) Parse(link string) (*rss.RSS, error) {
	request, err := http.NewRequest(http.MethodGet, link, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/rss+xml")
	//request.Header.Set("Accept", "application/xhtml+xml,application/xml")
	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	byteFeed, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var feed rss.RSS
	err = xml.Unmarshal(byteFeed, &feed)
	return &feed, err
}

func (p *Parser) Merge(today bool, feeds ...*rss.RSS) *rss.RSS {
	var myFeed rss.RSS
	myFeed.Channel.Title = p.title
	myFeed.Channel.Description = p.description
	myFeed.Channel.Link = p.link
	for _, feed := range feeds {
		for _, item := range feed.Channel.Items {
			date, _ := time.Parse(time.RFC1123, item.PubDate)
			if today && (date.Day() != time.Now().Day()) {
				continue
			}
			if feed.Channel.Title != "" {
				item.Source = feed.Channel.Title
			} else {
				item.Source = feed.Channel.Description
			}
			myFeed.Channel.Items = append(myFeed.Channel.Items, item)
		}
	}
	myFeed.Channel.Items = p.sortByDate(myFeed.Channel.Items...)
	return &myFeed
}

func (p *Parser) sortByDate(items ...rss.Item) []rss.Item {
	if len(items) < 2 {
		return items
	}
	var sortedItems []rss.Item
	pivot := items[0]
	var less, greater []rss.Item
	for _, item := range items[1:] {
		pivotDate, _ := time.Parse(time.RFC1123, pivot.PubDate)
		itemDate, _ := time.Parse(time.RFC1123, item.PubDate)
		if itemDate.Before(pivotDate) {
			less = append(less, item)
		} else {
			greater = append(greater, item)
		}
	}
	sortedItems = append(sortedItems, p.sortByDate(less...)...)
	sortedItems = append(sortedItems, pivot)
	sortedItems = append(sortedItems, p.sortByDate(greater...)...)

	return sortedItems
}
