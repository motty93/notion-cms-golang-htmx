package notion

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	gn "github.com/dstotijn/go-notion"
	"github.com/gorilla/mux"
)

func FetchCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	apiKey := os.Getenv("NOTION_API_KEY")
	databaseID := os.Getenv("NOTION_DATABASE_ID")

	client := gn.NewClient(apiKey)

	database, err := client.FindDatabaseByID(context.Background(), databaseID)
	if err != nil {
		http.Error(w, "Failed to fetch Notion data", http.StatusInternalServerError)
		return
	}

	// categories一覧を取得
	var categories []string
	for name, property := range database.Properties {
		if name == "Category" {
			if property.Select != nil {
				for _, option := range property.Select.Options {
					categories = append(categories, option.Name)
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

func FetchCMSDataHandler(w http.ResponseWriter, r *http.Request) {
	apiKey := os.Getenv("NOTION_API_KEY")
	databaseID := os.Getenv("NOTION_DATABASE_ID")

	client := gn.NewClient(apiKey)

	query := gn.DatabaseQuery{
		Filter: &gn.DatabaseQueryFilter{
			Property: "Published",
			DatabaseQueryPropertyFilter: gn.DatabaseQueryPropertyFilter{
				Checkbox: &gn.CheckboxDatabaseQueryFilter{
					Equals: gn.BoolPtr(true),
				},
			},
		},
		PageSize: 20,
	}

	res, err := client.QueryDatabase(context.Background(), databaseID, &query)
	if err != nil {
		http.Error(w, "Failed to fetch Notion data", http.StatusInternalServerError)
		return
	}

	var html string
	html += "<ul>"
	for _, page := range res.Results {
		properties, ok := page.Properties.(map[string]gn.DatabasePageProperty)
		if !ok {
			http.Error(w, "Failed to parse Notion data", http.StatusInternalServerError)
			return
		}

		title := properties["Title"].Title[0].Text.Content
		category := properties["Category"].Select.Name
		slug := properties["Slug"].RichText[0].Text.Content
		// content := properties["Content"].RichText[0].Text.Content

		html += fmt.Sprintf(
			`<li><a href="#" hx-get="/cms/%s/%s" hx-target="#content" hx-push-url="/cms/%s/%s">%s</a></li>`,
			category, slug, category, slug, title,
		)
		html += "</ul>"
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(html))
}

func FetchArticleHandler(w http.ResponseWriter, r *http.Request) {
	apiKey := os.Getenv("NOTION_API_KEY")
	databaseID := os.Getenv("NOTION_DATABASE_ID")

	client := gn.NewClient(apiKey)
	params := mux.Vars(r)
	category := params["category"]
	slug := params["slug"]

	query := gn.DatabaseQuery{
		Filter: &gn.DatabaseQueryFilter{
			And: []gn.DatabaseQueryFilter{
				{
					Property: "Category",
					DatabaseQueryPropertyFilter: gn.DatabaseQueryPropertyFilter{
						Select: &gn.SelectDatabaseQueryFilter{
							Equals: category,
						},
					},
				},
				{
					Property: "Slug",
					DatabaseQueryPropertyFilter: gn.DatabaseQueryPropertyFilter{
						RichText: &gn.TextPropertyFilter{
							Equals: slug,
						},
					},
				},
			},
		},
	}

	res, err := client.QueryDatabase(context.Background(), databaseID, &query)
	if err != nil || len(res.Results) == 0 {
		http.Error(w, "Article not found", http.StatusNotFound)
		return
	}

	page := res.Results[0]
	properties, ok := page.Properties.(map[string]gn.DatabasePageProperty)
	if !ok {
		http.Error(w, "Failed to parse Notion data", http.StatusInternalServerError)
		return
	}

	title := properties["Title"].Title[0].Text.Content
	content := properties["Content"].RichText[0].Text.Content

	// HTML を生成
	html := fmt.Sprintf(`
		<article>
			<h1>%s</h1>
			<p><strong>Category:</strong> %s</p>
			<div>%s</div>
			<a href="#" hx-get="/cms" hx-target="#content" hx-push-url="/cms">Back to Articles</a>
		</article>
	`, title, category, content)

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}
