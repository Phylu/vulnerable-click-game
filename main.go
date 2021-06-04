package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"

	scs "github.com/alexedwards/scs/v2"
	_ "github.com/mattn/go-sqlite3"
)

var sessionManager *scs.SessionManager

var DB *sql.DB

func InitDB() {
	database, err := sql.Open("sqlite3", "./sqlite.db")
	if err != nil {
		log.Fatal(err)
	}
	DB = database
}

func main() {
	// Initialize Database
	InitDB()
	defer DB.Close()

	// Initialize a new session manager and configure the session lifetime.
	sessionManager = scs.New()
	sessionManager.Lifetime = 24 * time.Hour

	mux := http.NewServeMux()

	// Register Static Files
	fileServer := http.FileServer((http.Dir("./static")))
	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))

	// Register Dynamic Files
	mux.HandleFunc("/login", login)
	mux.HandleFunc("/logout", logout)
	mux.HandleFunc("/register", staticPage)
	mux.HandleFunc("/game", loggedInStaticPage)
	mux.HandleFunc("/highscore", loggedInStaticPage)
	mux.HandleFunc("/", index)

	// Register API
	mux.HandleFunc("/api/score", submitPoints)
	mux.HandleFunc("/api/highscore", getHighscore)

	port := "8080"
	if _, ok := os.LookupEnv("PORT"); ok {
		port = os.Getenv("PORT")
	}

	// Start Webserver
	log.Println("Starting Webserver on port ", port)
	err := http.ListenAndServe(":"+port, sessionManager.LoadAndSave(mux))
	if err != nil {
		log.Fatal(err)
	}

}

func index(w http.ResponseWriter, r *http.Request) {
	userID := sessionManager.GetInt(r.Context(), "userID")
	log.Println("Checking login for userID: ", userID)

	if userID == 0 {
		http.Redirect(w, r, "/login", 302)
	} else {
		executeTemplate("/index", w)
	}

}

func login(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()

	if r.Method == "POST" {
		userID := checkLogin(r.Form)
		if userID > 0 {
			log.Println("Login detected. Setting userID to: ", userID)
			sessionManager.Put(r.Context(), "userID", userID)
			http.Redirect(w, r, "/", 302)
		}
	}

	executeTemplate(r.URL.Path, w)
}

func logout(w http.ResponseWriter, r *http.Request) {
	log.Println("Removing Session ID from Session")
	sessionManager.Remove(r.Context(), "userID")

	http.Redirect(w, r, "/login", 302)
}

func loggedInStaticPage(w http.ResponseWriter, r *http.Request) {
	loggedIn(w, r)
	staticPage(w, r)
}

func staticPage(w http.ResponseWriter, r *http.Request) {
	executeTemplate(r.URL.Path, w)
}

func executeTemplate(templateFile string, w http.ResponseWriter) {
	templatePath := fmt.Sprintf("templates%s.html", templateFile)
	t, err := template.ParseFiles(templatePath)
	if err != nil {
		log.Println("Template Parsing Error: ", err)
	}
	err = t.Execute(w, nil)
	if err != nil {
		log.Println("Template Execution Error: ", err)
	}
}

func checkLogin(form map[string][]string) int {
	// Prepare Query
	// This query is intentionally built like this to allow an sql injection.
	// Instead you should run db().QueryRow("SELECT id FROM users WHERE email = ? and password = ?", form["email"][0], form["password"][0]").Scan(&id)
	query := fmt.Sprintf("SELECT id FROM users WHERE email = '%s' and password = '%s'", form["email"][0], form["password"][0])
	log.Println(query)

	// Get User from Database
	var id int
	err := DB.QueryRow(query).Scan(&id)
	if err != nil {
		log.Println("Problem retrieving user from database: ", err)
	}

	return id
}

func userID(w http.ResponseWriter, r *http.Request) int {
	return sessionManager.GetInt(r.Context(), "userID")
}

func loggedIn(w http.ResponseWriter, r *http.Request) int {
	userID := userID(w, r)
	// If the userID == 0, the user is not loggid in
	if userID == 0 {
		http.Redirect(w, r, "/login", 302)
	}

	return userID
}

type Score struct {
	UserID int    `json:"userid"`
	Name   string `json:"name"`
	Points int    `json:"points"`
}

func submitPoints(w http.ResponseWriter, r *http.Request) {
	var score Score

	userID := userID(w, r)
	if userID == 0 {
		w.WriteHeader(http.StatusForbidden)
	}

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		log.Println("Problem receiving Points: ", err)
	}
	if err := r.Body.Close(); err != nil {
		log.Println("Problem receiving Points: ", err)
	}

	if err := json.Unmarshal(body, &score); err != nil {
		log.Println(err)
	}

	stmt, err := DB.Prepare("INSERT INTO highscore(name, userid, points) values(?, ?, ?)")
	if err != nil {
		log.Println(err)
	}

	log.Println(score.Name, userID, score.Points)
	_, err = stmt.Exec(score.Name, userID, score.Points)
	if err != nil {
		log.Println(err)
	}

}

func getHighscore(w http.ResponseWriter, r *http.Request) {
	/*userID := userID(w, r)
	if userID == 0 {
		w.WriteHeader(http.StatusForbidden)
	}*/

	var scores []Score

	result, err := DB.Query("SELECT name, userid, points FROM highscore")
	if err != nil {
		log.Println(err)
	}

	defer result.Close()
	for result.Next() {
		var score Score
		err = result.Scan(&score.Name, &score.UserID, &score.Points)
		if err != nil {
			log.Println(err)
		}
		scores = append(scores, score)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(scores)
}
