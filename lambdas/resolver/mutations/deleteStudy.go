package mutations

import (
	"context"
	"fmt"
	"github.com/ankit-lilly/dtd-go-backend/internal/neptunedb"
	"log"
)

func HandleMutationDeleteStudy(ctx context.Context, args map[string]interface{}) (bool, error) {
	studyID, ok := args["id"].(string)
	if !ok || studyID == "" {
		return false, fmt.Errorf("study ID is required for deletion")
	}

	query := `
	MATCH (s:Study {id: $id})
	OPTIONAL MATCH (s)-[*]->(descendant)
	DETACH DELETE s, descendant
	`
	params := map[string]interface{}{"id": studyID}

	err := neptunedb.ExecuteWriteQuery(ctx, query, params)
	if err != nil {
		log.Printf("Error deleting study %s: %v", studyID, err)
		return false, err
	}

	log.Printf("Successfully deleted study %s and its descendants", studyID)
	return true, nil
}
