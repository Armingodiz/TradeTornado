package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type MockEvent struct {
	Msg      string `json:"msg"`
	HasError bool   `json:"hasError"`
}

func NewMockEvent(msg string) *MockEvent {
	return &MockEvent{Msg: msg}
}

func (e *MockEvent) Marshal() ([]byte, error) {
	return json.Marshal(*e)
}

func (e *MockEvent) Unmarshal(data []byte) error {
	err := json.Unmarshal(data, e)
	if err != nil {
		return err
	}
	return nil
}

func (e *MockEvent) GetName() string {
	return "testEvent"
}

type mockDispatcher struct {
	events []IEvent
}

func NewMockDispatcher() *mockDispatcher {
	return &mockDispatcher{}
}

func (m *mockDispatcher) Dispatch(ctx context.Context, event IEvent, ack AckFunc) error {
	m.events = append(m.events, event)
	return nil
}

func (m *mockDispatcher) ClearEvents() {
	m.events = nil
}

func (m *mockDispatcher) GetEvents() []IEvent {
	return m.events
}

type mockMessagePublisher struct {
	events []IEvent
}

func NewMockMessagePublisher() *mockMessagePublisher {
	return &mockMessagePublisher{}
}

func (m *mockMessagePublisher) Publish(ctx context.Context, event IEvent) error {
	m.events = append(m.events, event)
	return nil
}

func (m *mockMessagePublisher) ClearEvents() {
	m.events = nil
}

func (m *mockMessagePublisher) GetEvents() []IEvent {
	return m.events
}

type mockQueuePublisher struct {
	messages [][]byte
}

func NewMockQueuePublisher() *mockQueuePublisher {
	return &mockQueuePublisher{}
}

func (m *mockQueuePublisher) Send(ctx context.Context, msg []byte) error {
	m.messages = append(m.messages, msg)
	return nil
}

func (m *mockQueuePublisher) ClearMessages() {
	m.messages = nil
}

func (m *mockQueuePublisher) GetMessages() [][]byte {
	return m.messages
}

type EventBusMetricCollectorMock struct{}

func (m *EventBusMetricCollectorMock) IncMetricEventDispatcherErrorCount() {
}
func (m *EventBusMetricCollectorMock) IncMetricEventDispatcherSuccessCount() {
}
func (m *EventBusMetricCollectorMock) IncMetricEventDispatcherEventCount() {
}

func (m *EventBusMetricCollectorMock) GetMetrics() []prometheus.Collector {
	return []prometheus.Collector{}
}

type mockSubscriber struct {
	msg            string
	errorOnDeliver bool
	waitTime       time.Duration
	received       []IEvent
}

func (r *mockSubscriber) Representation() string {
	return "mockSubscriber"
}

func NewMockSubscriber(msg string, errorOnDeliver bool, waitTime time.Duration) *mockSubscriber {
	return &mockSubscriber{msg: msg, errorOnDeliver: errorOnDeliver, received: make([]IEvent, 0), waitTime: waitTime}
}

func (r *mockSubscriber) Subscribe(ctx context.Context, inpEvent IEvent) error {
	event := inpEvent.(*MockEvent)
	logrus.Info("started: ", r.msg, "msg: ", event.Msg)
	time.Sleep(r.waitTime)
	if r.errorOnDeliver || event.HasError {
		return fmt.Errorf("can not process it")
	}

	logrus.Info("end: ", r.msg, "msg: ", event.Msg)
	r.received = append(r.received, event)
	return nil
}

type MockEventBus struct{}

func (e *MockEventBus) Publish(ctx context.Context, event IEvent) error {
	return nil
}
func (e *MockEventBus) Dispatch(ctx context.Context, event IEvent, ack AckFunc) error {
	return nil
}
func (e *MockEventBus) RegisterAtLeastOnce(subscriber ISubscriber) {}

func (e *MockEventBus) RegisterAtMostOnce(subscriber ISubscriber) {}

func NewMockEventBus() *MockEventBus {
	return &MockEventBus{}
}
