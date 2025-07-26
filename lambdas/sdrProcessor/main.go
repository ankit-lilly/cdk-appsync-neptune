package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sdrProcessor/models"

	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
	"github.com/aws/aws-lambda-go/events"
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

func parseStudyData(data string) (models.Study, error) {
	var study models.Study
	err := json.Unmarshal([]byte(data), &study)
	if err != nil {
		return models.Study{}, fmt.Errorf("failed to parse study data: %w", err)
	}
	return study, nil
}

func handler(ctx context.Context, event events.SQSEvent) error {
	defer cleanup()
	for _, message := range event.Records {
		log.Printf("Processing message ID: %s", message.MessageId)
		log.Printf("Message body: %s", message.Body)

		study, err := parseStudyData(message.Body)
		if err != nil {
			log.Printf("Error parsing study data: %v", err)
			continue
		}
		log.Printf("Parsed study: %+v", study)

		processStudy(study)

	}
	return nil
}

func processStudy(study models.Study) {

	std, err := g.V().Has("study", "studyId", study.ID).
		Fold().
		Coalesce(
			gremlingo.T__.Unfold(),
			g.AddV("study").
				Property("studyId", study.ID).
				Property("name", study.Name).
				Property("description", study.Description).
				Property("label", study.Label),
		).Next()

	if err != nil {
		log.Printf("Error upserting study %s: %v", study.ID, err)
		return
	}

	log.Printf("Upserted study: %v", std)

	for _, version := range study.Versions {
		versionResult, err := g.V().Has("studyVersion", "versionIdentifier", version.VersionIdentifier).
			Fold().
			Coalesce(
				gremlingo.T__.Unfold(),
				gremlingo.T__.AddV("studyVersion").
					Property("versionIdentifier", version.VersionIdentifier).
					Property("rationale", version.Rationale),
			).Next()

		if err != nil {
			log.Printf("Error upserting study version %s: %v", version.VersionIdentifier, err)
			continue
		}
		log.Printf("Upserted study version: %v", versionResult)
	}
}

func main() {
	lambda.Start(handler)
}
