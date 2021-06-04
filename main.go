package main

import (
	"database/sql"
	"fmt"
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
	mux.HandleFunc("/", index)

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
