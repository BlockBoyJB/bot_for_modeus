package broker

import (
	"bot_for_modeus/internal/model/brokermodel"
	"bot_for_modeus/pkg/rabbitmq"
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
	log "github.com/sirupsen/logrus"
)

const (
	brokerConsumerPrefixLog = "/broker/consumer"
)

type rmqConsumer struct {
	*rabbitmq.RabbitMQ
}

type HandleMessage interface {
	SendDaySchedule(userId int64, messageId int) error
	SendWeekSchedule(userId int64, messageId int) error
	SendUserGrades(userId int64, messageId int) error
	SendSubjectGradesInfo(userId int64, messageId int) error
}

func newConsumer(rmq *rabbitmq.RabbitMQ) *rmqConsumer {
	return &rmqConsumer{rmq}
}

// StartConsume запускается самой последней. Туда передаем уже инициализированный хэндлер бота,
// где инициализированы функции для работы брокера
func (r *rmqConsumer) StartConsume(handler HandleMessage) {
	daySchedule, err := r.InitQueue(dayScheduleQueue)
	if err != nil {
		log.Errorf("%s/StartComsume init day schedule queue error: %s", brokerConsumerPrefixLog, err)
	}
	go r.processMessage(handler.SendDaySchedule, daySchedule)
	weekSchedule, err := r.InitQueue(weekScheduleQueue)
	if err != nil {
		log.Errorf("%s/StartComsume init week schedule queue error: %s", brokerConsumerPrefixLog, err)
	}
	go r.processMessage(handler.SendWeekSchedule, weekSchedule)
	userGrades, err := r.InitQueue(userGradesQueue)
	if err != nil {
		log.Errorf("%s/StartConsume init user grade queue error: %s", brokerConsumerPrefixLog, err)
	}
	go r.processMessage(handler.SendUserGrades, userGrades)
	subjectInfo, err := r.InitQueue(subjectGradesQueue)
	if err != nil {
		log.Errorf("%s/StartConsume init subject grade info queue error: %s", brokerConsumerPrefixLog, err)
	}
	go r.processMessage(handler.SendSubjectGradesInfo, subjectInfo)
}

func (r *rmqConsumer) processMessage(h func(int64, int) error, messages <-chan amqp.Delivery) {
	for msg := range messages {
		var m brokermodel.Message
		if err := json.Unmarshal(msg.Body, &m); err != nil {
			log.Errorf("%s/processMessage decoding message error: %s", brokerConsumerPrefixLog, err)
			continue // если сообщение "неправильное", то даже нет смысла запускать обработчик
		}
		go func() {
			if err := h(m.UserId, m.MessageId); err != nil {
				log.Errorf("%s/processMessage handler func error: %s", brokerConsumerPrefixLog, err)
			}
		}()
	}
}
