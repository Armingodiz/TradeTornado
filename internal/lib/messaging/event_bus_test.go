package messaging

import (
	"context"
	"strconv"
	"testing"
	"time"

	"tradeTornado/internal/lib"

	"github.com/stretchr/testify/suite"
)

type EventBusTestSuite struct {
	suite.Suite
}

func TestEventLoopSuite(t *testing.T) {
	suite.Run(t, new(EventBusTestSuite))
}

func (suite *EventBusTestSuite) TestPublish() {

	ctx := lib.Terminable()
	eventBus := NewEventBus()
	publisher := NewMockMessagePublisher()
	eventBus.SetPublisher(publisher)
	for i := 0; i < 10; i++ {
		event := NewMockEvent("hello" + strconv.Itoa(i))
		err := eventBus.Publish(ctx, event)
		suite.Require().Nil(err)
	}

	suite.Require().Equal(len(publisher.GetEvents()), 10)
}

func (suite *EventBusTestSuite) TestHappyRegisterAtLeastOnceReceiverDispatch() {
	ctx := lib.Terminable()
	eventBus := NewEventBus()
	goodReceiver := NewMockSubscriber("good receiver", false, 1*time.Microsecond)
	eventBus.RegisterAtLeastOnce(goodReceiver)
	event := NewMockEvent("message")
	ack := false
	ackFunc := func(ctx context.Context, givenAck bool) error {
		ack = givenAck
		return nil
	}
	eventBus.Dispatch(ctx, event, ackFunc)
	suite.Require().Equal(len(goodReceiver.received), 1)
	suite.Require().Equal(ack, true)
}

func (suite *EventBusTestSuite) TestSadRegisterAtLeastOnceReceiverDispatch() {
	ctx := lib.Terminable()
	eventBus := NewEventBus()
	goodReceiver := NewMockSubscriber("good receiver", false, 1*time.Microsecond)
	badReceiver := NewMockSubscriber("bad receiver", true, 1*time.Microsecond)
	eventBus.RegisterAtLeastOnce(goodReceiver)
	eventBus.RegisterAtLeastOnce(badReceiver)
	// In this case no message received to at most once
	eventBus.RegisterAtMostOnce(goodReceiver)
	event := NewMockEvent("message")
	ack := false
	ackFunc := func(ctx context.Context, givenAck bool) error {
		ack = givenAck
		return nil
	}
	eventBus.Dispatch(ctx, event, ackFunc)
	suite.Require().LessOrEqual(len(goodReceiver.received), 1)
	suite.Require().Equal(len(badReceiver.received), 0)
	suite.Require().Equal(ack, false)
}

func (suite *EventBusTestSuite) TestRegisterAtMostOnceReceiverDispatch() {
	ctx := lib.Terminable()
	eventBus := NewEventBus()
	goodReceiver := NewMockSubscriber("good receiver", false, 1*time.Microsecond)
	badReceiver := NewMockSubscriber("bad receiver", true, 1*time.Microsecond)
	eventBus.RegisterAtMostOnce(goodReceiver)
	eventBus.RegisterAtMostOnce(badReceiver)
	event := NewMockEvent("message")
	ack := false
	ackFunc := func(ctx context.Context, givenAck bool) error {
		ack = givenAck
		return nil
	}
	eventBus.Dispatch(ctx, event, ackFunc)
	suite.Require().Equal(len(goodReceiver.received), 1)
	suite.Require().Equal(len(badReceiver.received), 0)

	suite.Require().Equal(ack, true)
}
