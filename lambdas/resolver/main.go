package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"

	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
	"github.com/aws/aws-lambda-go/lambda"
)

var (
	writerGraphTraversalSource *gremlingo.GraphTraversalSource
	readerGraphTraversalSource *gremlingo.GraphTraversalSource
	writerConn                 *gremlingo.DriverRemoteConnection
	readerConn                 *gremlingo.DriverRemoteConnection
)

type GraphQLArticle struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description *string  `json:"description,omitempty"`
	Body        *string  `json:"body,omitempty"`
	Link        string   `json:"link"`
	PublishedAt *string  `json:"publishedAt,omitempty"`
	Categories  []string `json:"categories"`
	Tags        []string `json:"tags"`
}

type AppSyncEvent struct {
	Info      AppSyncInfo            `json:"info"`
	Arguments map[string]interface{} `json:"arguments"`
	Source    map[string]interface{} `json:"source"`
}

type AppSyncInfo struct {
	FieldName        string   `json:"fieldName"`
	ParentTypeName   string   `json:"parentTypeName"`
	SelectionSetList []string `json:"selectionSetList"`
}


func init() {
	neptuneWriterHostname := os.Getenv("NEPTUNE_ENDPOINT")
	neptuneReaderHostname := os.Getenv("NEPTUNE_READER_ENDPOINT")
	neptunePort := os.Getenv("NEPTUNE_PORT")

	if neptuneWriterHostname == "" || neptuneReaderHostname == "" || neptunePort == "" {
		log.Fatal("NEPTUNE_ENDPOINT, NEPTUNE_READER_ENDPOINT, and NEPTUNE_PORT environment variables must be set.")
	}

	writerConnStr := fmt.Sprintf("wss://%s:%s/gremlin", neptuneWriterHostname, neptunePort)
	readerConnStr := fmt.Sprintf("wss://%s:%s/gremlin", neptuneReaderHostname, neptunePort)

	log.Printf("DEBUG (Resolver Init): Constructed Writer Connection String: '%s'", writerConnStr)
	log.Printf("DEBUG (Resolver Init): Constructed Reader Connection String: '%s'", readerConnStr)

	var err error
	writerConn, err = gremlingo.NewDriverRemoteConnection(writerConnStr)
	if err != nil {
		log.Fatalf("ERROR (Resolver Init): Failed to create writer connection with string '%s': %v", writerConnStr, err)
	}
	writerGraphTraversalSource = gremlingo.Traversal_().WithRemote(writerConn)
	log.Println("Neptune writer connection initialized.")

	readerConn, err = gremlingo.NewDriverRemoteConnection(readerConnStr)
	if err != nil {
		log.Fatalf("ERROR (Resolver Init): Failed to create reader connection with string '%s': %v", readerConnStr, err)
	}
	readerGraphTraversalSource = gremlingo.Traversal_().WithRemote(readerConn)
	log.Println("Neptune reader connection initialized.")
}

