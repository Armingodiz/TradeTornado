package provider

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/sirupsen/logrus"
)

type KafkaConsumerConfig struct {
	Brokers    string
	RetryTopic string
	BatchSize  int
}

type KafkaConsumerProvider struct {
	consumer *kafka.Consumer
	producer *KafkaProducerProvider
	cnf      KafkaConsumerConfig
	GroupID  string
	Topic    string
}

func NewKafkaConsumerProvider(cnf KafkaConsumerConfig, pr *KafkaProducerProvider, topic, consumerGroup string) (*KafkaConsumerProvider, error) {
	kafkaProvider, err := NewKafkaConnection(cnf, topic, consumerGroup)
	if err != nil {
		return nil, err
	}
	kafkaProvider.producer = pr
	return kafkaProvider, nil
}

func (receiver *KafkaConsumerProvider) GetRepresentation() string {
	return "KafkaConsumerProvider"
}

func (receiver *KafkaConsumerProvider) Recover(ctx context.Context) error {
	if err := receiver.IsAlive(ctx); err != nil {
		logrus.WithField("error", err.Error()).Errorln("trying to reconnect...")
		kafkaProvider, err := NewKafkaConnection(receiver.cnf, receiver.Topic, receiver.GroupID)
		if err != nil {
			return err
		}
		receiver.consumer = kafkaProvider.consumer
	}
	return nil
}

func (receiver *KafkaConsumerProvider) Run(ctx context.Context) error {
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
			logrus.Infoln("Shutting down Kafka consumer...")
			return nil
		}
	}
}

func (receiver *KafkaConsumerProvider) IsAlive(ctx context.Context) error {
	if receiver.consumer == nil {
		return errors.New("kafka consumer is not alive")
	}
	return nil
}

func NewKafkaConnection(config KafkaConsumerConfig, topic, group string) (*KafkaConsumerProvider, error) {
	provider := KafkaConsumerProvider{cnf: config, Topic: topic, GroupID: group}

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":    config.Brokers,
		"group.id":             group,
		"auto.offset.reset":    "earliest",
		"enable.auto.commit":   false,
		"enable.partition.eof": false,
	})
	fmt.Println(config.Brokers, topic)
	if err != nil {
		return &provider, err
	}
	provider.consumer = consumer

	logrus.WithField("config", fmt.Sprintf("%+v", config)).Infoln("Connected to Kafka successfully...")
	return &provider, nil
}

func (receiver *KafkaConsumerProvider) Consume(ctx context.Context, process func(string) error) error {
	if err := receiver.consumer.SubscribeTopics([]string{receiver.Topic}, nil); err != nil {
		return err
	}
	var wg sync.WaitGroup
	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			return receiver.consumer.Close()
		default:
			batch := receiver.fetchBatch(ctx)
			if len(batch) == 0 {
				continue
			}
			wg.Add(len(batch))
			for _, msg := range batch {
				go func(msg *kafka.Message) {
					defer wg.Done()
					if err := process(string(msg.Value)); err != nil {
						fmt.Println("#### --> here error")
						// TODO: handle produce error
						receiver.producer.ProduceWithKey(ctx, receiver.Topic, string(msg.Key), string(msg.Value))
					}
				}(msg)
			}
			wg.Wait()
			receiver.commitOffsets(batch)
		}
	}
}

func (receiver *KafkaConsumerProvider) fetchBatch(ctx context.Context) []*kafka.Message {
	var batch []*kafka.Message
	for len(batch) < receiver.cnf.BatchSize {
		ev := receiver.consumer.Poll(100)
		if ev == nil {
			break
		}
		switch e := ev.(type) {
		case *kafka.Message:
			batch = append(batch, e)
		case kafka.Error:
			logrus.WithError(e).Error("Error consuming from Kafka")
			if e.Code() == kafka.ErrAllBrokersDown {
				return nil
			}
		default:
			logrus.Infof("Ignored %v\n", e)
		}
	}
	return batch
}

func (receiver *KafkaConsumerProvider) commitOffsets(batch []*kafka.Message) {
	// Kafka prefers batch commit
	offsets := make([]kafka.TopicPartition, len(batch))
	for i, msg := range batch {
		fmt.Println("#### --> here commits", msg.TopicPartition.Offset+1, string(msg.Value))
		offsets[i] = kafka.TopicPartition{
			Topic:     msg.TopicPartition.Topic,
			Partition: msg.TopicPartition.Partition,
			Offset:    msg.TopicPartition.Offset + 1,
		}
	}
	if _, err := receiver.consumer.CommitOffsets(offsets); err != nil {
		logrus.WithError(err).Error("Failed to commit offsets")
	}
}
