package main

import (
	"database/sql"
)

// Article - this structure is used to represent an article for most of the functions
type Article struct {
	ID    int      `json:"id"`
	Title string   `json:"title"`
	Date  string   `json:"date"`
	Body  string   `json:"body"`
	Tags  []string `json:"tags"`
}

func (a *Article) getArticleById(db *sql.DB) error {
	article := Article{}
	rows, err := db.Query("SELECT id, title, date, body FROM articles WHERE id = $1", a.ID)

	if err != nil {
		return err
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)

	for rows.Next() {
		if err := rows.Scan(&article.ID, &article.Title, &article.Date, &article.Body); err != nil {
			return err
		}
	}

	rows, err = db.Query("SELECT tag FROM tags WHERE article_id = $1", a.ID)

	if err != nil {
		return err
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)

	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return err
		}
		article.Tags = append(article.Tags, tag)
	}

	*a = article

	return nil
}

func (a *Article) createArticle(db *sql.DB) error {
	err := db.QueryRow(
		"INSERT INTO public.articles(title, date, body) VALUES($1, $2, $3) RETURNING id",
		a.Title, a.Date, a.Body).Scan(&a.ID)

	if err != nil {
		return err
	}

	for _, tag := range a.Tags {
		_, err := db.Exec("INSERT INTO public.tags(tag, article_id) VALUES($1, $2)", tag, a.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func getArticles(db *sql.DB) ([]Article, error) {
	rows, err := db.Query("SELECT id, title, date, body FROM articles")

	if err != nil {
		return nil, err
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)

	var articles []Article

	for rows.Next() {
		article := Article{}
		if err := rows.Scan(&article.ID, &article.Title, &article.Date, &article.Body); err != nil {
			return nil, err
		}

		rowsTags, err := db.Query("SELECT tag FROM tags WHERE article_id = $1", article.ID)
		if err != nil {
			return nil, err
		}

		for rowsTags.Next() {
			var tag string
			if err := rowsTags.Scan(&tag); err != nil {
				return nil, err
			}
			article.Tags = append(article.Tags, tag)
		}

		err = rowsTags.Close()
		if err != nil {
			return nil, err
		}

		articles = append(articles, article)
	}
	return articles, nil
}
