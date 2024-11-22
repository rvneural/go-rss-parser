package rss

import (
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"rvneural/rss/internal/models/rss"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/denisbrodbeck/striphtmltags"
	"github.com/mmcdole/gofeed"
)

type Parser struct {
	title       string
	description string
	link        string
	parser      *gofeed.Parser
	timeout     time.Duration
}

func New(title, description, link string) *Parser {
	maxTimeOut, err := strconv.Atoi(os.Getenv("MAX_TIMEOUT"))
	if err != nil {
		maxTimeOut = 1
	}
	if maxTimeOut < 1 {
		maxTimeOut = 1
	}
	return &Parser{
		title:       title,
		description: description,
		link:        link,
		parser:      gofeed.NewParser(),
		timeout:     time.Second * time.Duration(maxTimeOut),
	}
}

func (p *Parser) parseAtomWithChannel(link string) <-chan *gofeed.Feed {
	ch := make(chan *gofeed.Feed, 1)
	feed, err := p.parser.ParseURL(link)
	if err != nil {
		log.Println("Atom ERROR with:", link, "Error:", err)
		close(ch)
		return ch
	}

	ch <- feed
	close(ch)
	return ch
}

func (p *Parser) parseAtom(link, title string) (*rss.RSS, error) {
	select {
	case <-time.After(p.timeout):
		return nil, fmt.Errorf("Timeout")
	case feed := <-p.parseAtomWithChannel(link):
		var myFeed rss.RSS
		if feed == nil {
			return nil, fmt.Errorf("Timeout")
		}
		myFeed.Channel.Title = feed.Title
		myFeed.Channel.Description = feed.Description
		myFeed.Channel.Link = feed.Link
		wg := sync.WaitGroup{}
		for _, item := range feed.Items {
			wg.Add(1)
			go func(myFeed *rss.RSS) {
				defer wg.Done()
				myItem := rss.Item{
					Title:       item.Title,
					Description: item.Description,
					Link:        item.Link,
					PubDate:     item.PublishedParsed.Format(time.RFC1123Z),
				}
				myFeed.Channel.Items = append(myFeed.Channel.Items, myItem)
			}(&myFeed)
		}
		wg.Wait()
		myFeed.Channel.Title = title
		return p.clearHTML(&myFeed), nil
	}
}

func (p *Parser) clearHTML(rss *rss.RSS) *rss.RSS {
	eastOfUTC := time.FixedZone("UTC+3", 3*60*60)
	for i, item := range rss.Channel.Items {

		if rss.Channel.Items[i].Description != "" {
			rss.Channel.Items[i].Description = strings.TrimSpace(striphtmltags.StripTags(item.Description))
			rss.Channel.Items[i].Description = strings.ReplaceAll(rss.Channel.Items[i].Description, "&laquo;", "«")
			rss.Channel.Items[i].Description = strings.ReplaceAll(rss.Channel.Items[i].Description, "&raquo;", "»")
			rss.Channel.Items[i].Description = strings.ReplaceAll(rss.Channel.Items[i].Description, "&nbsp;", " ")
			rss.Channel.Items[i].Description = strings.ReplaceAll(rss.Channel.Items[i].Description, "&mdash;", "—")
			rss.Channel.Items[i].Description = strings.ReplaceAll(rss.Channel.Items[i].Description, "&ndash;", "–")
			rss.Channel.Items[i].Description = strings.ReplaceAll(rss.Channel.Items[i].Description, "&quot;", "\"")
			rss.Channel.Items[i].Description = strings.ReplaceAll(rss.Channel.Items[i].Description, "&amp;", "&")
		}

		if rss.Channel.Items[i].FullText != "" {
			rss.Channel.Items[i].FullText = strings.TrimSpace(striphtmltags.StripTags(item.FullText))
			rss.Channel.Items[i].FullText = strings.ReplaceAll(rss.Channel.Items[i].FullText, "&laquo;", "«")
			rss.Channel.Items[i].FullText = strings.ReplaceAll(rss.Channel.Items[i].FullText, "&raquo;", "»")
			rss.Channel.Items[i].FullText = strings.ReplaceAll(rss.Channel.Items[i].FullText, "&nbsp;", " ")
			rss.Channel.Items[i].FullText = strings.ReplaceAll(rss.Channel.Items[i].FullText, "&mdash;", "—")
			rss.Channel.Items[i].FullText = strings.ReplaceAll(rss.Channel.Items[i].FullText, "&ndash;", "–")
			rss.Channel.Items[i].FullText = strings.ReplaceAll(rss.Channel.Items[i].FullText, "&quot;", "\"")
			rss.Channel.Items[i].FullText = strings.ReplaceAll(rss.Channel.Items[i].FullText, "&amp;", "&")
		}

		if rss.Channel.Items[i].Title != "" {
			rss.Channel.Items[i].Title = strings.TrimSpace(striphtmltags.StripTags(item.Title))
			rss.Channel.Items[i].Title = strings.ReplaceAll(rss.Channel.Items[i].Title, "&laquo;", "«")
			rss.Channel.Items[i].Title = strings.ReplaceAll(rss.Channel.Items[i].Title, "&raquo;", "»")
			rss.Channel.Items[i].Title = strings.ReplaceAll(rss.Channel.Items[i].Title, "&nbsp;", " ")
			rss.Channel.Items[i].Title = strings.ReplaceAll(rss.Channel.Items[i].Title, "&mdash;", "—")
			rss.Channel.Items[i].Title = strings.ReplaceAll(rss.Channel.Items[i].Title, "&ndash;", "–")
			rss.Channel.Items[i].Title = strings.ReplaceAll(rss.Channel.Items[i].Title, "&quot;", "\"")
			rss.Channel.Items[i].Title = strings.ReplaceAll(rss.Channel.Items[i].Title, "&amp;", "&")
		}

		rss.Channel.Items[i].Link = strings.TrimSpace(item.Link)

		date, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			continue
		}
		date = p.inSameClock(date, eastOfUTC)
		rss.Channel.Items[i].PubDate = date.Format(time.RFC1123Z)

	}
	return rss
}

