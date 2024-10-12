package parser

import (
	"bot_for_modeus/pkg/modeus"
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
)

const (
	tokenPrefix = "token:"
)

func (p *parser) rootToken(ctx context.Context) (string, error) {
	return p.parseToken(ctx, p.rootLogin, p.rootPass)
}

func (p *parser) userToken(ctx context.Context, login, password string) (string, error) {
	return p.parseToken(ctx, login, password)
}

func (p *parser) parseToken(ctx context.Context, login, password string) (string, error) {
	if t, err := p.redis.Get(ctx, tokenKey(login)).Result(); err == nil && t != "" {
		return t, nil
	}
	t, err := p.modeus.GetToken(login, password)
	if err != nil {
		if errors.Is(err, modeus.ErrIncorrectInputData) {
			return "", ErrIncorrectLoginPassword
		}
		log.Errorf("%s/parseToken error get token from service: %s", parserServicePrefixLog, err)
		return "", err
	}
	return t, nil
}

func (p *parser) DeleteToken(login string) error {
	if err := p.modeus.DeleteToken(login); err != nil {
		log.Errorf("%s/DeleteToken error request to delete token from service: %s", parserServicePrefixLog, err)
		return err
	}
	return nil
}

func tokenKey(login string) string {
	return tokenPrefix + login
}
