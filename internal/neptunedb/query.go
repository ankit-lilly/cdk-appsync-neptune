package  neptunedb

import (
	"context"
	"log"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)


func ExecuteReadQuery(ctx context.Context, query string, params map[string]interface{}) ([]*neo4j.Record, error) {
	driver := GetDriver()
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		log.Printf("Executing Query: %s\nWith Params: %+v", query, params)
		res, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}
		return res.Collect(ctx)
	})
	if err != nil {
		return nil, err
	}
	return result.([]*neo4j.Record), nil
}

func ExecuteWriteQuery(ctx context.Context, query string, params map[string]interface{}) error {
	driver := GetDriver()
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		log.Printf("Executing Write Query: %s\nWith Params: %+v", query, params)
		res, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}
		return res.Consume(ctx)
	})
	return err
}
