package provider

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/sirupsen/logrus"
)

type KafkaProducerConfig struct {
	Brokers string
}

type KafkaProducerProvider struct {
	Producer *kafka.Producer
	cnf      KafkaProducerConfig
}

func (receiver *KafkaProducerProvider) GetRepresentation() string {
	return "KafkaProducerProvider"
}

func (receiver *KafkaProducerProvider) Recover(ctx context.Context) error {
	if err := receiver.IsAlive(ctx); err != nil {
		logrus.WithField("error", err.Error()).Errorln("trying to reconnect...")
		kafkaProvider, err := NewKafkaProducer(receiver.cnf)
		if err != nil {
			return err
		}
		receiver.Producer = kafkaProvider.Producer
	}
	return nil
}

func (receiver *KafkaProducerProvider) Run(ctx context.Context) error {
	if err := receiver.Recover(ctx); err != nil {
		return err
	}
	for {
		select {
		case <-time.After(time.Minute):
			if err := receiver.Recover(ctx); err != nil {
				logrus.Errorln(err)
			}
		case <-ctx.Done():
			logrus.Infoln("Shutting down Kafka producer...")
			if receiver.Producer != nil {
				receiver.Producer.Close()
			}
			return nil
		}
	}
}

func (receiver *KafkaProducerProvider) IsAlive(ctx context.Context) error {
	if receiver.Producer == nil {
		return errors.New("kafka producer is not alive")
	}
	return nil
}

func NewKafkaProducer(config KafkaProducerConfig) (*KafkaProducerProvider, error) {
	provider := KafkaProducerProvider{cnf: config}

	producer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": config.Brokers})
	if err != nil {
		return &provider, err
	}
	provider.Producer = producer
	logrus.WithField("config", fmt.Sprintf("%+v", config)).Infoln("Connected to Kafka producer successfully...")
	return &provider, nil
}

func (receiver *KafkaProducerProvider) Produce(ctx context.Context, topic, message string) error {
	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          []byte(message),
	}
	err := receiver.Producer.Produce(msg, nil)
	if err != nil {
		return err
	}
	return nil
}

func (receiver *KafkaProducerProvider) ProduceWithKey(ctx context.Context, topic, key, message string) error {
	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            []byte(key),
		Value:          []byte(message),
	}
	err := receiver.Producer.Produce(msg, nil)
	if err != nil {
		return err
	}
	return nil
}
