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
	GroupID    string
	Topics     []string
	RetryTopic string
	BatchSize  int
}

type KafkaConsumerProvider struct {
	Consumer *kafka.Consumer
	Producer *kafka.Producer
	cnf      KafkaConsumerConfig
}

func NewKafkaConsumerProvider(cnf KafkaConsumerConfig) (*KafkaConsumerProvider, error) {
	kafkaProvider, err := NewKafkaConnection(cnf)
	if err != nil {
		return nil, err
	}
	return kafkaProvider, nil
}

func (receiver *KafkaConsumerProvider) GetRepresentation() string {
	return "KafkaConsumerProvider"
}

func (receiver *KafkaConsumerProvider) Recover(ctx context.Context) error {
	if err := receiver.IsAlive(ctx); err != nil {
		logrus.WithField("error", err.Error()).Errorln("trying to reconnect...")
		kafkaProvider, err := NewKafkaConnection(receiver.cnf)
		if err != nil {
			return err
		}
		receiver.Consumer = kafkaProvider.Consumer
		receiver.Producer = kafkaProvider.Producer
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
			if receiver.Producer != nil {
				receiver.Producer.Close()
			}
			return nil
		}
	}
}

func (receiver *KafkaConsumerProvider) IsAlive(ctx context.Context) error {
	if receiver.Consumer == nil || receiver.Producer == nil {
		return errors.New("kafka consumer or producer is not alive")
	}
	return nil
}

func NewKafkaConnection(config KafkaConsumerConfig) (*KafkaConsumerProvider, error) {
	provider := KafkaConsumerProvider{cnf: config}

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":    config.Brokers,
		"group.id":             config.GroupID,
		"auto.offset.reset":    "earliest",
		"enable.auto.commit":   false,
		"enable.partition.eof": false,
	})
	fmt.Println(config.Brokers, config.Topics)
	if err != nil {
		return &provider, err
	}

	producer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": config.Brokers})
	if err != nil {
		return &provider, err
	}

	provider.Consumer = consumer
	provider.Producer = producer

	logrus.WithField("config", fmt.Sprintf("%+v", config)).Infoln("Connected to Kafka successfully...")
	return &provider, nil
}

func (receiver *KafkaConsumerProvider) Produce(topic, key, message string) error {
	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            []byte(key),
		Value:          []byte(message),
	}
	return receiver.Producer.Produce(msg, nil)
}

func (receiver *KafkaConsumerProvider) Consume(ctx context.Context, process func(string) error) error {
	if err := receiver.Consumer.SubscribeTopics(receiver.cnf.Topics, nil); err != nil {
		return err
	}
	var wg sync.WaitGroup
	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			return receiver.Consumer.Close()
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
						receiver.Produce(receiver.cnf.RetryTopic, string(msg.Key), string(msg.Value))
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
		ev := receiver.Consumer.Poll(100)
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
	offsets := make([]kafka.TopicPartition, len(batch))
	for i, msg := range batch {
		fmt.Println("$$$$$$--->", msg.TopicPartition.Offset+1, msg.TopicPartition.Partition)
		offsets[i] = kafka.TopicPartition{
			Topic:     msg.TopicPartition.Topic,
			Partition: msg.TopicPartition.Partition,
			Offset:    msg.TopicPartition.Offset + 1,
		}
	}

	if _, err := receiver.Consumer.CommitOffsets(offsets); err != nil {
		logrus.WithError(err).Error("Failed to commit offsets")
	}
}
