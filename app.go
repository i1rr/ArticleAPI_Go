package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"io"
	"log"
	"net/http"
	"strconv"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
}

func (a *App) Initialize(user, password, dbname string) {
	connectionString := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", user, password, dbname)

	var err error
	a.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}
	a.Router = mux.NewRouter()

	a.initializeRoutes()
}

func (a *App) Run(string) {
	log.Fatal(http.ListenAndServe(":8010", a.Router))
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err := w.Write(response)
	if err != nil {
		return
	}
}

// This function is used to get a single article by id
func (a *App) getArticleById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid article ID")
		return
	}

	ar := Article{ID: id}
	if err := ar.getArticleById(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Article not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	respondWithJSON(w, http.StatusOK, ar)
}

// This function is used to get all articles(no limit)
func (a *App) getAllArticles(w http.ResponseWriter, _ *http.Request) {
	articles, err := getArticles(a.DB)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, articles)
}

// This function is used to create a new article
func (a *App) createArticle(w http.ResponseWriter, r *http.Request) {
	var ar Article
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&ar); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(r.Body)

	if err := ar.createArticle(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, ar)
}

// This function is used to get a summary of articles by tag and date
func (a *App) getArticleSummaryByTagAndDate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tag := vars["tag"]
	date := vars["date"]

	td := TagDate{Tag: tag}

	tagDate, _ := td.getArticleSummaryByTagAndDate(a.DB, date)
	respondWithJSON(w, http.StatusOK, tagDate)
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/articles", a.getAllArticles).Methods("GET")
	a.Router.HandleFunc("/articles", a.createArticle).Methods("POST")
	a.Router.HandleFunc("/articles/{id:[0-9]+}", a.getArticleById).Methods("GET")
	a.Router.HandleFunc("/articles/{tag}/{date}", a.getArticleSummaryByTagAndDate).Methods("GET")
}
