module github.com/ankit-lilly/dtd-go-backend/lambdas/resolver

go 1.24.5

require (
	github.com/ankit-lilly/dtd-go-backend v0.0.0-00010101000000-000000000000
	github.com/aws/aws-lambda-go v1.49.0
	github.com/neo4j/neo4j-go-driver/v5 v5.28.1
)

require github.com/stretchr/testify v1.9.0 // indirect

replace github.com/ankit-lilly/dtd-go-backend => ../..
