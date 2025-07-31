package gremlin

import (
	"os"
	"fmt"
	"log"
	"sync"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

var (
	writerGraphTraversalSource *gremlingo.GraphTraversalSource
	readerGraphTraversalSource *gremlingo.GraphTraversalSource
	writerConn                 *gremlingo.DriverRemoteConnection
	readerConn                 *gremlingo.DriverRemoteConnection
	once 										sync.Once
)

func GetWriterConn() *gremlingo.DriverRemoteConnection {
	once.Do(func() {

	neptuneWriterHostname := os.Getenv("NEPTUNE_ENDPOINT")
	neptunePort := os.Getenv("NEPTUNE_PORT")

	if neptuneWriterHostname == "" || neptunePort == "" {
		log.Fatal("NEPTUNE_ENDPOINT, NEPTUNE_PORT environment variables must be set.")
	}

	writerConnStr := fmt.Sprintf("wss://%s:%s/gremlin", neptuneWriterHostname, neptunePort)

	log.Printf("DEBUG (Resolver Init): Constructed Writer Connection String: '%s'", writerConnStr)

	var err error
	writerConn, err = gremlingo.NewDriverRemoteConnection(writerConnStr)
	if err != nil {
		log.Fatalf("ERROR (Resolver Init): Failed to create writer connection with string '%s': %v", writerConnStr, err)
	}
	})
	return writerConn
}

func GetReaderConn() *gremlingo.DriverRemoteConnection {
	once.Do(func() {
		neptuneReaderHostname := os.Getenv("NEPTUNE_READER_ENDPOINT")
		neptunePort := os.Getenv("NEPTUNE_PORT")
		if neptuneReaderHostname == "" || neptunePort == "" {
			log.Fatal("NEPTUNE_READER_ENDPOINT, NEPTUNE_PORT environment variables must be set.")
		}
		readerConnStr := fmt.Sprintf("wss://%s:%s/gremlin", neptuneReaderHostname, neptunePort)
		log.Printf("DEBUG (Resolver Init): Constructed Reader Connection String: '%s'", readerConnStr)
		var err error
		readerConn, err = gremlingo.NewDriverRemoteConnection(readerConnStr)
		if err != nil {
			log.Fatalf("ERROR (Resolver Init): Failed to create reader connection with string '%s': %v", readerConnStr, err)
		}
		log.Println("DEBUG (Resolver Init): Successfully created reader connection")
	})
	return readerConn
}

func GetWriterGraphTraversalSource() *gremlingo.GraphTraversalSource {
	writerConn := GetWriterConn()
	if writerConn == nil {
		log.Fatal("Writer connection is not initialized")
	}

	writerGraphTraversalSource := gremlingo.Traversal_().WithRemote(writerConn)
	return writerGraphTraversalSource
}

func GetReaderGraphTraversalSource() *gremlingo.GraphTraversalSource {
	readerConn := GetReaderConn()
	if readerConn == nil {
		log.Fatal("Reader connection is not initialized")
	}

	readerGraphTraversalSource := gremlingo.Traversal_().WithRemote(readerConn)
	return readerGraphTraversalSource
}


func CloseDriver() {
	writerConn := GetWriterConn()
	if writerConn != nil {
		writerConn.Close()
	}
	readerConn := GetReaderConn()
	if readerConn != nil {
		readerConn.Close()
	}
}

