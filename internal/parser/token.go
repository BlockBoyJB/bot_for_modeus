package parser

import (
	"fmt"
	"net/http"
)

const (
	deleteTokenUri = "/token"
)

type deleteTokenRequest struct {
	Login string `json:"login"`
}

func (p *parser) DeleteToken(login string) error {
	resp, err := p.makeRequest(http.MethodDelete, deleteTokenUri, deleteTokenRequest{Login: login})
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("parser/DeleteToken unexpected code: %d", resp.StatusCode)
	}
	return nil
}
