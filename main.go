package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Tag is used for sub-menu on articles page
type Tag struct {
	Slug     string
	Category string
}

var (
	templates      map[string]*template.Template
	allPages       map[string]*Page
	allArticles    map[string]*Page
	sortedArticles []*Page
	authors        map[string]*Author
	categories     map[string]string
	tags           map[string]Tag
)

func sidebarOdd(index int, byAuthor bool) string {
	if index == 0 || index%2 == 0 {
		if byAuthor {
			return ""
		}
		return "1"
	}
	return "2"
}

func isFirst(index int) bool {
	return index == 0
}

func loadTemplates() {
	if templates == nil {
		templates = make(map[string]*template.Template)
	}

	layouts, err := filepath.Glob("views/layouts/*.hbs")
	if err != nil {
		log.Fatal(err)
	}

	pages, err := filepath.Glob("views/pages/*.hbs")
	if err != nil {
		log.Fatal(err)
	}

	for _, page := range pages {
		for _, layout := range layouts {
			t := template.New(page).Funcs(template.FuncMap{"sidebarOdd": sidebarOdd, "isFirst": isFirst})
			templates[filepath.Base(page)] = template.Must(t.ParseFiles(layout, page))
		}
	}
}

func render(w http.ResponseWriter, name string, data *pageData) (err error) {
	template, ok := templates[name]
	if !ok {
		err = fmt.Errorf("The template %s does not exists", name)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	template.ExecuteTemplate(w, "base", data)

	return nil
}

func cacheData() {
	err := openConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer closeConnection()

	if len(authors) == 0 {
		aa, err := getAuthors()
		if err != nil {
			log.Fatal(err)
		}

		for _, a := range aa {
			authors[a.AuthorName] = a
		}
	}

	// load all data
	ap, err := getArticles()
	if err != nil {
		log.Fatal(err)
	}

	sortedArticles = ap

	re := regexp.MustCompile("[^a-z0-9]+")
	for _, p := range ap {
		if _, ok := allArticles[p.Slug]; !ok {
			allArticles[p.Slug] = p
		}

		if _, ok := tags[p.Tag]; !ok && len(p.Tag) > 0 {
			tags[p.Tag] = Tag{Slug: strings.Trim(re.ReplaceAllString(strings.ToLower(p.Tag), "-"), "-"), Category: p.Category}
		}

		if _, ok := categories[p.Category]; !ok && len(p.Category) > 0 {
			categories[p.Category] = strings.Trim(re.ReplaceAllString(strings.ToLower(p.Category), "-"), "-")
		}
	}

	pg, err := getPages()
	if err != nil {
		log.Fatal(err)
	}

	for _, p := range pg {
		if _, ok := allPages[p.Slug]; !ok {
			allPages[p.Slug] = p
		}
	}
}

func main() {
	loadTemplates()

	allPages = make(map[string]*Page)
	allArticles = make(map[string]*Page)
	authors = make(map[string]*Author)
	tags = make(map[string]Tag)
	categories = make(map[string]string)

	cacheData()

	http.HandleFunc("/content/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, r.URL.Path[1:])
	})

	http.HandleFunc("/article/", article)
	http.HandleFunc("/edit/", edit)
	http.HandleFunc("/save/", save)
	http.HandleFunc("/articles/", articles)
	http.HandleFunc("/del/", delHandler)

	http.HandleFunc("/auteur/", authorHandler)

	http.HandleFunc("/infolettre", func(w http.ResponseWriter, r *http.Request) {
		data := &pageData{}
		data.Title = "Inscription à l'infolettre de Parentzie"

		render(w, "nl.hbs", data)
	})

	http.HandleFunc("/recherche", func(w http.ResponseWriter, r *http.Request) {
		data := &pageData{}
		data.Title = "Résultat de recherche"
		render(w, "search.hbs", data)
	})

	http.HandleFunc("/accueil/", index)
	http.HandleFunc("/", index)

	port := os.Getenv("HTTP_PLATFORM_PORT")
	if len(port) == 0 {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
