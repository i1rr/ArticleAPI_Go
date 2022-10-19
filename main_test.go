package main_test

import (
	"bytes"
	"encoding/json"
	"github.com/i1rr/ArticleAPI_Go"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
)

var a main.App

func TestMain(m *testing.M) {
	err := godotenv.Load("test_properties.env")
	if err != nil {
		log.Fatal("Error loading properties.env")
	}

	a.Initialize(
		os.Getenv("APP_DB_USERNAME"),
		os.Getenv("APP_DB_PASSWORD"),
		os.Getenv("APP_DB_NAME"))

	ensureTableExists()
	code := m.Run()
	clearTable()
	os.Exit(code)
}

func ensureTableExists() {
	if _, err := a.DB.Exec(tableCreationQuery); err != nil {
		log.Fatal(err)
	}
}

const tableCreationQuery = `CREATE TABLE IF NOT EXISTS articles
(
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    date DATE NOT NULL,
    body TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS tags
(
    tag TEXT,
    article_id INT,
    CONSTRAINT fk_articles FOREIGN KEY(article_id) REFERENCES articles(id) ON DELETE CASCADE
)`

func clearTable() {
	_, err := a.DB.Exec("DELETE FROM articles")
	if err != nil {
		return
	}
	_, err = a.DB.Exec("ALTER SEQUENCE articles_id_seq RESTART WITH 1")
	if err != nil {
		return
	}
	_, err = a.DB.Exec("DELETE FROM tags")
	if err != nil {
		return
	}
	_, err = a.DB.Exec("ALTER SEQUENCE tags_id_seq RESTART WITH 1")
	if err != nil {
		return
	}
}

func TestEmptyTable(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/articles", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func TestGetNonExistentArticle(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/article/-1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

	var m map[string]string
	err := json.Unmarshal(response.Body.Bytes(), &m)
	if err != nil {
		return
	}
	if m["error"] != "Article not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Article not found'. Got '%s'", m["error"])
	}
}

func TestCreateArticle(t *testing.T) {
	clearTable()

	var jsonStr = []byte(`{"title": "test article", "date": "2018-01-01", "body": "test body", "tags": ["test", "test2"]}`)
	req, _ := http.NewRequest("POST", "/article", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	err := json.Unmarshal(response.Body.Bytes(), &m)
	if err != nil {
		return
	}

	if m["title"] != "test article" {
		t.Errorf("Expected article name to be 'test article'. Got '%v'", m["title"])
	}

	if m["date"] != "2018-01-01" {
		t.Errorf("Expected article date to be '2018-01-01'. Got '%v'", m["date"])
	}

	if m["body"] != "test body" {
		t.Errorf("Expected article body to be 'test body'. Got '%v'", m["body"])
	}

	// the id is compared to 1.0 because JSON unmarshalling converts numbers to
	// float, when the target is a map[string]interface{}
	if m["id"] != 1.0 {
		t.Errorf("Expected article ID to be '1'. Got '%v'", m["id"])
	}
}

func TestGetArticle(t *testing.T) {
	clearTable()
	addArticles(1)

	req, _ := http.NewRequest("GET", "/article/1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func addArticles(count int) {
	if count < 1 {
		count = 1
	}

	for i := 0; i < count; i++ {
		_, err := a.DB.Exec("INSERT INTO articles(title, date, body) VALUES($1, $2, $3) RETURNING id",
			"Article "+strconv.Itoa(i), "2018-01-01", "This is a test article")
		if err != nil {
			return
		}
	}
}
