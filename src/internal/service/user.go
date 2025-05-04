package service

import (
	"code-server-launcher/internal/config"
	"code-server-launcher/internal/domain"
	"code-server-launcher/internal/logger"
	"encoding/json"
	"io"
	"net/http"
)

type UserService struct {
	githubUrl   string
	userListUrl string
	log         *logger.Logger
}

func NewUserService(config *config.AppConfig) *UserService {
	return &UserService{
		githubUrl:   config.Github.GithubUrl,
		userListUrl: config.UserListUrl,
		log:         logger.NewLogger("UserService"),
	}
}

func (s *UserService) LoadUsers() (*domain.UserList, error) {
	resp, err := http.Get(s.userListUrl)
	if err != nil {
		s.log.Error("Failed to get JSON file: %v from %s", err, s.userListUrl)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		s.log.Error("Fail to get JSON file: %s, status code: %d", s.userListUrl, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.log.Error("Failed to read response body: %v", err)
	}

	users := domain.NewUserList()

	err = json.Unmarshal(body, &users)

	if err != nil {
		s.log.Error("Failed to unmarshal JSON: %v", err)
	}

	if len(users.Users) == 0 {
		s.log.Error("No users found in JSON file: %s", s.userListUrl)
	}

	s.log.Info("Users loaded from JSON file: %s -> %v", s.userListUrl, users.Users)

	for _, user := range users.Users {
		if user.PubKey == "" {
			s.log.Debug("User %s has no public key, getting from github", user.Login)
			pubKey, err := s.getPubKeyFromGithub(user.Login)
			if err != nil {
				s.log.Error("Failed to get public key from github: %v", err)
				continue
			}

			if len(pubKey) == 0 {
				s.log.Error("No public key found for user %s", user.Login)
				continue
			}

			user.SetPubKey(pubKey)
			s.log.Debug("User %s public key: %s", user.Login, pubKey)
		}
	}

	s.log.Info("Users loaded from JSON file: %s -> %v", s.userListUrl, users.Users)
	return users, nil
}

func (s *UserService) getPubKeyFromGithub(user string) (string, error) {
	resp, err := http.Get(s.githubUrl + user)
	if err != nil {
		s.log.Error("Failed to get public key from github: %v", err)
		return "", err
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		s.log.Error("Failed to get public key from github: %s - %s, status code: %d", s.githubUrl, user, resp.StatusCode)
		return "", err
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		s.log.Error("Failed to read response body: %v", err)
		return "", err
	}

	return string(body), nil
}
