package rss

import (
	"encoding/xml"
	"io"
	"log"
	"net/http"
	"rvneural/rss/internal/models/rss"
	"strings"
	"time"

	"github.com/denisbrodbeck/striphtmltags"
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

func (p *Parser) parseAtom(link string) (*rss.RSS, error) {
	feed, err := p.parser.ParseURL(link)
	if err != nil {
		return nil, err
	}
	var myFeed rss.RSS
	myFeed.Channel.Title = feed.Title
	myFeed.Channel.Description = feed.Description
	myFeed.Channel.Link = feed.Link
	for _, item := range feed.Items {
		myItem := rss.Item{
			Title:       item.Title,
			Description: item.Description,
			Link:        item.Link,
			PubDate:     item.Published,
		}
		myFeed.Channel.Items = append(myFeed.Channel.Items, myItem)
	}
	return p.clearHTML(&myFeed), nil
}

func (p *Parser) clearHTML(rss *rss.RSS) *rss.RSS {
	for i, item := range rss.Channel.Items {
		rss.Channel.Items[i].Description = strings.TrimSpace(striphtmltags.StripTags(item.Description))
		rss.Channel.Items[i].FullText = strings.TrimSpace(striphtmltags.StripTags(item.FullText))
		rss.Channel.Items[i].Summary = strings.TrimSpace(striphtmltags.StripTags(item.Summary))
		rss.Channel.Items[i].Title = strings.TrimSpace(striphtmltags.StripTags(item.Title))
	}
	return rss
}

func (p *Parser) Parse(link string) (*rss.RSS, error) {
	request, err := http.NewRequest(http.MethodGet, link, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/rss+xml")
	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	request.Header.Set("Accept-Language", "ru-RU,ru;q=0.6")
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
	if err != nil {
		log.Println("Parsing atom")
		return p.parseAtom(link)
	}
	newItems := make([]rss.Item, len(feed.Channel.Items))
	for i, item := range feed.Channel.Items {
		if len(strings.TrimSpace(item.FullText)) == 0 && strings.TrimSpace(item.Summary) != "" {
			item.FullText = item.Summary
			item.Summary = ""
		}
		newItems[i] = item
	}
	feed.Channel.Items = newItems
	return p.clearHTML(&feed), nil
}

func (p *Parser) Merge(today bool, feeds ...*rss.RSS) *rss.RSS {
	var myFeed rss.RSS
	myFeed.Channel.Title = p.title
	myFeed.Channel.Description = p.description
	myFeed.Channel.Link = p.link
	myFeed.Version = "2.0"
	for _, feed := range feeds {
		for _, item := range feed.Channel.Items {
			date, _ := time.Parse(time.RFC1123Z, item.PubDate)
			if today {
				log.Println("Параметры:", item.PubDate)
				log.Println("Дата:", date)
				log.Println("Сегодня:", time.Now())
			}
			if today && (date.Day() != time.Now().Day() || date.Month() != time.Now().Month()) {
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
		pivotDate, _ := time.Parse(time.RFC1123Z, pivot.PubDate)
		itemDate, _ := time.Parse(time.RFC1123Z, item.PubDate)
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