func mapGremlinValueMapToGraphQLArticle(result *gremlingo.Result) (*GraphQLArticle, error) {
	if result == nil {
		return nil, fmt.Errorf("nil Gremlin valueMap result")
	}

	valMap, ok := result.Data.(map[interface{}]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to cast result.Data to map[interface{}]interface{} - got type %T", result.Data)
	}

	article := &GraphQLArticle{}

	getStringValue := func(key string) string {
		if val, found := valMap[key]; found {
			if vals, isSlice := val.([]interface{}); isSlice && len(vals) > 0 {
				if str, isStr := vals[0].(string); isStr {
					return str
				}
			}
		}
		return ""
	}

	getOptionalStringValue := func(key string) *string {
		if val, found := valMap[key]; found {
			if vals, isSlice := val.([]interface{}); isSlice && len(vals) > 0 {
				if str, isStr := vals[0].(string); isStr {
					return &str
				}
			}
		}
		return nil
	}

	if idVal, ok := valMap[gremlingo.T.Id]; ok {
		if strID, isStr := idVal.(string); isStr {
			article.ID = strID
		} else {
			log.Printf("Warning: T.Id is not a string for article: %T %v", idVal, idVal)
		}
	}
	article.ID = getStringValue("id")
	article.Link = getStringValue("link")
	if article.ID == "" && article.Link != "" {
		article.ID = article.ID
	}

	article.Title = getStringValue("title")
	article.Description = getOptionalStringValue("description")
	article.Body = getOptionalStringValue("body")
	article.PublishedAt = getOptionalStringValue("publishedAt")

	articleLink := getStringValue("link")
	if articleLink == "" {
		return nil, fmt.Errorf("article link property missing in valueMap result, cannot fetch edges")
	}

	categoriesRes, err := readerGraphTraversalSource.V().Has("Article", "link", articleLink).Out("HAS_CATEGORY").Values("name").ToList()
	if err != nil {
		log.Printf("Warning: Failed to fetch categories for article %s: %v", article.Link, err)
		article.Categories = []string{}
	} else {
		fmt.Printf("DEBUG: Fetched %d categories for article %s\n", len(categoriesRes), categoriesRes)
		cats := make([]string, len(categoriesRes))
		for i, r := range categoriesRes {
			cats[i] = r.GetString()
		}
		article.Categories = cats
	}

	keywordsRes, err := readerGraphTraversalSource.V().Has("Article", "link", articleLink).Out("HAS_TAG").Values("name").ToList()
	if err != nil {
		log.Printf("Warning: Failed to fetch tags (keywords) for article %s: %v", article.Link, err)
		article.Tags = []string{}
	} else {
		fmt.Printf("DEBUG: Fetched %d tags for article %s\n", len(keywordsRes), keywordsRes)
		kws := make([]string, len(keywordsRes))
		for i, r := range keywordsRes {
			kws[i] = r.GetString()
		}
		article.Tags = kws
	}

	return article, nil
}

func handleQueryFeed(args map[string]interface{}) ([]*GraphQLArticle, error) {
	g := readerGraphTraversalSource

	limit := int64(10)
	if val, ok := args["limit"]; ok {
		if floatVal, isFloat := val.(float64); isFloat {
			limit = int64(floatVal)
		}
	}

	offset := int64(0)
	if val, ok := args["offset"]; ok {
		if floatVal, isFloat := val.(float64); isFloat {
			offset = int64(floatVal)
		}
	}

	results, err := g.V().HasLabel("Article").
		Order().By("publishedAt", gremlingo.Order.Desc).
		Range(offset, offset+limit).
		ValueMap(true).
		ToList()

	if err != nil {
		return nil, fmt.Errorf("failed to fetch feed articles: %w", err)
	}

	articles := make([]*GraphQLArticle, 0, len(results))
	for _, res := range results {
		article, mapErr := mapGremlinValueMapToGraphQLArticle(res)
		if mapErr != nil {
			log.Printf("Warning: Failed to map article valueMap result to GraphQLArticle: %v", mapErr)
			continue
		}
		articles = append(articles, article)
	}
	return articles, nil
}


func buildArticleProjection(selectionSet []string)  *gremlingo.GraphTraversal {
	projection := gremlingo.T__.Project("article").By(gremlingo.T__.ValueMap(true))

	hasField := func(field string) bool {
		if slices.Contains(selectionSet, field) {
			return true
		}
		return false
	}

	if hasField("categories") {
		projection = projection.By("categories", gremlingo.T__.Out("HAS_CATEGORY").Values("name"))
	}

	if hasField("tags") {
		projection = projection.By("tags", gremlingo.T__.Out("HAS_TAG").Values("name"))
	}

	return projection
}

func handleQueryArticle(args map[string]interface{}) (*GraphQLArticle, error) {
	g := readerGraphTraversalSource

	articleID, ok := args["id"].(string)
	if !ok || articleID == "" {
		return nil, fmt.Errorf("article ID is required")
	}

	result, err := g.V().Has("Article", "link", articleID).ValueMap(true).Next()
	if err != nil {
		if strings.Contains(err.Error(), "No next value") {
			return nil, nil // Article not found, return nil without error
		}
		return nil, fmt.Errorf("failed to fetch article %s: %w", articleID, err)
	}

	article, mapErr := mapGremlinValueMapToGraphQLArticle(result)
	if mapErr != nil {
		return nil, fmt.Errorf("failed to map article valueMap result to GraphQLArticle: %v", mapErr)
	}
	return article, nil
}

