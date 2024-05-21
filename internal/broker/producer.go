package broker

import (
	"bot_for_modeus/internal/model/brokermodel"
	"bot_for_modeus/pkg/rabbitmq"
	"context"
)

type rmqProducer struct {
	*rabbitmq.RabbitMQ
}

func newProducer(rmq *rabbitmq.RabbitMQ) *rmqProducer {
	return &rmqProducer{rmq}
}

func (r *rmqProducer) DaySchedule(ctx context.Context, m brokermodel.Message) error {
	return r.PublishMessage(ctx, dayScheduleQueue, m)
}

func (r *rmqProducer) WeekSchedule(ctx context.Context, m brokermodel.Message) error {
	return r.PublishMessage(ctx, weekScheduleQueue, m)
}

func (r *rmqProducer) UserGrades(ctx context.Context, m brokermodel.Message) error {
	return r.PublishMessage(ctx, userGradesQueue, m)
}

func (r *rmqProducer) SubjectInfo(ctx context.Context, m brokermodel.Message) error {
	return r.PublishMessage(ctx, subjectGradesQueue, m)
}
