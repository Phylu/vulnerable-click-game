package main

import (
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"

	scs "github.com/alexedwards/scs/v2"
)

var sessionManager *scs.SessionManager

func main() {

	// Initialize a new session manager and configure the session lifetime.
	sessionManager = scs.New()
	sessionManager.Lifetime = 24 * time.Hour

	mux := http.NewServeMux()

	// Register Static Files
	fileServer := http.FileServer((http.Dir("./static")))
	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))

	// Register Dynamic Files
	mux.HandleFunc("/login", login)
	mux.HandleFunc("/register", staticPage)
	mux.HandleFunc("/", index)

	// Start Webserver
	log.Println("Starting Webserver on http://localhost:8080")
	err := http.ListenAndServe(":8080", sessionManager.LoadAndSave(mux))
	if err != nil {
		log.Fatal(err)
	}

}

func index(w http.ResponseWriter, r *http.Request) {
	userID := sessionManager.GetString(r.Context(), "userID")
	log.Println("Checking login for userID: ", userID)

	if userID == "" {
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
	return 1
}
