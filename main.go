package main

import (
	"crypto/rand"
	"encoding/base64"
	"html/template"
	"log"
	"net/http"
	"sync"
)

type User struct {
	Password string
	Role     string
}

var users = map[string]User{
	"student1": {Password: "password", Role: "student"},
	"admin1":   {Password: "password", Role: "admin"},
}

var (
	sessions   = make(map[string]string)
	sessionMux sync.RWMutex
)

func generateSessionID() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		sessionMux.RLock()
		_, exists := sessions[cookie.Value]
		sessionMux.RUnlock()

		if !exists {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		next(w, r)
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	//User role validation
	var role string
	cookie, err := r.Cookie("session_id")
	if err == nil {
		sessionMux.RLock()
		if username, exists := sessions[cookie.Value]; exists {
			role = users[username].Role
		}
		sessionMux.RUnlock()
	}

	var files []string
	var scheduleData []DayRow

	if role == "admin" {
		files = []string{
			"./templates/base.tmpl",
			"./templates/admin_dashboard.tmpl",
		}
	} else {
		files = []string{
			"./templates/base.tmpl",
			"./templates/schedule.tmpl",
		}

	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	scheduleData = NewScheduleHandler(w, r)

	data := struct {
		Schedule []DayRow
		Role     string
	}{
		Schedule: scheduleData,
		Role:     role,
	}

	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// GET /login - Serve login form
func loginPage(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"./templates/base.tmpl",
		"./templates/login.tmpl",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	err = ts.ExecuteTemplate(w, "base", nil)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func loginPost(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	user, exists := users[username]
	if !exists || user.Password != password {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("<p style='color:red;'>Invalid username or password</p>"))
		return
	}

	sessionID := generateSessionID()
	sessionMux.Lock()
	sessions[sessionID] = username
	sessionMux.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
	})

	w.Header().Set("HX-Redirect", "/")
}

func logoutPost(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		sessionMux.Lock()
		delete(sessions, cookie.Value)
		sessionMux.Unlock()
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	w.Header().Set("HX-Redirect", "/login")
}

func main() {
	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("./static/"))

	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	mux.HandleFunc("GET /login", loginPage)
	mux.HandleFunc("POST /login", loginPost)
	mux.HandleFunc("POST /logout", logoutPost)

	mux.HandleFunc("GET /", RequireAuth(home))
	mux.HandleFunc("POST /schedule/select", RequireAuth(UpdateScheduleHandler))

	log.Print("starting server on :4000")
	err := http.ListenAndServe(":4000", mux)
	log.Fatal(err)
}
