package bot

import (
	"bot_for_modeus/config"
	v2 "bot_for_modeus/internal/handler/v2"
	"bot_for_modeus/internal/metrics"
	"bot_for_modeus/internal/model/tgmodel"
	"bot_for_modeus/internal/repo"
	"bot_for_modeus/internal/service"
	"bot_for_modeus/pkg/bot"
	"bot_for_modeus/pkg/crypter"
	"bot_for_modeus/pkg/mongo"
	"bot_for_modeus/pkg/redis"
	"context"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func Run() {
	ctx := context.Background()
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("config init error")
	}
	logger := setLogger(cfg.Log.Level, cfg.Log.Output)

	// Initializing database
	mongodb, err := mongo.NewMongo(ctx, cfg.MongoDB.Url, cfg.MongoDB.DB)
	if err != nil {
		log.Fatal().Err(err).Msg("database init error")
	}
	defer mongodb.Disconnect()

	// redis database
	rdb := redis.NewRedis(cfg.Redis.Url)
	defer rdb.Close()

	d := &service.ServicesDependencies{
		Repos:      repo.NewRepositories(mongodb),
		Crypter:    crypter.NewCrypter(cfg.Crypter.Secret),
		ParserHost: cfg.Parser.Host,
	}
	services := service.NewServices(d)

	s := &bot.Settings{
		Token:     cfg.Bot.Token,
		IsWebhook: cfg.Bot.IsWebhook,
		Ctx:       ctx,
	}
	// tg client
	b, err := bot.NewBot(s, bot.SetCommands(tgmodel.UICommands), bot.RedisStorage(ctx, rdb.Conn()), bot.SetLogger(logger))
	if err != nil {
		log.Fatal().Err(err).Msg("tg client init error")
	}
	v2.NewHandler(b, services)
	go b.ListenAndServe()

	go func() {
		if err = metrics.Listen(net.JoinHostPort("", "8082")); err != nil {
			log.Fatal().Err(err).Msg("metrics error")
		}
	}()

	log.Info().Msg("all services are running!")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-interrupt

	b.Shutdown()
	log.Info().Msg("bot shutdown with exit code 0")
}

// loading environment params from .env
func init() {
	if _, ok := os.LookupEnv("BOT_TOKEN"); !ok {
		if err := godotenv.Load(); err != nil {
			log.Fatal().Err(err).Msg("load env file error")
		}
	}
}
