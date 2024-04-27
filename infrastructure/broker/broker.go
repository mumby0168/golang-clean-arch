package broker

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
)

type MessageProcessorService struct {
	waitGroup     sync.WaitGroup
	subscriptions []Subscription
	client        *azservicebus.Client
	stopSignal    chan os.Signal
}

type Subscription struct {
	TopicName      string
	SubscriberName string
	Handle         func(msg string) HandlerResponse
}

type HandlerResponse interface {
}

type DeadLetter struct {
	Reason      string
	Description string
}

type Abandon struct{}

type Complete struct{}

var (
	_ HandlerResponse = DeadLetter{}
	_ HandlerResponse = Abandon{}
	_ HandlerResponse = Complete{}
)

func NewMessageProcessorService(subscriptions []Subscription, client *azservicebus.Client, stopSignal chan os.Signal) *MessageProcessorService {
	return &MessageProcessorService{
		subscriptions: subscriptions,
		client:        client,
		stopSignal:    stopSignal,
	}
}

func (mps *MessageProcessorService) Start() {

	for _, sub := range mps.subscriptions {

		mps.waitGroup.Add(1)
		go mps.processMessages(&sub)
		mps.waitGroup.Wait()
	}
}

func (mps *MessageProcessorService) processMessages(sub *Subscription) {
	defer mps.waitGroup.Done()

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-mps.stopSignal
		println("shutting down receiver from os shutdown")
		cancel()
	}()

	r, err := mps.client.NewReceiverForSubscription(sub.TopicName, sub.SubscriberName, nil)
	if err != nil {
		panic(err)
	}

messageLoop:
	for {
		select {
		case <-ctx.Done():
			println("shutting down receiver")
			break messageLoop
		default:
			mps.readMessage(r, sub, ctx)
			time.Sleep(1 * time.Second)
		}
	}
}

func (mps *MessageProcessorService) readMessage(receiver *azservicebus.Receiver, sub *Subscription, ctx context.Context) {

	messages, err := receiver.ReceiveMessages(ctx, 5, nil)

	if err != nil {
		if err == context.Canceled {
			return
		}
		panic(err)
	}

	for _, msg := range messages {
		r := sub.Handle(string(msg.Body))

		switch r.(type) {
		case DeadLetter:
			deadLetter := r.(DeadLetter)
			receiver.DeadLetterMessage(ctx, msg, &azservicebus.DeadLetterOptions{
				ErrorDescription: to.Ptr(deadLetter.Description),
				Reason:           to.Ptr(deadLetter.Reason),
			})
		case Abandon:
			receiver.AbandonMessage(ctx, msg, nil)
		case Complete:
			receiver.CompleteMessage(ctx, msg, nil)
		}
	}

}
