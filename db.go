package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"os"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
)

var db *sql.DB

// Page contains all fields for a page data
type Page struct {
	ID            int    `json:"id"`
	AuthorName    string `json:"author"`
	Category      string `json:"category"`
	CategorySlug  string
	Slug          string `json:"slug"`
	Thumbnail     string `json:"thumbnail"`
	Title         string `json:"title"`
	Body          string `json:"content"`
	BodyHTML      template.HTML
	Excerpt       string
	IsFeatured    bool `json:"isFeatured"`
	IsFree        bool `json:"isFree"`
	HasAudio      bool `json:"hasAudio"`
	HasVideo      bool `json:"hasVideo"`
	Rating        float32
	Viewed        int
	Commented     int
	Created       time.Time
	Published     time.Time `json:"pubDate"`
	PublishedText string
	Tag           string `json:"tag"`
	TagSlug       string
	Author        *Author
}

// Author contains all fields for an author
type Author struct {
	AuthorName string
	Slug       string
	Bio        string
	BioHTML    template.HTML
	Thumbnail  string
	Image      string
}

func openConnection() error {
	d, err := sql.Open("mssql", os.Getenv("FOCUSDB"))
	if err != nil {
		return err
	}

	err = d.Ping()
	if err != nil {
		return err
	}

	db = d

	return nil
}

func closeConnection() {
	if db != nil {
		db.Close()
	}
}

func readPage(rows *sql.Rows, p *Page) error {
	err := rows.Scan(
		&p.ID,
		&p.AuthorName,
		&p.Category,
		&p.Slug,
		&p.Thumbnail,
		&p.Title,
		&p.Body,
		&p.IsFeatured,
		&p.IsFree,
		&p.HasAudio,
		&p.HasVideo,
		&p.Rating,
		&p.Viewed,
		&p.Commented,
		&p.Created,
		&p.Published,
		&p.Tag,
	)

	p.BodyHTML = template.HTML(p.Body)

	p.Excerpt = stripHTML(p.Body)
	if len(p.Excerpt) > 250 {
		p.Excerpt = p.Excerpt[0:250] + "..."
	}

	if s, ok := tags[p.Tag]; ok {
		p.TagSlug = s.Slug
	}

	if s, ok := categories[p.Category]; ok {
		p.CategorySlug = s
	}

	if a, ok := authors[p.AuthorName]; ok {
		p.Author = a
	}

	months := []string{"décembre", "janvier", "février", "mars", "avril", "mai", "juin", "juillet", "août", "septembre", "octobre", "novembre", "décembre", "janvier"}

	p.PublishedText = fmt.Sprintf("%d %s %d", p.Published.Day(), months[p.Published.Month()], p.Published.Year())

	return err
}

func readAuthor(rows *sql.Rows, a *Author) error {
	err := rows.Scan(
		&a.AuthorName,
		&a.Slug,
		&a.Bio,
		&a.Thumbnail,
		&a.Image,
	)

	a.BioHTML = template.HTML(a.Bio)
	return err
}

func getPages() ([]*Page, error) {
	sql, err := db.Prepare("SELECT p.* FROM _Pages p INNER JOIN _Authors a ON p.AuthorName = a.AuthorName WHERE p.Category = '' ORDER BY Published DESC")
	if err != nil {
		return nil, err
	}
	defer sql.Close()

	rows, err := sql.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pages []*Page

	for rows.Next() {
		p := &Page{}
		err := readPage(rows, p)
		if err != nil {
			return nil, err
		}
		pages = append(pages, p)
	}

	return pages, nil
}

func getArticles() ([]*Page, error) {
	sql, err := db.Prepare("SELECT p.* FROM _Pages p INNER JOIN _Authors a ON p.AuthorName = a.AuthorName WHERE p.Category <> '' ORDER BY Published DESC")
	if err != nil {
		return nil, err
	}
	defer sql.Close()

	rows, err := sql.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pages []*Page

	for rows.Next() {
		p := &Page{}
		err := readPage(rows, p)
		if err != nil {
			return nil, err
		}
		pages = append(pages, p)
	}

	return pages, nil
}

func getAuthors() ([]*Author, error) {
	sql, err := db.Prepare("SELECT * FROM _Authors")
	if err != nil {
		return nil, err
	}
	defer sql.Close()

	rows, err := sql.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var authors []*Author

	for rows.Next() {
		a := &Author{}
		err := readAuthor(rows, a)
		if err != nil {
			return nil, err
		}
		authors = append(authors, a)
	}

	return authors, nil
}

func savePage(p Page) (err error) {
	openConnection()
	defer closeConnection()

	if p.ID > 0 {
		log.Println("article exists")
		stmt, err := db.Prepare(`UPDATE _Pages SET
		AuthorName = ?,
    Category = ?,
    Thumbnail = ?,
    Title = ?,
    Body = ?,
    IsFeatured = ?,
    IsFree = ?,
    HasAudio = ?,
    HasVideo = ?,
    Published = ?,
    Tag = ?
    WHERE ID = ?`)
		if err != nil {
			log.Println("update error: " + err.Error())
			return err
		}
		defer stmt.Close()

		_, err = stmt.Exec(p.AuthorName,
			p.Category,
			p.Thumbnail,
			p.Title,
			p.Body,
			p.IsFeatured,
			p.IsFree,
			p.HasAudio,
			p.HasVideo,
			p.Published,
			p.Tag,
			p.ID,
		)

		log.Println("executed update")

		if _, ok := allArticles[p.Slug]; ok {
			delete(allArticles, p.Slug)
		}

	} else {
		log.Println("adding article")
		stmt, err := db.Prepare("INSERT INTO _Pages VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
		if err != nil {
			log.Println("error preparing insert: " + err.Error())
			return err
		}
		defer stmt.Close()

		fmt.Printf("%v", p)

		_, err = stmt.Exec(p.AuthorName,
			p.Category,
			p.Slug,
			p.Thumbnail,
			p.Title,
			p.Body,
			p.IsFeatured,
			p.IsFree,
			p.HasAudio,
			p.HasVideo,
			p.Rating,
			p.Viewed,
			p.Commented,
			time.Now(),
			p.Published,
			p.Tag,
		)

		if err != nil {
			log.Println(err.Error())
		}

		log.Println("executed insert")
	}

	log.Printf("removing %s", p.Slug)
	if _, ok := allArticles[p.Slug]; ok {
		delete(allArticles, p.Slug)
	}

	if _, ok := allPages[p.Slug]; ok {
		delete(allPages, p.Slug)
	}

	if err == nil {
		cacheData()
	}

	return err
}

func removePage(id int) error {
	openConnection()
	defer closeConnection()

	log.Printf("Deleting: %d\n", id)
	stmt, err := db.Prepare("DELETE FROM _Pages WHERE ID = ?")
	if err != nil {
		log.Println("error preparing insert: " + err.Error())
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	return err
}
