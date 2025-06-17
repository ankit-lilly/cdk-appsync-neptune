package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"scraper/article"
	"scraper/models"
	"strings"

	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
	"github.com/aws/aws-lambda-go/lambda"
)

var (
	g    *gremlingo.GraphTraversalSource
	conn *gremlingo.DriverRemoteConnection
)

func init() {
	neptuneEndpoint := os.Getenv("NEPTUNE_ENDPOINT")
	neptunePort := os.Getenv("NEPTUNE_PORT")

	if neptuneEndpoint == "" || neptunePort == "" {
		log.Fatal("NEPTUNE_ENDPOINT and NEPTUNE_PORT environment variables must be set in Lambda configuration.")
	}

	connString := fmt.Sprintf("wss://%s:%s/gremlin", neptuneEndpoint, neptunePort)

	var err error
	conn, err = gremlingo.NewDriverRemoteConnection(connString)
	if err != nil {
		log.Fatalf("Failed to establish Gremlin connection: %v", err)
	}
	g = gremlingo.Traversal_().WithRemote(conn)

	log.Println("Neptune Gremlin connection established in Lambda init().")
}

func cleanup() {
	if conn != nil {
		conn.Close()
	}
}

func handler(ctx context.Context) error {
	defer cleanup()

	articleFetcher := article.NewArticle()
	articles, err := articleFetcher.Fetch()
	if err != nil {
		return fmt.Errorf("failed to fetch articles: %w", err)
	}

	for i, article := range articles {
		log.Printf("Processing article %d: %s", i+1, article.Title)

		articleVertex, err := upsertArticle(article)
		if err != nil {
			log.Printf("ERROR: Failed to upsert Article %s: %v", article.Title, err)
			continue
		}

		if err := processCategories(articleVertex, article.Categories, article.Title); err != nil {
			log.Printf("ERROR processing categories for article %s: %v", article.Title, err)
		}

		if err := processTags(articleVertex, article.Tags, article.Title); err != nil {
			log.Printf("ERROR processing tags for article %s: %v", article.Title, err)
		}
	}

	return nil
}

func upsertArticle(article models.ArticleModel) (*gremlingo.Vertex, error) {
	articleVertexResult, err := g.V().Has("Article", "link", article.Link).
		Fold().
		Coalesce(
			gremlingo.T__.Unfold(),
			gremlingo.T__.AddV("Article").
				Property("id", article.ID).
				Property("title", article.Title).
				Property("description", article.Description).
				Property("body", article.Body).
				Property("link", article.Link).
				Property("publishedAt", article.PublishedAt),
		).Next()

	if err != nil {
		return nil, fmt.Errorf("failed to upsert article: %w", err)
	}

	articleVertex, err := articleVertexResult.GetVertex()
	if err != nil {
		return nil, fmt.Errorf("failed to get article vertex: %w", err)
	}

	return articleVertex, nil
}

func processCategories(articleVertex *gremlingo.Vertex, categories []string, articleTitle string) error {
	for _, category := range categories {
		category = strings.TrimSpace(category)
		if category == "" {
			continue
		}

		// First, upsert the category vertex
		categoryVertexResult, err := g.V().Has("Category", "name", category).
			Fold().
			Coalesce(
				gremlingo.T__.Unfold(),
				gremlingo.T__.AddV("Category").Property("name", category),
			).Next()

		if err != nil {
			log.Printf("ERROR: Failed to upsert Category %s: %v", category, err)
			continue
		}

		categoryVertex, err := categoryVertexResult.GetVertex()
		if err != nil {
			log.Printf("ERROR: Failed to get Category vertex for %s: %v", category, err)
			continue
		}

		// Create edge if it doesn't exist
		_, err = g.V(articleVertex).
			Coalesce(
				gremlingo.T__.OutE("HAS_CATEGORY").Where(gremlingo.T__.InV().Is(categoryVertex)),
				gremlingo.T__.AddE("HAS_CATEGORY").To(categoryVertex),
			).Next()

		if err != nil {
			log.Printf("ERROR: Failed to create HAS_CATEGORY edge for Article %s and Category %s: %v", 
				articleTitle, category, err)
		}
	}
	return nil
}

func processTags(articleVertex *gremlingo.Vertex, tags []string, articleTitle string) error {
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}

		// upsert the tag vertex
		tagVertexResult, err := g.V().Has("Tag", "name", tag).
			Fold().
			Coalesce(
				gremlingo.T__.Unfold(),
				gremlingo.T__.AddV("Tag").Property("name", tag),
			).Next()

		if err != nil {
			log.Printf("ERROR: Failed to upsert Tag %s: %v", tag, err)
			continue
		}

		tagVertex, err := tagVertexResult.GetVertex()
		if err != nil {
			log.Printf("ERROR: Failed to get Tag vertex for %s: %v", tag, err)
			continue
		}

		// Create edge if it doesn't exist
		_, err = g.V(articleVertex).
			Coalesce(
				gremlingo.T__.OutE("HAS_TAG").Where(gremlingo.T__.InV().Is(tagVertex)),
				gremlingo.T__.AddE("HAS_TAG").To(tagVertex),
			).Next()

		if err != nil {
			log.Printf("ERROR: Failed to create HAS_TAG edge for Article %s and Tag %s: %v", 
				articleTitle, tag, err)
		}
	}
	return nil
}

func main() {
	lambda.Start(handler)
}
