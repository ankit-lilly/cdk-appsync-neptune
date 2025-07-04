## What

 This is a sample GraphQL API using AWS AppSync and AWS Lambda. The API is built using AWS CDK
 and uses NeptuneDB for data storage.

 This is just meant to explore the capabilities of AWS AppSync, and Neptune DB.



## How it works


 It contains a two lamdba functions:


 - Scraper: This gets triggered by a scheduled event every few hours and scrapes the news from a source and stores it in neptune DB.
 It creates three Vertexes: `Article`, `Tags`, and `Category`


 - Resolver: This is the resolver for the AppSync API. It queries the Nepturne DB and returns the data in format that is expected by the GraphQL API.


To return all the artilces:

 ```gql

query GetFeed {
  feed(limit: 10, offset: 100) {
    id
    title
    description
    tags
    link
    categories
    publishedAt
  }
}

 ```


 To return articles related to a specific article ( this is meant to show how relationships work in the graph) we pass article ID


```gql
query Related {
  related (articleId:"https://www.wired.com/story/openworm-worm-simulator-biology-code/" limit:10) {
    id
    title
    tags
    categories
  }
}

```


## TODO and Limitations:

- There's no authorization in place. The authentication works via   AppSync api key.

- There's no UI so the graphl queries have to be run via AppSync console or any client that supports graphql ( like Bruno or  postman or just plain old curl  post request)

- It doesn't use transactions while inserting data into Neptune DB. For instance, we create the Article and then create the tags and categories if the second or third step 
fails the Article will still exist in database but it will not have any tags or categories associated with it.


## Building And Deploying


```sh
make build
```

It should build lambda functions and create a binary named 'bootstrap' inside each lambda function directory and then put it in a zip file `function.zip`

And then deploy the CDK stack:


```
cdk deploy

or 

npx aws-cdk deploy --profile <your-aws-profile>
```

Note: if it doesn't work then you may need to specify AccountId inside Env inside aws.CdkProps
