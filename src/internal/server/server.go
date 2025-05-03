package server

import (
	"code-server-launcher/internal/config"
	"code-server-launcher/internal/logger"
	"code-server-launcher/internal/service"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
)

type Server struct {
	log            *logger.Logger
	config         *config.Server
	userService    *service.UserService
	proxmoxService *service.ProxmoxService
	githubConfig   *config.Github
	oauth2         *oauth2.Config
}

func NewServer(cfg *config.Config) *Server {
	return &Server{
		log:            logger.NewLogger("Server"),
		config:         cfg.Server,
		userService:    service.NewUserService(cfg),
		proxmoxService: service.NewProxmoxService(cfg.Proxmox),
		githubConfig:   cfg.Github,
		oauth2:         cfg.Github.GetOAuth(),
	}
}

func (s *Server) Start() {
	http.HandleFunc("/", s.handleHome)
	http.HandleFunc("/login", s.handleLogin)
	http.HandleFunc("/callback", s.handleCallback)
	http.HandleFunc("/users", s.handleUsers)

	addr := s.config.Host + ":" + string(s.config.Port)

	s.log.Info("Server started at %s", addr)
	err := http.ListenAndServe(addr, nil)

	if err != nil {
		s.log.Error("HTTP Server Return: %v", err)
	}
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, `<a href="/login">Login with GitHub</a>`)
	log.Println("Home page accessed")
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	log.Println("Login page accessed")

	url := s.oauth2.AuthCodeURL("state-token", oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (s *Server) handleCallback(w http.ResponseWriter, r *http.Request) {
	log.Println("Callback page accessed")
	code := r.URL.Query().Get("code")
	token, err := s.oauth2.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	client := s.oauth2.Client(context.Background(), token)
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

	if !s.authUser(user.Login) {
		log.Printf("Access denied for user: %s", user.Login)
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	http.SetCookie(w, &http.Cookie{Name: "user", Value: user.Login})

	target := fmt.Sprintf("https://%s.vm.learnops.duckdns.org", user.Login)
	http.Redirect(w, r, target, http.StatusSeeOther)
}

func (s *Server) handleUsers(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("user")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user := cookie.Value

	if !s.authUser(user) {
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

func (s *Server) authUser(user string) bool {
	log.Printf("Auth user: %s", user)

	if port, ok := allowedUsers[user]; ok {
		log.Printf("User %s is allowed with port %d", user, port)
		return true
	}

	log.Printf("User %s is not allowed", user)
	return false
}
