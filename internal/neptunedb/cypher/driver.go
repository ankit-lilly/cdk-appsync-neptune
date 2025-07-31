package cypher

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var (
	driver neo4j.DriverWithContext
	once   sync.Once
)

func GetDriver() neo4j.DriverWithContext {
	once.Do(func() {
		endpoint := os.Getenv("NEPTUNE_ENDPOINT")
		if endpoint == "" {
			log.Fatal("NEPTUNE_ENDPOINT must be set")
		}

		uri := fmt.Sprintf("bolt+s://%s:8182", endpoint)

		var err error
		driver, err = neo4j.NewDriverWithContext(uri, neo4j.NoAuth())
		if err != nil {
			log.Fatalf("Failed to connect to Neo4j: %v", err)
		}

		log.Println("Neo4j driver initialized")
	})
	return driver
}

func CloseDriver(ctx context.Context) {
	if driver != nil {
		if err := driver.Close(ctx); err != nil {
			log.Printf("Error closing driver: %v", err)
		}
	}
}

