package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"encoding/json"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

var (
	oauthConf = &oauth2.Config{
		ClientID:     "Ov23libexupt9yDqNXEP",                     //os.Getenv(),
		ClientSecret: "2a700d4fe53424d22756a2d3d1044166a4dbe85d", //os.Getenv("2a700d4fe53424d22756a2d3d1044166a4dbe85d"),
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
		RedirectURL:  "https://code.learnops.duckdns.org/callback",
	}

	// Lista de usuários permitidos (login do GitHub)
	allowedUsers = map[string]int{}
)

func main() {
	allowedUsers = loadUsers()

	http.HandleFunc("/", handleHome)
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/callback", handleCallback)
	http.HandleFunc("/users", handleUsers)

	fmt.Println("Server started at http://0.0.0.0:8080/")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, `<a href="/login">Login with GitHub</a>`)
	log.Println("Home page accessed")
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	log.Println("Login page accessed")
	url := oauthConf.AuthCodeURL("state-token", oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	log.Println("Callback page accessed")
	code := r.URL.Query().Get("code")
	token, err := oauthConf.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	client := oauthConf.Client(context.Background(), token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var user struct {
		Login string `json:"login"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		http.Error(w, "Failed to parse user info", http.StatusInternalServerError)
		return
	}

	user.Login = strings.ToLower(user.Login)

	log.Printf("User info: %+v\n", user)

	if !authUser(user.Login) {
		log.Printf("Access denied for user: %s", user.Login)
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Sessão fake: salva o usuário em um cookie
	http.SetCookie(w, &http.Cookie{Name: "user", Value: user.Login})

	target := fmt.Sprintf("https://%s.vm.learnops.duckdns.org", user.Login)
	http.Redirect(w, r, target, http.StatusSeeOther)
}

func handleUsers(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("user")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user := cookie.Value

	if !authUser(user) {
		log.Printf("Access denied for user: %s", user)
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	users := make([]string, 0, len(allowedUsers))
	for u := range allowedUsers {
		users = append(users, u)
	}

	json.NewEncoder(w).Encode(users)
}

func authUser(user string) bool {
	log.Printf("Auth user: %s", user)

	if port, ok := allowedUsers[user]; ok {
		log.Printf("User %s is allowed with port %d", user, port)
		return true
	}

	log.Printf("User %s is not allowed", user)
	return false
}

func loadUsers() map[string]int {
	return map[string]int{
		"rafaelfino": 7500,
	}
}

func startDocker(user string, port int) error {
	return nil
}
