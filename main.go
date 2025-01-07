package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	notion "github.com/dstotijn/go-notion"
	"github.com/gorilla/mux"
)

type CMSData struct {
	Title    string `json:"title"`
	Category string `json:"category"`
	Slug     string `json:"slug"`
	Content  string `json:"content"`
}

func fetchNotionData(apiKey, databaseID string, limit int) ([]CMSData, error) {
	client := notion.NewClient(apiKey)

	query := notion.DatabaseQuery{
		Filter: &notion.DatabaseQueryFilter{
			Property: "Published",
			DatabaseQueryPropertyFilter: notion.DatabaseQueryPropertyFilter{
				Checkbox: &notion.CheckboxDatabaseQueryFilter{
					Equals: notion.BoolPtr(true),
				},
			},
		},
		PageSize: limit,
	}

	res, err := client.QueryDatabase(context.Background(), databaseID, &query)
	if err != nil {
		return nil, err
	}

	var data []CMSData
	for _, page := range res.Results {
		properties, ok := page.Properties.(map[string]notion.DatabasePageProperty)
		if !ok {
			return nil, fmt.Errorf("invalid page properties")
		}

		title := properties["Title"].Title[0].Text.Content
		category := properties["Category"].Select.Name
		slug := properties["Slug"].RichText[0].Text.Content
		content := properties["Content"].RichText[0].Text.Content

		data = append(data, CMSData{
			Title:    title,
			Category: category,
			Slug:     slug,
			Content:  content,
		})
	}

	return data, nil
}

func cmsHandler(w http.ResponseWriter, r *http.Request) {
	apiKey := os.Getenv("NOTION_API_KEY")
	databaseID := os.Getenv("NOTION_DATABASE_ID")

	// Limit the number of entries (e.g., top 20 items)
	limit := 20
	data, err := fetchNotionData(apiKey, databaseID, limit)
	if err != nil {
		http.Error(w, "Failed to fetch Notion data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func articleHandler(w http.ResponseWriter, r *http.Request) {
	apiKey := os.Getenv("NOTION_API_KEY")
	databaseID := os.Getenv("NOTION_DATABASE_ID")
	params := mux.Vars(r)

	category := params["category"]
	slug := params["slug"]

	client := notion.NewClient(apiKey)

	query := notion.DatabaseQuery{
		Filter: &notion.DatabaseQueryFilter{
			And: []notion.DatabaseQueryFilter{
				{
					Property: "Category",
					DatabaseQueryPropertyFilter: notion.DatabaseQueryPropertyFilter{
						Select: &notion.SelectDatabaseQueryFilter{
							Equals: category,
						},
					},
				},
				{
					Property: "Slug",
					DatabaseQueryPropertyFilter: notion.DatabaseQueryPropertyFilter{
						RichText: &notion.TextPropertyFilter{
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
	properties := page.Properties.(map[string]notion.DatabasePageProperty)
	title := properties["Title"].Title[0].Text.Content
	content := properties["Content"].RichText[0].Text.Content

	data := CMSData{
		Title:   title,
		Slug:    slug,
		Content: content,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// Serve static files
func staticFileHandler(staticDir string) http.Handler {
	return http.StripPrefix("/", http.FileServer(http.Dir(staticDir)))
}

func main() {
	r := mux.NewRouter()

	// Static files handler
	staticDir := "./static"
	r.PathPrefix("/").Handler(staticFileHandler(staticDir))

	// API handlers
	r.HandleFunc("/cms", cmsHandler).Methods("GET")
	r.HandleFunc("/cms/{category}/{slug}", articleHandler).Methods("GET")

	fmt.Println("Server is running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
