package main

import (
	"os"
	"os/signal"
	"sbprocessor/infrastructure"
	"sbprocessor/infrastructure/broker"
	"syscall"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
)

func main() {

	connectionString := os.Getenv("SB_CS")

	shutdownRequestChannel := make(chan os.Signal, 1)
	signal.Notify(shutdownRequestChannel, os.Interrupt, syscall.SIGTERM)

	client, err := azservicebus.NewClientFromConnectionString(connectionString, nil)

	if err != nil {
		panic(err)
	}

	subscriptions := []broker.Subscription{
		{
			TopicName:      "my-topic",
			SubscriberName: "my-subscription",
			Handle:         infrastructure.HandleHelloWorldMessage,
		},
	}

	mps := broker.NewMessageProcessorService(subscriptions, client, shutdownRequestChannel)

	mps.Start()
}
