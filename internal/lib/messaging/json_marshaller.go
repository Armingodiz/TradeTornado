package messaging

import (
	"context"
	"encoding/json"
	"errors"
)

type JsonMarshaller struct {
	busDispatcher  IDispatcher // High Level Event Dispatcher
	registry       *EventRegistry
	queuePublisher IQueuePublisher // Low Level Sender
}

func NewJsonMarshaller(dispatcher IDispatcher, queuePublisher IQueuePublisher) *JsonMarshaller {
	eventRegistry := NewEventRegistry()
	return &JsonMarshaller{busDispatcher: dispatcher, queuePublisher: queuePublisher, registry: eventRegistry}
}

func (j *JsonMarshaller) Publish(ctx context.Context, event IEvent) error {
	if _, ok := j.registry.eventGenerators[event.GetName()]; !ok {
		return errors.New("event not find")
	}

	marshaledEvent, err := event.Marshal()
	if err != nil {
		return err
	}

	msg := message{
		Name: event.GetName(),
		Body: marshaledEvent,
	}

	bytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return j.queuePublisher.Send(ctx, bytes)
}

func (j *JsonMarshaller) Receive(ctx context.Context, msg []byte, ackFunc AckFunc) error {
	message1 := message{}
	if err := json.Unmarshal(msg, &message1); err != nil {
		return err
	}
	generator, ok := j.registry.eventGenerators[message1.Name]
	if !ok {
		return errors.New("Event not find ")
	}
	eventBody := generator()
	if err := eventBody.Unmarshal(message1.Body); err != nil {
		return err
	}
	return j.busDispatcher.Dispatch(ctx, eventBody, ackFunc)
}

type IEventGenerator func() IEvent
type EventRegistry struct {
	eventGenerators map[string]IEventGenerator
}

func NewEventRegistry() *EventRegistry {
	return &EventRegistry{eventGenerators: make(map[string]IEventGenerator)}
}

func (r *EventRegistry) Register(name string, generator IEventGenerator) {
	r.eventGenerators[name] = generator
}

func (m *JsonMarshaller) RegisterEvent(generator IEventGenerator) {
	sample := generator()
	m.registry.Register(sample.GetName(), generator)
}

type message struct {
	Name string `json:"name"`
	Body []byte `json:"body"`
}