func (p *Parser) Parse(link, title string) (*rss.RSS, error) {

	if strings.Contains(link, "www.rospotrebnadzor.ru") {
		return p.parseAtom(link, title)
	}

	request, err := http.NewRequest(http.MethodGet, link, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Accept", "application/rss+xml")
	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	request.Header.Set("Accept-Language", "ru-RU,ru;q=0.6")
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: p.timeout,
	}
	response, err := client.Do(request)
	if err != nil {
		log.Println("ERROR with:", link, "Error:", err)
		return nil, err
	}
	defer response.Body.Close()
	byteFeed, err := io.ReadAll(response.Body)
	if err != nil {
		log.Println("ERROR with:", link, "Error:", err)
		return nil, err
	}

	var feed rss.RSS
	err = xml.Unmarshal(byteFeed, &feed)
	if err != nil {
		return p.parseAtom(link, title)
	}
	newItems := make([]rss.Item, len(feed.Channel.Items))
	wg := sync.WaitGroup{}
	for i, item := range feed.Channel.Items {
		wg.Add(1)
		go func(newItems *[]rss.Item, i int) {
			defer wg.Done()
			if len(strings.TrimSpace(item.FullText)) == 0 && strings.TrimSpace(item.Summary) != "" {
				item.FullText = item.Summary
				item.Summary = ""
			}
			(*newItems)[i] = item
		}(&newItems, i)
	}
	wg.Wait()
	feed.Channel.Items = newItems
	feed.Channel.Title = title
	return p.clearHTML(&feed), nil
}

func (p *Parser) Merge(today bool, feeds ...*rss.RSS) *rss.RSS {
	var myFeed rss.RSS
	myFeed.Channel.Title = p.title
	myFeed.Channel.Description = p.description
	myFeed.Channel.Link = p.link
	myFeed.Version = "2.0"
	wg := sync.WaitGroup{}
	mutex := sync.Mutex{}
	eastOfUTC := time.FixedZone("UTC+3", 3*60*60)
	for _, feed := range feeds {
		wg.Add(1)
		go func(feed *rss.RSS, myFeed *rss.RSS) {
			defer wg.Done()
			wg2 := sync.WaitGroup{}
			for _, item := range feed.Channel.Items {
				wg2.Add(1)
				go func(item *rss.Item, myFeed *rss.RSS) {
					defer wg2.Done()
					date, _ := time.Parse(time.RFC1123Z, item.PubDate)
					date = p.inSameClock(date, eastOfUTC)
					if today && (time.Now().Day() > date.Day() || time.Now().Month() != date.Month() || time.Now().Year() != date.Year()) {
						return
					}
					if feed.Channel.Title != "" {
						item.Source = feed.Channel.Title
					} else {
						item.Source = feed.Channel.Description
					}
					mutex.Lock()
					myFeed.Channel.Items = append(myFeed.Channel.Items, *item)
					mutex.Unlock()
				}(&item, myFeed)
			}
			wg2.Wait()
		}(feed, &myFeed)
	}
	wg.Wait()
	myFeed.Channel.Items = p.sortByDate(myFeed.Channel.Items...)
	myFeed.Length = strconv.Itoa(len(myFeed.Channel.Items))
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
	sortedItems = append(sortedItems, p.sortByDate(greater...)...)
	sortedItems = append(sortedItems, pivot)
	sortedItems = append(sortedItems, p.sortByDate(less...)...)

	return sortedItems
}

func (p *Parser) inSameClock(t time.Time, loc *time.Location) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), loc)
}
