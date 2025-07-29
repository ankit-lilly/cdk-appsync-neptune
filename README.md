## What

GraphQL API For SDR.


It exposes two APIs:

1. GraphQL API for querying the SDR ( neptune database)
2. A REST API for submitting new SDR.


## How it works

1. The end user/client can submit a new SDR via a REST POST request to the `/sdr` endpoint.
2. The request is received by the Amazon API Gateway, which triggers the SQS event and responds with a 202 status code.
3. The SDRProcessor Lambda function receives the message from the SQS quueue, parses the SDR data, and transforms it into a format suitable for storage in the 
Neptune database.

```shell
+-----------------------+
|   End User / Client   |
+-----------+-----------+
            |
+-------------------------------------------------------------------------+
| (REST POST to /sdr)                                             | (GraphQL Query)
v                                                                         v
+------------------------+                                +-----------------------------+
|  Amazon API Gateway    |                                |      AWS AppSync API        |
|  (REST API Endpoint)   |                                |    (GraphQL Endpoint)       |
+-----------+------------+                                +--------------+--------------+
            |                                                            |
            | 1. Triggers Writer Lambda                                  | 1. Invokes Resolver Lambda
            v                                                            v
+------------------------+                                +-----------------------------+
|     "SDRHandler" Lambda    |                                |   "Resolver" Lambda         |
|  (Receives & Validates)|                                | (Handles GraphQL queries)   |
+-----------+------------+                                +--------------+--------------+
            |                                                            | 2. Connects to Neptune
            | 2. Sends message to SQS                                    |    and reads data
            v                                                            |
+------------------------+                                             |
|      Amazon SQS        |                                             |
|        (Queue)         |                                             |
+-----------+------------+                                             |
            |                                                            |
            | 3. Triggers Processor Lambda                               |
            v                                                            |
+------------------------+                                             |
|   "SDRProcessor" Lambda   |                                             |
|  (Parses & Transforms) |                                             |
+-----------+------------+                                             |
            |                                                            |
            | 4. Connects to Neptune                                     |
            |    and writes data                                         |
            v                                                            v
+-----------------------------------------------------------------------------------+
|                                                                                   |
|                               Amazon Neptune DB                                   |
|                                (Graph Database)                                   |
|                                                                                   |
+-----------------------------------------------------------------------------------+

```


## Building and Deploying

The application uses Golang with CDK. The resources are defined within the stack/ directory.

The lambda functions are defined in the `lambda/` directory. To build and deploy the application the lambda functions must be built first.

```bash
make build
```

Then, you can use cdk deploy to deploy the stack:

```bash
cdk deploy 
```
or 

```bash
 npx aws-cdk bootstrap --profile la
```


If you are deploying for the first time, you may need to bootstrap your AWS environment:

```bash
npx aws-cdk bootstrap --profile <your-profile>
```
This command sets up the necessary resources for CDK to deploy your stack, such as an S3 bucket for storing assets.

