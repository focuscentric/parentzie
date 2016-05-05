package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type pageData struct {
	Title       string
	HasImage    bool
	Description string
	Keywords    string
	Thumbnail   string
	URL         string

	Article  *Page
	Author   *Author
	ByAuthor []*Page
	Popular  []*Page
	Articles []*Page
	Featured *Page

	HasPreviousPage bool
	HasNextPage     bool
	PreviousPageURL string
	NextPageURL     string
	PageIndex       int

	SubMenu map[string]Tag

	Categories map[string]string
	Tags       map[string]Tag
	Authors    map[string]*Author

	EstimatedReadTime int
}

func pageHandler(w http.ResponseWriter, r *http.Request) {
	page := getID(r.URL.Path, "/")
	p, ok := allPages[page]
	if len(page) == 0 || !ok {
		http.NotFound(w, r)
		return
	}

	data := &pageData{}
	data.Title = p.Title
	data.Article = p

	render(w, "page.hbs", data)
}

func index(w http.ResponseWriter, r *http.Request) {
	if ok := getID(r.URL.Path, "/"); len(ok) > 0 && strings.Index(ok, "accueil/") == -1 {
		pageHandler(w, r)
		return
	}

	data := &pageData{}

	data.Description = "Vivre la famille naturellement"
	data.HasImage = false
	data.Title = "Parentzie - Vivre la famille naturellement"

	data.PageIndex = 1
	data.HasNextPage = true

	page := getID(r.URL.Path, "/accueil/")
	if len(page) > 0 {
		pageIndex, err := strconv.ParseInt(page, 10, 32)
		if err != nil {
			log.Println(err.Error())
			data.Featured = sortedArticles[0]
			data.Articles = sortedArticles[1:9]
		} else {
			start := pageIndex*10 - 10

			data.Featured = sortedArticles[start]
			data.Articles = sortedArticles[start+1 : start+9]
		}

		data.HasPreviousPage = true
	} else {
		data.Featured = sortedArticles[0]
		data.Articles = sortedArticles[1:9]
	}

	render(w, "index.hbs", data)
}

func articles(w http.ResponseWriter, r *http.Request) {
	data := &pageData{}
	data.SubMenu = make(map[string]Tag)

	id := getID(r.URL.Path, "/articles/")
	if len(id) == 0 {
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
		return
	}

	var matches []*Page
	var catName string
	var tagName string

	for k, v := range tags {
		if v.Slug == id {
			tagName = k
			catName = v.Category
			break
		}
	}

	if len(tagName) > 0 {
		for _, p := range sortedArticles {
			if p.Tag == tagName {
				matches = append(matches, p)
			}
		}
	} else {
		for k, v := range categories {
			if v == id {
				catName = k
				break
			}
		}

		if len(catName) > 0 {
			for _, p := range sortedArticles {
				if p.Category == catName {
					matches = append(matches, p)
				}
			}
		}
	}

	if len(catName) > 0 {
		for k, v := range tags {
			if v.Category == catName {
				data.SubMenu[k] = v
			}
		}
	}

	data.Description = "Vivre la famille naturellement"
	data.HasImage = false
	if len(tagName) > 0 {
		data.Title = tagName + " : articles"
	} else if len(catName) > 0 {
		data.Title = catName + " : articles"
	}

	data.PageIndex = 1
	data.HasNextPage = true

	page := r.URL.Query().Get("p")
	if len(page) > 0 {
		pageIndex, err := strconv.ParseInt(page, 10, 32)
		if err != nil {
			log.Println(err.Error())
			data.Featured = matches[0]
			if len(matches) > 10 {
				data.Articles = matches[1:9]
			} else {
				data.Articles = matches[1:]
			}
		} else {
			start := int(pageIndex*10 - 10)
			end := start + 9

			if end > len(matches)-1 {
				end = len(matches) - 1
			} else {
				data.NextPageURL = fmt.Sprintf("/articles/%s?p=%d", id, pageIndex+1)
			}

			data.Featured = matches[start]
			data.Articles = matches[start+1 : end]
		}

		data.HasPreviousPage = true

		data.PreviousPageURL = fmt.Sprintf("/articles/%s?p=%d", id, pageIndex-1)
	} else if len(matches) > 0 {
		data.Featured = matches[0]
		if len(matches) > 10 {
			data.Articles = matches[1:9]

			data.NextPageURL = fmt.Sprintf("/articles/%s?p=2", id)
		} else {
			data.Articles = matches[1:]
		}
	}

	render(w, "articles.hbs", data)
}

