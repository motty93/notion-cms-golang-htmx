package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/motty93/notion-cms-golang-htmx/internal/notion"
)

// Serve static files
func staticFileHandler(staticDir string) http.Handler {
	return http.StripPrefix("/", http.FileServer(http.Dir(staticDir)))
}

func main() {
	r := mux.NewRouter()

	// API handlers
	r.HandleFunc("/cms", notion.FetchArticlesHandler).Methods("GET")
	r.HandleFunc("/cms/categories", notion.FetchCategoriesHandler).Methods("GET")
	r.HandleFunc("/cms/{category}/{slug}", notion.FetchArticleHandler).Methods("GET")

	// Static files handler
	staticFileDirectory := http.Dir("./static/")
	r.PathPrefix("/").Handler(staticFileHandler(string(staticFileDirectory))).Methods("GET")

	fmt.Println("Server is running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
