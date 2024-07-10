package bot

import (
	"bot_for_modeus/config"
	v2 "bot_for_modeus/internal/handler/v2"
	"bot_for_modeus/internal/model/tgmodel"
	"bot_for_modeus/internal/repo"
	"bot_for_modeus/internal/service"
	"bot_for_modeus/pkg/bot"
	"bot_for_modeus/pkg/crypter"
	"bot_for_modeus/pkg/modeus"
	"bot_for_modeus/pkg/mongo"
	"bot_for_modeus/pkg/redis"
	"context"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

func Run() {
	ctx := context.Background()
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("config init error: %s", err)
	}
	setLogger(cfg.Log.Level, cfg.Log.Output)

	// Initializing database
	mongodb, err := mongo.NewMongo(ctx, cfg.MongoDB.Url, cfg.MongoDB.DB, cfg.MongoDB.Collection)
	if err != nil {
		log.Fatalf("database init error: %s", err)
	}
	defer mongodb.Disconnect()

	// selenium for modeus parser
	selenium, err := modeus.NewSeleniumFromConfig(cfg.Selenium.ClientMode, cfg.Selenium.Url, cfg.Selenium.LocalPath)
	if err != nil {
		log.Fatalf("selenium init error: %s", err)
	}
	defer selenium.CloseClient()

	// redis database
	rdb := redis.NewRedis(cfg.Redis.Url, redis.MaxPoolSize(cfg.Redis.MaxPoolSize))
	defer rdb.Close()

	d := service.ServicesDependencies{
		Repos:     repo.NewRepositories(mongodb),
		Parser:    modeus.NewModeus(selenium),
		Redis:     rdb,
		Crypter:   crypter.NewPasswordCrypter(cfg.Crypter.Secret),
		RootLogin: cfg.Root.Login,
		RootPass:  cfg.Root.Password,
	}
	services := service.NewServices(d)

	s := bot.Settings{
		Token:     cfg.Bot.Token,
		IsWebhook: cfg.Bot.IsWebhook,
		Redis:     rdb.Client,
		Ctx:       ctx,
	}
	// tg client
	b, err := bot.NewBot(s, bot.SetCommands(tgmodel.UICommands))
	if err != nil {
		log.Fatalf("tg client init error: %s", err)
	}
	v2.NewRouter(b, services)
	go b.ListenAndServe()

	log.Info("all services are running!")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-interrupt

	b.Shutdown()
	log.Infof("Bot shutdown with exit code 0")
}

// loading environment params from .env
func init() {
	if _, ok := os.LookupEnv("BOT_TOKEN"); !ok {
		if err := godotenv.Load(); err != nil {
			log.Fatalf("load env file error: %s", err)
		}
	}
}
