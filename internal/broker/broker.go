package broker

import (
	"bot_for_modeus/internal/model/brokermodel"
	"bot_for_modeus/pkg/rabbitmq"
	"context"
)

const (
	dayScheduleQueue   = "day_schedule"
	weekScheduleQueue  = "week_schedule"
	userGradesQueue    = "user_grades"
	subjectGradesQueue = "subject_grades_info"
)

type (
	producer interface {
		DaySchedule(ctx context.Context, m brokermodel.Message) error
		WeekSchedule(ctx context.Context, m brokermodel.Message) error
		UserGrades(ctx context.Context, m brokermodel.Message) error
		SubjectInfo(ctx context.Context, m brokermodel.Message) error
	}
	consumer interface {
		StartConsume(handler HandleMessage)
	}
)

type Broker struct {
	Producer producer
	Consumer consumer
}

func NewBroker(rmq *rabbitmq.RabbitMQ) *Broker {
	return &Broker{
		Producer: newProducer(rmq),
		Consumer: newConsumer(rmq),
	}
}
