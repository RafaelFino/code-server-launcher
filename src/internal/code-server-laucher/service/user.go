package service

import (
	"code-server-launcher/internal/code-server-laucher/config"
	"code-server-launcher/internal/code-server-laucher/domain"
	"code-server-launcher/internal/code-server-laucher/logger"
	"encoding/json"
	"io"
	"net/http"
)

type UserService struct {
	githubUrl   string
	userListUrl string
}

var log = logger.Logger

func NewUserService(config *config.Config) *UserService {
	return &UserService{
		githubUrl:   config.GithubUrl,
		userListUrl: config.UserListUrl,
	}
}

func (s *UserService) LoadUsers() (*domain.UserList, error) {
	resp, err := http.Get(s.userListUrl)
	if err != nil {
		log.Errorf("[UserService.LoadUsers] Failed to get JSON file: %v from %s", err, s.userListUrl)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Errorf("[UserService.LoadUsers] Fail to get JSON file: %s, status code: %d", s.userListUrl, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("[UserService.LoadUsers] Failed to read response body: %v", err)
	}

	users := domain.NewUserList()

	err = json.Unmarshal(body, &users)

	if err != nil {
		log.Errorf("[UserService.LoadUsers] Failed to unmarshal JSON: %v", err)
	}

	if len(users.Users) == 0 {
		log.Errorf("[UserService.LoadUsers] No users found in JSON file: %s", s.userListUrl)
	}

	log.Printf("[UserService.LoadUsers] Users loaded from JSON file: %s -> %v", s.userListUrl, users.Users)

	for _, user := range users.Users {
		if user.PubKey == "" {
			log.Debugf("[UserService.LoadUsers] User %s has no public key, getting from github", user.Login)
			pubKey, err := s.getPubKeyFromGithub(user.Login)
			if err != nil {
				log.Errorf("[UserService.LoadUsers] Failed to get public key from github: %v", err)
				continue
			}

			if len(pubKey) == 0 {
				log.Errorf("[UserService.LoadUsers] No public key found for user %s", user.Login)
				continue
			}

			user.SetPubKey(pubKey)
			log.Debugf("[UserService.LoadUsers] User %s public key: %s", user.Login, pubKey)
		}
	}
}

func (s *UserService) getPubKeyFromGithub(user string) (string, error) {
	resp, err := http.Get(s.githubUrl + user)
	if err != nil {
		log.Errorf("[UserService.getPubKeyFromGithub] Failed to get public key from github: %v", err)
		return "", err
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Errorf("[UserService.getPubKeyFromGithub] Failed to get public key from github: %s, status code: %d", s.githubUrl+user, resp.StatusCode)
		return "", err
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Errorf("[UserService.getPubKeyFromGithub] Failed to read response body: %v", err)
		return "", err
	}

	return string(body), nil
}
