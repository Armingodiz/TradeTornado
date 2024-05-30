package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"sync"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/spf13/cast"
)

var cfg Config

type Order struct {
	OrderID  int       `json:"orderID"`
	Price    int       `json:"price"`
	Quantity int       `json:"quantity"`
	Side     OrderSide `json:"side"`
}

type OrderSide string

const (
	SellOrderSide OrderSide = "sell"
	BuyOrderSide  OrderSide = "buy"
)

type OrderIDGenerator struct {
	counter int
	lock    sync.Mutex
}

func (g *OrderIDGenerator) getOrderID() int {
	g.lock.Lock()
	defer g.lock.Unlock()
	g.counter++
	return g.counter
}

func produceOrder(producer *kafka.Producer, order Order, topic string, wg *sync.WaitGroup) {
	defer wg.Done()

	orderBytes, err := json.Marshal(order)
	if err != nil {
		fmt.Printf("Failed to marshal order: %v\n", err)
		return
	}

	partitionKey := strconv.Itoa(order.OrderID % 2)
	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            []byte(partitionKey),
		Value:          orderBytes,
	}
	fmt.Println(partitionKey)
	err = producer.Produce(msg, nil)
	if err != nil {
		fmt.Printf("Failed to produce message: %v\n", err)
	} else {
		fmt.Println("success")
	}
}

func worker(id int, orders <-chan Order, producer *kafka.Producer, topic string, wg *sync.WaitGroup) {
	for order := range orders {
		produceOrder(producer, order, topic, wg)
	}
}

func main() {
	cfg = configFromEnv()

	producer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": cfg.Broker})
	if err != nil {
		panic(err)
	}
	defer producer.Close()

	var wg sync.WaitGroup

	orderIDGen := &OrderIDGenerator{counter: 0}
	orders := make(chan Order, cfg.NumOrders)

	for i := 0; i < cfg.NumWorkers; i++ {
		go worker(i, orders, producer, cfg.Topic, &wg)
	}

	go func() {
		for i := 0; i < cfg.NumOrders; i++ {
			wg.Add(1)
			orders <- createRandomOrder(orderIDGen)
		}
		close(orders)
	}()
	wg.Wait()
	for e := range producer.Events() {
		switch ev := e.(type) {
		case *kafka.Message:
			if ev.TopicPartition.Error != nil {
				fmt.Printf("Failed to deliver message: %v\n", ev.TopicPartition.Error)
			} else {
				fmt.Printf("Message delivered to topic %s, partition %d, offset %d\n",
					*ev.TopicPartition.Topic, ev.TopicPartition.Partition, ev.TopicPartition.Offset)
			}
		}
	}
	producer.Flush(150000)
}

type Config struct {
	Broker      string
	Topic       string
	NumWorkers  int
	NumOrders   int
	MinPrice    int
	MaxPrice    int
	MinQuantity int
	MaxQuantity int
}

func getEnv(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func configFromEnv() Config {
	return Config{
		Broker:      getEnv("BROKER", "localhost:29092"),
		Topic:       getEnv("TOPIC", "order-events"),
		NumWorkers:  cast.ToInt(getEnv("NUM_WORKERS", "5")),
		NumOrders:   cast.ToInt(getEnv("NUM_ORDERS", "10")),
		MinPrice:    cast.ToInt(getEnv("MIN_PRICE", "1")),
		MaxPrice:    cast.ToInt(getEnv("MAX_PRICE", "10")),
		MinQuantity: cast.ToInt(getEnv("MIN_QUANTITY", "1")),
		MaxQuantity: cast.ToInt(getEnv("MAX_QUANTITY", "20")),
	}
}

func createRandomOrder(orderIDGen *OrderIDGenerator) Order {
	return Order{
		OrderID:  orderIDGen.getOrderID(),
		Price:    rand.Intn(cfg.MaxPrice-cfg.MinPrice) + cfg.MinPrice,
		Quantity: rand.Intn(cfg.MaxQuantity-cfg.MinQuantity) + cfg.MinQuantity,
		Side:     []OrderSide{BuyOrderSide, SellOrderSide}[rand.Intn(2)],
	}
}