func handleQueryRelated(args map[string]interface{}) ([]*GraphQLArticle, error) {
	g := readerGraphTraversalSource

	articleID, ok := args["articleId"].(string)
	if !ok || articleID == "" {
		return nil, fmt.Errorf("articleId is required for related query")
	}

	limit := int64(10)
	if val, ok := args["limit"]; ok {
		if floatVal, isFloat := val.(float64); isFloat {
			limit = int64(floatVal)
		}
	}

	results, err := g.V().Has("Article", "link", articleID).As("originalArticle").
		Out("HAS_CATEGORY", "HAS_TAG").As("sharedTopic").
		In("HAS_CATEGORY", "HAS_TAG").
		Dedup().
		Limit(limit).
		ValueMap(true).
		ToList()

	if err != nil {
		return nil, fmt.Errorf("failed to fetch related articles for %s: %w", articleID, err)
	}

	articles := make([]*GraphQLArticle, 0, len(results))
	for _, res := range results {
		article, mapErr := mapGremlinValueMapToGraphQLArticle(res)
		if mapErr != nil {
			log.Printf("Warning: Failed to map related article valueMap result to GraphQLArticle: %v", mapErr)
			continue
		}
		articles = append(articles, article)
	}
	return articles, nil
}

func handleQueryRecommended(args map[string]interface{}) ([]*GraphQLArticle, error) {
	g := readerGraphTraversalSource

	userID, ok := args["userId"].(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("userId is required for recommended query")
	}

	limit := int64(10)
	if val, ok := args["limit"]; ok {
		if floatVal, isFloat := val.(float64); isFloat {
			limit = int64(floatVal)
		}
	}

	results, err := g.V().Has("User", "id", userID).
		Out("FAVORITED").As("favoritedArticle").
		Out("HAS_CATEGORY", "HAS_TAG").As("sharedTopic").
		In("HAS_CATEGORY", "HAS_TAG").
		Dedup().
		Limit(limit).
		ValueMap(true).
		ToList()

	if err != nil {
		return nil, fmt.Errorf("failed to fetch recommended articles for user %s: %w", userID, err)
	}

	articles := make([]*GraphQLArticle, 0, len(results))
	for _, res := range results {
		article, mapErr := mapGremlinValueMapToGraphQLArticle(res)
		if mapErr != nil {
			log.Printf("Warning: Failed to map recommended article valueMap result to GraphQLArticle: %v", mapErr)
			continue
		}
		articles = append(articles, article)
	}
	return articles, nil
}

func handleMutationFavorite(args map[string]interface{}) (bool, error) {
	return false, fmt.Errorf("Mutation favorite not implemented yet")
}

func handler(ctx context.Context, event AppSyncEvent) (interface{}, error) {
	log.Printf("Received AppSync event: TypeName=%s, FieldName=%s, Arguments=%+v", event.Info.ParentTypeName, event.Info.FieldName, event.Arguments)

	if writerGraphTraversalSource == nil || readerGraphTraversalSource == nil {
		log.Print("Gremlin traversal sources not initialized. Attempting re-init.")
		if writerGraphTraversalSource == nil || readerGraphTraversalSource == nil {
			return nil, fmt.Errorf("Gremlin traversal sources failed to initialize.")
		}
	}

	switch event.Info.ParentTypeName {
	case "Query":
		switch event.Info.FieldName {
		case "study":
			return handleQueryFeed(event.Arguments)
		case "studies":
			return handleQueryArticle(event.Arguments)
		case "studyVersion":
			return handleQueryRelated(event.Arguments)
		case "organization":
			return handleQueryRecommended(event.Arguments)
		default:
			return nil, fmt.Errorf("unknown query field: %s", event.Info.FieldName)
		}
	case "Mutation":
		switch event.Info.FieldName {
		case "favorite":
			return handleMutationFavorite(event.Arguments)
		default:
			return nil, fmt.Errorf("unknown mutation field: %s", event.Info.FieldName)
		}
	default:
		return nil, fmt.Errorf("unknown AppSync type: %s", event.Info.ParentTypeName)
	}
}

func main() {
	lambda.Start(handler)
}
