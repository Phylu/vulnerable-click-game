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
	"strconv"
	"text/template"
	"time"

	scs "github.com/alexedwards/scs/v2"
	_ "github.com/mattn/go-sqlite3"
	"github.com/phylu/vulnerable-click-game/seeder"
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
	seeder.Seed()
	InitDB()
	defer DB.Close()

	// Initialize a new session manager and configure the session lifetime.
	sessionManager = scs.New()
	sessionManager.Lifetime = 24 * time.Hour

	mux := http.NewServeMux()

	// Register Static Files
	fileServer := http.FileServer((http.Dir(".")))
	mux.Handle("/", fileServer)
	// !!!But not in the root directory of the repository!!!
	// fileServer := http.FileServer((http.Dir("./static")))
	// mux.Handle("/static/", http.StripPrefix("/static/", fileServer))

	// Register Dynamic Files
	mux.HandleFunc("/login", login)
	mux.HandleFunc("/logout", logout)
	mux.HandleFunc("/register", staticPage)
	mux.HandleFunc("/dashboard", dashboard)
	mux.HandleFunc("/imprint", staticPage)
	mux.HandleFunc("/game", loggedInStaticPage)
	mux.HandleFunc("/highscore", loggedInStaticPage)
	mux.HandleFunc("/profile", loggedInStaticPage)
	mux.HandleFunc("/setup", setup)

	// Register API
	mux.HandleFunc("/api/score", submitPoints)
	mux.HandleFunc("/api/highscore", getHighscore)
	mux.HandleFunc("/api/userscores", getUserScores)

	port := "8080"
	if _, ok := os.LookupEnv("PORT"); ok {
		port = os.Getenv("PORT")
	}

	// Start Webserver
	log.Println("Starting Webserver on port", port)
	// We need to make sure that we add security headers to all requests using e.g. a middleware
	// w.Header().Add("X-Frame-Options", "DENY")
	err := http.ListenAndServe(":"+port, sessionManager.LoadAndSave(mux))
	if err != nil {
		log.Fatal(err)
	}

}

func dashboard(w http.ResponseWriter, r *http.Request) {
	userID := sessionManager.GetInt(r.Context(), "userID")

	if userID == 0 {
		log.Println("No user... :(")
		http.Redirect(w, r, "/", http.StatusFound)
	} else {
		email := getUserEmail(userID)
		var d TemplateData
		d.Email = email
		executeTemplate("/dashboard", w, d)
	}

}

func login(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()

	if r.Method == "POST" {
		userID := checkLogin(r.Form)
		if userID > 0 {
			log.Println("Login detected. Setting userID to: ", userID)
			sessionManager.Put(r.Context(), "userID", userID)
			http.Redirect(w, r, "/dashboard", http.StatusFound)
			return
		}
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func logout(w http.ResponseWriter, r *http.Request) {
	log.Println("Removing Session ID from Session")
	sessionManager.Remove(r.Context(), "userID")

	http.Redirect(w, r, "/login", http.StatusFound)
}

func loggedInStaticPage(w http.ResponseWriter, r *http.Request) {
	userID := loggedIn(w, r)
	email := getUserEmail(userID)
	var d TemplateData
	d.Email = email

	executeTemplate(r.URL.Path, w, d)
}

func staticPage(w http.ResponseWriter, r *http.Request) {
	var d TemplateData
	executeTemplate(r.URL.Path, w, d)
}

type TemplateData struct {
	Email string
}

func executeTemplate(templateFile string, w http.ResponseWriter, d TemplateData) {
	templatePath := fmt.Sprintf("templates%s.html", templateFile)
	t, err := template.ParseFiles(templatePath)
	if err != nil {
		log.Println("Template Parsing Error: ", err)
	}
	err = t.Execute(w, d)
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
	// If the userID == 0, the user is not logged in
	if userID == 0 {
		http.Redirect(w, r, "/login", http.StatusFound)
	}

	return userID
}

func getUserEmail(userID int) string {
	var email string
	err := DB.QueryRow("SELECT email FROM users WHERE id = ?", userID).Scan(&email)
	if err != nil {
		log.Println("Problem retrieving user from database: ", err)
	}
	return email
}

type Score struct {
	UserID       int    `json:"userid"`
	Email        string `json:"email"`
	Points       int    `json:"points"`
	VictoryShout string `json:"victoryshout"`
}

func submitPoints(w http.ResponseWriter, r *http.Request) {
	userID := userID(w, r)
	if userID == 0 {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var score Score

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

	stmt, err := DB.Prepare("INSERT INTO highscore(userid, points, victoryshout) values(?, ?, ?)")
	if err != nil {
		log.Println(err)
	}

	log.Println(userID, score.Points, score.VictoryShout)
	// There is no SQL-I possible here, but an XSS. Why don't we use html.EscapeString() on the values?
	_, err = stmt.Exec(userID, score.Points, score.VictoryShout)
	if err != nil {
		log.Println(err)
	}

}

type Highscore struct {
	Data []Score `json:"data"`
}

func getHighscore(w http.ResponseWriter, r *http.Request) {
	userID := userID(w, r)
	if userID == 0 {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	scores := []Score{}

	result, err := DB.Query("SELECT userid, points, victoryshout FROM highscore")
	if err != nil {
		log.Println(err)
	}

	defer result.Close()
	for result.Next() {
		var score Score
		err = result.Scan(&score.UserID, &score.Points, &score.VictoryShout)
		if err != nil {
			log.Println(err)
		}
		score.Email = getUserEmail(score.UserID)
		scores = append(scores, score)
	}

	var highscore Highscore
	highscore.Data = scores

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(highscore)
}

func getUserScores(w http.ResponseWriter, r *http.Request) {
	userID := userID(w, r)
	if userID == 0 {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Why do we even allow the user to provide a userid in the query instead of using the userID from the session directly?
	profileID := r.URL.Query().Get("userid")
	if profileID == "me" {
		profileID = strconv.Itoa(userID)
	}

	scores := []Score{}

	result, err := DB.Query("SELECT userid, points, victoryshout FROM highscore WHERE userid = ?", profileID)
	if err != nil {
		log.Println(err)
	}

	defer result.Close()
	for result.Next() {
		var score Score
		err = result.Scan(&score.UserID, &score.Points, &score.VictoryShout)
		if err != nil {
			log.Println(err)
		}
		score.Email = getUserEmail(score.UserID)
		scores = append(scores, score)
	}

	var userHighscore Highscore
	userHighscore.Data = scores

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userHighscore)
}

func setup(w http.ResponseWriter, r *http.Request) {
	DB.Close()
	seeder.Seed()
	InitDB()
	io.WriteString(w, "Success.\n")
}
