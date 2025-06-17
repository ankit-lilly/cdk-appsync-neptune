### 1. Find Existing Articles
```go
g.V().Has("Article", "link", article.Link)
```
- `g.V()` - Start with ALL vertices in the graph
- `.Has("Article", "link", article.Link)` - Filter to vertices that:
  - Have label "Article" AND
  - Have property "link" equal to `article.Link`

**Result:** Either 0 or 1 vertex (assuming links are unique)

### 2. Fold the Results
```go
.Fold()
```
- Takes the stream of results and puts them into a single list
- If article exists: `[ArticleVertex]`
- If article doesn't exist: `[]` (empty list)

**Why fold?** The `Coalesce` step needs to work with a single value, not a stream.

### 3. The Coalesce Decision
```go
.Coalesce(
    gremlingo.T__.Unfold(),           // Option A: If list is NOT empty
    gremlingo.T__.AddV("Article")...  // Option B: If list IS empty
)
```

`Coalesce` is like an "if-else" statement:
- It tries Option A first
- If Option A produces no results, it tries Option B
- It returns the first option that produces results

### 4. Option A: Article EXISTS
```go
gremlingo.T__.Unfold()
```
- `Unfold()` takes the list `[ArticleVertex]` and extracts the vertex
- Returns the existing article vertex
- **No new vertex is created**

### 5. Option B: Article DOESN'T EXIST
```go
gremlingo.T__.AddV("Article").
    Property("id", article.ID).
    Property("title", article.Title).
    Property("description", article.Description).
    Property("body", article.Body).
    Property("link", article.Link).
    Property("publishedAt", article.PublishedAt)
```
- `AddV("Article")` creates a new vertex with label "Article"
- Each `.Property()` adds a property to the new vertex
- Returns the newly created vertex

### 6. Execute and Get Results
```go
.Next()
```
- Executes the traversal and gets the first result
- Returns either the existing vertex or the newly created vertex

```go
articleVertex, err := articleVertexResult.GetVertex()
```
- Converts the result into a proper Vertex object that we can use

## Visual Example

Let's say we have these scenarios:

### Scenario 1: Article Already Exists
```
Graph before: [Article: link="example.com/news1"]

Query execution:
1. g.V().Has("Article", "link", "example.com/news1") → [ExistingVertex]
2. Fold() → [[ExistingVertex]]
3. Coalesce tries Unfold() → SUCCESS → ExistingVertex
4. Returns ExistingVertex (no new vertex created)

Graph after: [Article: link="example.com/news1"] (unchanged)
```

### Scenario 2: Article Doesn't Exist
```
Graph before: [Article: link="example.com/other-news"]

Query execution:
1. g.V().Has("Article", "link", "example.com/news1") → []
2. Fold() → [[]]
3. Coalesce tries Unfold() → FAILS (empty list)
4. Coalesce tries AddV("Article")... → SUCCESS → NewVertex
5. Returns NewVertex

Graph after: [Article: link="example.com/other-news"], [Article: link="example.com/news1"]
```

## WHy?

### 1. **Atomic Operation**
- The entire check-and-create happens in one database operation
- No race conditions between checking and creating

### 2. **Efficient**
- Single network round-trip to the database
- No need for separate "check if exists" + "create if not" operations

### 3. **Idempotent**
- Running the same query multiple times produces the same result
- Safe to retry on failures




## Comparison with Cypher:


```cypher
MERGE (article:Article {link: $link})
ON CREATE SET
    article.id = $id,
    article.title = $title,
    article.description = $description,
    article.body = $body,
    article.publishedAt = $publishedAt,
    article.createdAt = datetime()
ON MATCH SET
    article.lastUpdated = datetime()
```

This basically does the same thing as the Gremlin query above. It checks if an Article with the given link exists, and if not, it creates a new Article node 
with the provided properties. If it exists, it just updates the `lastUpdated` property.
