package article

import (
	"fmt"
	"net/http"
	"scraper/formatter"
	"scraper/models"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/mmcdole/gofeed"
)

type ContentFilter struct {
	RemoveSelectors      []string
	RemoveAfterSelectors []string
	CustomFilter         func(*goquery.Selection)
}

type ParseConfig struct {
	ContentSelector  string
	TitleSelector    string
	SubtitleSelector string
	DateSelector     string
	ContentFilter    ContentFilter
	TitleProcessor   func(string) string
}

type Article struct {
	BaseURL      string
	RSSEndPoints []string
	ParseConfig  ParseConfig
}

func NewArticle() *Article {
	return &Article{
		BaseURL: "https://www.wired.com",
		RSSEndPoints: []string{
			"feed/tag/ai/latest/rss",
			"feed/rss",
			"feed/category/science/latest/rss",
			"feed/category/backchannel/latest/rss",
			"feed/category/ideas/latest/rss",
			"feed/category/security/latest/rss",
			"feed/tag/wired-guide/latest/rss",
		},
		ParseConfig: ParseConfig{
			TitleSelector:    "title",
			ContentSelector:  "[class*='ArticlePageChunks']",
			SubtitleSelector: "div.[data-testid='ContentHeaderHed']",
			DateSelector:     "div.dateTime",
			ContentFilter: ContentFilter{
				RemoveSelectors: []string{"div.article__body table", "div.article__body div.container--body-inner"},
			},
		},
	}
}

func (a *Article) Fetch() ([]models.ArticleModel, error) {
	parser := gofeed.NewParser()

	var (
		articles []models.ArticleModel
		mu       sync.Mutex
		wg       sync.WaitGroup
		errChan  = make(chan error, len(a.RSSEndPoints))
	)

	concurrencyLimit := 5
	sem := make(chan struct{}, concurrencyLimit)

	for _, feedURL := range a.RSSEndPoints {
		wg.Add(1)

		go func(url string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			articleKywords := []string{}
			feed, err := parser.ParseURL(fmt.Sprintf("%s/%s", a.BaseURL, url))
			if err != nil {
				errChan <- fmt.Errorf("failed to parse RSS feed %s: %w", url, err)
				return
			}

			for _, item := range feed.Items {

				media, ok := item.Extensions["media"]
				if !ok {
					fmt.Println("No media extension found in feed")
					break
				}

				keywords, ok := media["keywords"]

				if !ok {
					fmt.Println("No keywords found in media extension")
					break
				}

				for _, keywordExt := range keywords {
					articleKywords = append(articleKywords, keywordExt.Value)
				}

				body, err := a.Parse(item.Link)

				if err != nil {
					errChan <- fmt.Errorf("failed to parse article %s: %w", item.Link, err)
					return
				}
				article := models.ArticleModel{
					Title:       item.Title,
					ID: item.GUID,
					Description: item.Description,
					Link:        item.Link,
					PublishedAt: item.PublishedParsed.Format("2006-01-02T15:04:05Z"),
					Categories:  item.Categories,
					Tags:    articleKywords,
					Body:        body,
				}
				mu.Lock()
				articles = append(articles, article)
				mu.Unlock()
				time.Sleep(1000 * time.Millisecond) // Throttle requests to avoid overwhelming the server	
			}
		}(feedURL)

	}

	wg.Wait()
	close(errChan)

	// If any errors occurred, return the first one
	for err := range errChan {
		return nil, err
	}

	return articles, nil
}

func (a *Article) applyContentFilters(doc *goquery.Document) {
	filter := a.ParseConfig.ContentFilter

	for _, selector := range filter.RemoveAfterSelectors {
		doc.Find(selector).NextAll().Remove()
	}

	for _, selector := range filter.RemoveSelectors {
		doc.Find(selector).Remove()
	}

	for _, selector := range filter.RemoveAfterSelectors {
		doc.Find(selector).NextAll().Remove()
	}
}

func (a *Article) Parse(url string) (string, error) {
	doc, err := a.fetchAndParse(url)
	if err != nil {
		return "", err
	}

	a.applyContentFilters(doc)

	var body strings.Builder
	doc.Find(a.ParseConfig.ContentSelector).First().Children().Each(func(j int, el *goquery.Selection) {
		body.WriteString(formatter.FormatNode(el))
	})

	title := strings.TrimSpace(doc.Find(a.ParseConfig.TitleSelector).Text())

	if a.ParseConfig.TitleProcessor != nil {
		title = a.ParseConfig.TitleProcessor(title)
	}

	verticalLine := strings.Index(title, "|")

	if verticalLine > 0 {
		title = title[:verticalLine]
	}

	return body.String(), nil

}

func (a *Article) fetchAndParse(url string) (*goquery.Document, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching article: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error fetching article: received status code %d", resp.StatusCode)
	}

	return goquery.NewDocumentFromReader(resp.Body)
}
