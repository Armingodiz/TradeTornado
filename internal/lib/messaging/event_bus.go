package messaging

import (
	"context"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
)

type EventBus struct {
	publisher   IMessagePublisher
	atMostOnce  []ISubscriber
	atLeastOnce []ISubscriber
}

func (e *EventBus) Publish(ctx context.Context, event IEvent) error {
	metricEventBusEventCount.WithLabelValues("nats").Inc()
	return e.publisher.Publish(ctx, event)
}

func (e *EventBus) Dispatch(ctx context.Context, event IEvent, ack AckFunc) error {
	// TODO: check and add this:
	// defer func() {
	// 	if err := recover(); err != nil {
	// 		// put out the events that cause panic
	// 		// if err := ack(ctx, true); err != nil {
	// 		// 	logrus.Errorln(err)
	// 		// }
	// 		fmt.Printf("RECOVERED: %v\n", err)
	// 	}
	// }()
	for _, subscriber := range e.atLeastOnce {
		if err := subscriber.Subscribe(ctx, event); err != nil {
			bts, err2 := event.Marshal()
			if err2 != nil {
				logrus.Errorln(err2)
			} else {
				logrus.Errorln(fmt.Sprintf("Error: %s for event: %s", err.Error(), string(bts)))
			}
			metricEventBusErrorCount.WithLabelValues("nats", event.GetName()).Inc()
			if err := ack(ctx, false); err != nil {
				logrus.Errorln(err)
			}
			logrus.WithFields(logrus.Fields{
				"event":      event.GetName(),
				"subscriber": subscriber.Representation(),
			}).Errorln(err)
			return err
		}
	}

	if err := ack(ctx, true); err != nil {
		logrus.Errorln(err)
	}

	//var wg1 sync.WaitGroup
	//errorCh := make(chan error)
	//for _, subscriber := range e.atLeastOnce {
	//	wg1.Add(1)
	//	go func(ctx context.Context, subscriber ISubscriber) {
	//		defer wg1.Done()
	//		if err := subscriber.Subscribe(ctx, event); err != nil {
	//			logrus.Errorln(err)
	//			metricEventBusErrorCount.WithLabelValues("nats").Inc()
	//			errorCh <- err // Send error to the error channel
	//		}
	//
	//	}(ctx, subscriber)
	//}
	//select {
	//case err := <-errorCh:
	//	if err := ack(ctx, false); err != nil {
	//		return err
	//	}
	//	return err
	//case <-waitGroupDone(&wg1):
	//	// All goroutines completed successfully
	//	if err := ack(ctx, true); err != nil {
	//		return err
	//	}
	//}
	//close(errorCh)

	var wg2 sync.WaitGroup
	for _, subscriber := range e.atMostOnce {
		wg2.Add(1)
		go func(ctx context.Context, subscriber ISubscriber) {
			defer wg2.Done()
			if err := subscriber.Subscribe(ctx, event); err != nil {
				metricEventBusErrorCount.WithLabelValues("nats", event.GetName()).Inc()
				logrus.Error(err)
			}

		}(ctx, subscriber)
	}
	wg2.Wait()
	metricEventBusSuccessCount.WithLabelValues("nats").Inc()
	return nil
}

func (e *EventBus) RegisterAtLeastOnce(subscriber ISubscriber) {
	e.atLeastOnce = append(e.atLeastOnce, subscriber)
}

func (e *EventBus) RegisterAtMostOnce(subscriber ISubscriber) {
	e.atMostOnce = append(e.atMostOnce, subscriber)
}

func NewEventBus() *EventBus {
	return &EventBus{}
}

func (e *EventBus) SetPublisher(publisher IMessagePublisher) {
	e.publisher = publisher
}

// Helper function to wait until all goroutines in the WaitGroup finish
func waitGroupDone(wg *sync.WaitGroup) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	return done
}
