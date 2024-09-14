package parser

import (
	"bot_for_modeus/pkg/modeus"
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"time"
)

const (
	tokenPrefix          = "token:"
	defaultParserTimeout = time.Minute
	defaultTTL           = time.Hour*23 + time.Minute*50
)

func (p *parser) rootToken(ctx context.Context) (string, error) {
	return p.parseToken(ctx, p.rootLogin, p.rootPass)
}

func (p *parser) userToken(ctx context.Context, login, password string) (string, error) {
	return p.parseToken(ctx, login, password)
}

func (p *parser) parseToken(ctx context.Context, login, password string) (string, error) {
	token, err := p.redis.Get(ctx, tokenKey(login)).Result()
	if err == nil && token != "" {
		return token, nil
	}
	token, err = p.modeus.ExtractToken(login, password, defaultParserTimeout)
	if err != nil {
		if errors.Is(err, modeus.ErrIncorrectInputData) {
			return "", ErrIncorrectLoginPassword
		}
		log.Errorf("%s/parseToken error parse token from selenium client: %s", parserServicePrefixLog, err)
		return "", err
	}
	if err = p.redis.Set(ctx, tokenKey(login), token, defaultTTL).Err(); err != nil {
		log.Errorf("%s/parseToken error save token into redis: %s", parserServicePrefixLog, err)
		return "", err
	}
	go p.watchToken(login, password, defaultTTL-time.Minute)
	return token, nil
}

// TODO функция продолжит работу даже после того как пользователь перестанет пользоваться ботом

// Такой немного костыльный метод, чтобы обновлять токен до его истечения
// Если этого не сделать, то в промежуток до его обновления селениум просто умрет от количества запросов
func (p *parser) watchToken(login, password string, d time.Duration) {
	for {
		time.Sleep(d)
		token, err := p.modeus.ExtractToken(login, password, defaultParserTimeout)
		if err != nil {
			log.Errorf("%s/watchToken error parse token from selenium client: %p", parserServicePrefixLog, err)
			return // завершаем работу, потому что parseToken в случае ошибки запустит еще одну такую же
		}
		if err = p.redis.Set(context.Background(), tokenKey(login), token, defaultTTL).Err(); err != nil {
			log.Errorf("%s/watchToken error save token into redis: %p", parserServicePrefixLog, err)
			return // завершаем работу, потому что parseToken в случае ошибки запустит еще одну такую же
		}
	}
}

func tokenKey(login string) string {
	return tokenPrefix + login
}
