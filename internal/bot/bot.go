package bot

import (
	"bot_for_modeus/config"
	"bot_for_modeus/internal/repo"
	"bot_for_modeus/internal/service"
	"bot_for_modeus/internal/tg/client"
	"bot_for_modeus/internal/tg/handler"
	"bot_for_modeus/pkg/parser"
	"bot_for_modeus/pkg/postgres"
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
	log.Info("Initializing database...")
	pg, err := postgres.New(cfg.PG.Url)
	if err != nil {
		log.Fatalf("database init error: %s", err)
	}
	defer pg.Close()

	// Initializing broker
	//log.Info("Initializing broker...")
	//rmq, err := rabbitmq.New(cfg.Broker.Url)
	//if err != nil {
	//	log.Fatalf("broker init error: %s", err)
	//}
	//defer rmq.Close()

	//rmqBroker := broker.NewBroker(rmq)

	// Local selenium client
	log.Info("Initializing local selenium client...")
	selenium, err := parser.NewLocalClient(cfg.Selenium.LocalUrl)
	defer selenium.CloseLocalClient()
	if err != nil {
		log.Fatalf("parser local init error: %s", err)
	}

	d := service.ServicesDependencies{
		Repos: repo.NewRepositories(pg),
		//Parser: parser.NewClient(cfg.Selenium.Url), // remote client
		Parser: selenium, // local selenium client
		//Rabbit:    rmqBroker,
		RootLogin: cfg.Root.Login,
		RootPass:  cfg.Root.Password,
	}
	services := service.NewServices(d)

	// Инициализируем хэндлер в явном виде, чтобы была возможность оборачивать его в мидлвари
	clientHandler := client.ProcessingMessage

	// Initializing tg client
	log.Info("Initializing tg client...")
	tgClient, err := client.NewClient(cfg.Bot.Token, clientHandler)
	if err != nil {
		log.Fatalf("tg client init error: %s", err)
	}

	h := handler.NewHandler(ctx, tgClient, services) // incoming message handler
	go tgClient.ListenAndServe(h)                    // start listening updates

	//go rmqBroker.Consumer.StartConsume(h) // start consumer for broker messages
	log.Info("all services are running!")

	// TODO добавлю graceful shutdown в будущем, пока просто заглушка
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	<-interrupt
}

// loading environment params from .env
func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("load env file error: %s", err)
	}
}
