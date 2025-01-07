package notion

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	gn "github.com/dstotijn/go-notion"
	"github.com/gorilla/mux"
)

type CustomPage struct {
	ID             string       `json:"id"`
	CreatedTime    time.Time    `json:"created_time"`
	CreatedBy      *gn.BaseUser `json:"created_by,omitempty"`
	LastEditedTime time.Time    `json:"last_edited_time"`
	LastEditedBy   *gn.BaseUser `json:"last_edited_by,omitempty"`
	Parent         gn.Parent    `json:"parent"`
	Archived       bool         `json:"archived"`
	URL            string       `json:"url"`
	Icon           *gn.Icon     `json:"icon,omitempty"`
	Cover          *gn.Cover    `json:"cover,omitempty"`

	Properties map[string]gn.DatabasePageProperty `json:"properties"`
}

func FetchCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	apiKey := os.Getenv("NOTION_API_KEY")
	databaseID := os.Getenv("NOTION_DATABASE_ID")

	client := gn.NewClient(apiKey)
	database, err := client.FindDatabaseByID(context.Background(), databaseID)
	if err != nil {
		log.Println(err)
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

func FetchArticlesHandler(w http.ResponseWriter, r *http.Request) {
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
		Sorts: []gn.DatabaseQuerySort{
			{
				Property:  "CreatedAt",
				Direction: gn.SortDirDesc,
			},
		},
		PageSize: 20,
	}

	res, err := client.QueryDatabase(context.Background(), databaseID, &query)
	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to fetch Notion data", http.StatusInternalServerError)
		return
	}

	var pages []CustomPage
	for _, page := range res.Results {
		var customPage CustomPage
		jsonData, err := json.Marshal(page)
		if err != nil {
			log.Println(err)
			continue
		}
		json.Unmarshal(jsonData, &customPage)
		pages = append(pages, customPage)
	}

	var html string
	html += "<ul>"
	for _, page := range pages {
		title := page.Properties["Title"].Title[0].Text.Content
		slug := page.Properties["Slug"].RichText[0].Text.Content
		category := page.Properties["Category"].Select.Name
		html += fmt.Sprintf(
			`<li><a href="#" hx-get="/cms/%s/%s" hx-target="#content" hx-push-url="/cms/%s/%s">%s</a></li>`,
			category, slug, category, slug, title,
		)
	}
	html += "</ul>"

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
		log.Printf("Failed to fetch Notion data: %v", err)
		http.Error(w, "Article not found", http.StatusNotFound)
		return
	}

	page := res.Results[0]
	pageID := page.ID

	// ページブロック（本文）を取得
	children, err := client.FindBlockChildrenByID(context.Background(), pageID, nil)
	if err != nil {
		log.Printf("Failed to fetch Notion data: %v", err)
		http.Error(w, "Failed to fetch blog content", http.StatusInternalServerError)
		return
	}

	// ブロック内容をHTMLに変換
	contentHTML := ""
	for _, block := range children.Results {
		content := ProcessBlock(block) // utils.go
		contentHTML += content
	}

	// NOTE: page.Properties["Title"]だとinterface{}型で返ってくるため、CustomPage型に変換する
	var customPage CustomPage
	jsonData, err := json.Marshal(page)
	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to fetch Notion data", http.StatusInternalServerError)
		return
	}
	json.Unmarshal(jsonData, &customPage)
	title := customPage.Properties["Title"].Title[0].Text.Content

	// HTML を生成
	html := fmt.Sprintf(`
		<article>
			<h1>%s</h1>
			<p><strong>Category:</strong> %s</p>
			<div>%s</div>
			<a href="#" hx-get="/cms" hx-target="#content" hx-push-url="/cms">Back to Articles</a>
		</article>
	`, title, category, contentHTML)

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}