func article(w http.ResponseWriter, r *http.Request) {
	slug := getID(r.URL.Path, "/article/")
	p, ok := allArticles[slug]
	if !ok {
		http.NotFound(w, r)
		return
	}

	data := &pageData{}
	data.Title = p.Title
	data.Description = p.Title
	data.HasImage = true
	data.Thumbnail = p.Thumbnail
	data.URL = r.URL.String()

	data.Article = p
	if a, ok := authors[p.AuthorName]; ok {
		data.Author = a
	} else {
		data.Author = &Author{AuthorName: "Non trouvé"}
	}

	var sameAuthor []*Page
	var sameCat []*Page
	for _, page := range sortedArticles {
		if page.Slug != slug && page.AuthorName == p.AuthorName && len(sameAuthor) < 10 {
			sameAuthor = append(sameAuthor, page)
		}

		if page.Slug != slug && page.Category == p.Category && len(sameCat) < 10 {
			sameCat = append(sameCat, page)
		}

		if len(sameCat) >= 10 && len(sameAuthor) >= 10 {
			break
		}
	}

	data.ByAuthor = sameAuthor
	data.Popular = sameCat

	// calculating read time
	wc := len(strings.Split(stripHTML(data.Article.Body), " "))
	fmt.Printf("wc: %d\n", wc)
	data.EstimatedReadTime = 1
	if wc > 200 {
		data.EstimatedReadTime = int(math.Ceil(float64(wc) / 200.0))
	}

	render(w, "article.hbs", data)
}

func edit(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key != os.Getenv("FOCUSKEY") {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}

	slug := getID(r.URL.Path, "/edit/")
	p, ok := allArticles[slug]
	if !ok {
		p, ok = allPages[slug]
		if !ok {
			p = &Page{}
		}
	}

	data := &pageData{}
	data.Title = p.Title
	data.Description = p.Title
	data.HasImage = true
	data.Thumbnail = p.Thumbnail
	data.URL = r.URL.String()

	data.Article = p
	if a, ok := authors[p.AuthorName]; ok {
		data.Author = a
	} else {
		data.Author = &Author{AuthorName: "Non trouvé"}
	}

	data.Categories = categories
	data.Tags = tags
	data.Authors = authors

	render(w, "edit.hbs", data)
}

func save(w http.ResponseWriter, r *http.Request) {
	var result = new(struct {
		State bool   `json:"state"`
		Slug  string `json:"slug"`
	})

	var data Page
	err := parseBody(r.Body, &data)
	if err != nil {
		log.Println(err.Error())
		respond(w, r, http.StatusBadRequest, false)
		return
	}

	log.Printf("Tag is: %s\n", data.Tag)

	found := false
	for _, p := range allArticles {
		if p.ID == data.ID {
			err = savePage(data)
			found = true
			break
		}
	}

	if !found {
		err = savePage(data)
	}

	if err != nil {
		log.Println(err.Error())
		result.State = false
	} else {
		result.State = true
		result.Slug = data.Slug
	}

	respond(w, r, http.StatusOK, result)
}

func authorHandler(w http.ResponseWriter, r *http.Request) {
	data := &pageData{}
	slug := getID(r.URL.Path, "/auteur/")
	for _, v := range authors {
		if v.Slug == slug {
			data.Author = v
			break
		}
	}

	if data.Author == nil {
		http.NotFound(w, r)
		return
	}

	data.Title = data.Author.AuthorName + " auteur chez Parentzie"

	for _, p := range sortedArticles {
		if p.AuthorName == data.Author.AuthorName {
			data.Articles = append(data.Articles, p)
		}
	}

	render(w, "author.hbs", data)
}

func delHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	id := getID(r.URL.Path, "/del/")
	pageID, err := strconv.ParseInt(id, 10, 64)
	if err != nil || key != "wtfhc" || len(id) == 0 {
		if err != nil {
			log.Println(err.Error())
		} else {
			log.Printf("invalid delete id: %s", id)
		}
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	err = removePage(int(pageID))
	if err != nil {
		log.Println(err.Error())
	} else {

		found := false
		for k, v := range allPages {
			if v.ID == int(pageID) {
				found = true
				delete(allPages, k)
				break
			}
		}

		if !found {
			for k, v := range allArticles {
				if v.ID == int(pageID) {
					delete(allArticles, k)
					break
				}
			}
		}

		cacheData()
	}

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func getID(url string, controller string) string {
	if len(url) < len(controller) || strings.ToUpper(url) == strings.ToUpper(controller) {
		return ""
	}

	return url[len(controller):]
}
