package server

import (
	"code-server-launcher/internal/config"
	"code-server-launcher/internal/domain"
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
	config         *config.ServerConfig
	userService    *service.UserService
	proxmoxService *service.ProxmoxService
	githubConfig   *config.GithubConfig
	oauth2         *oauth2.Config
	allowedUsers   map[string]*domain.User
}

func NewServer(cfg *config.AppConfig) *Server {
	return &Server{
		log:            logger.NewLogger("Http-Server"),
		config:         cfg.Server,
		userService:    service.NewUserService(cfg),
		proxmoxService: service.NewProxmoxService(cfg.Proxmox),
		githubConfig:   cfg.Github,
		oauth2:         cfg.Github.GetOAuth(),
		allowedUsers:   map[string]*domain.User{},
	}

}

func (s *Server) Start() error {
	http.HandleFunc("/", s.handleHome)
	http.HandleFunc("/login", s.handleLogin)
	http.HandleFunc("/callback", s.handleCallback)

	addr := s.config.Host + ":" + string(s.config.Port)

	s.log.Info("Server started at %s", addr)
	err := http.ListenAndServe(addr, nil)

	if err != nil {
		s.log.Error("HTTP Server Return: %v", err)
	}

	return nil
}

func (s *Server) refreshUsers() error {
	users, err := s.userService.LoadUsers()

	if err != nil {
		s.log.Error("Failed to load users: %v", err)
		return err
	}

	for _, user := range users.Users {
		s.allowedUsers[user.Login] = user
	}

	return nil
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, `<a href="/login">Login with GitHub</a>`)
	s.log.Debug("Home page accessed")
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	s.log.Debug("Login page accessed")

	url := s.oauth2.AuthCodeURL("state-token", oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (s *Server) handleCallback(w http.ResponseWriter, r *http.Request) {
	s.log.Debug("Callback page accessed")
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

	s.log.Debug("User info: %+v\n", user)

	if !s.authUser(user.Login) {
		log.Printf("Access denied for user: %s", user.Login)
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	http.SetCookie(w, &http.Cookie{Name: "user", Value: user.Login})

	target := fmt.Sprintf("https://%s.learnops.duckdns.org", user.Login)
	http.Redirect(w, r, target, http.StatusSeeOther)
}

func (s *Server) authUser(user string) bool {
	s.log.Debug("Auth user: %s", user)

	if _, ok := s.allowedUsers[user]; ok {
		s.log.Debug("User %s found in allowed users", user)
		return true
	} else {
		s.log.Debug("User %s not found! Trying to refresh user list", user)
		err := s.refreshUsers()
		if err != nil {
			s.log.Error("Failed to refresh user list: %v", err)
		}

		if _, ok := s.allowedUsers[user]; ok {
			s.log.Debug("User %s found in allowed users", user)
			return true
		} else {
			s.log.Debug("User %s not found in allowed users", user)
		}
	}

	return false
}
