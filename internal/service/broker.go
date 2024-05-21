package service

import (
	"bot_for_modeus/internal/broker"
	"bot_for_modeus/internal/model/brokermodel"
	"context"
)

type BrokerService struct {
	rmq *broker.Broker
}

func newBrokerService(rmq *broker.Broker) *BrokerService {
	return &BrokerService{rmq: rmq}
}

func (s *BrokerService) RequestDaySchedule(ctx context.Context, userId int64, messageId int) error {
	return s.rmq.Producer.DaySchedule(ctx, brokermodel.Message{
		UserId:    userId,
		MessageId: messageId,
	})
}

func (s *BrokerService) RequestWeekSchedule(ctx context.Context, userId int64, messageId int) error {
	return s.rmq.Producer.WeekSchedule(ctx, brokermodel.Message{
		UserId:    userId,
		MessageId: messageId,
	})
}

func (s *BrokerService) RequestUserGrades(ctx context.Context, userId int64, messageId int) error {
	return s.rmq.Producer.UserGrades(ctx, brokermodel.Message{
		UserId:    userId,
		MessageId: messageId,
	})
}

func (s *BrokerService) RequestSubjectInfo(ctx context.Context, userId int64, messageId int) error {
	return s.rmq.Producer.SubjectInfo(ctx, brokermodel.Message{
		UserId:    userId,
		MessageId: messageId,
	})
}
