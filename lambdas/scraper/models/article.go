package models

type ArticleModel struct {
	ID 				string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Link        string `json:"link"`
	Body     string `json:"body"`
	PublishedAt string `json:"publishedAt"`
	Categories	[]string `json:"categories,omitempty"`
	Tags []string `json:"tags,omitempty"`
}

