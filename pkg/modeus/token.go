package modeus

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultTokenServiceTimeout = time.Minute
)

type TokenService struct {
	defaultUrl string
	client     *http.Client
}

func NewTokenService(url string) *TokenService {
	return &TokenService{
		defaultUrl: url,
		client: &http.Client{
			Timeout: defaultTokenServiceTimeout,
		},
	}
}

type getTokenInput struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (s *TokenService) GetToken(login, password string) (string, error) {
	r, err := s.makeRequest(http.MethodGet, "", getTokenInput{
		Login:    login,
		Password: password,
	})
	if err != nil {
		return "", err
	}
	defer func() { _ = r.Body.Close() }()

	if err = handleError(r); err != nil {
		return "", err
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}

	var response struct {
		Token string `json:"token"`
	}
	if err = json.Unmarshal(body, &response); err != nil {
		return "", err
	}
	return response.Token, nil
}

type deleteTokenInput struct {
	Login string `json:"login"`
}

func (s *TokenService) DeleteToken(login string) error {
	r, err := s.makeRequest(http.MethodDelete, "", deleteTokenInput{
		Login: login,
	})
	if err != nil {
		return err
	}
	return handleError(r)
}

func (s *TokenService) makeRequest(method, uri string, v any) (*http.Response, error) {
	body, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	r, err := http.NewRequest(method, s.defaultUrl+uri, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	r.Header.Set("Content-Type", "application/json")

	return s.client.Do(r)
}

func handleError(r *http.Response) error {
	switch r.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusBadRequest:
		return ErrIncorrectInputData
	case http.StatusInternalServerError:
		return errors.New("token service internal service error")
	default:
		return fmt.Errorf("unexpected code from response: %d", r.StatusCode)
	}
}
