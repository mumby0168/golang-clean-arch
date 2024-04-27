package infrastructure

import (
	"encoding/json"
	"sbprocessor/application"
	"sbprocessor/application/contracts"
	"sbprocessor/infrastructure/broker"
)

func HandleHelloWorldMessage(msg string) broker.HandlerResponse {

	var helloWorldMessage contracts.HelloWorldMessage
	err := json.Unmarshal([]byte(msg), &helloWorldMessage)
	if err != nil {
		return broker.DeadLetter{
			Reason:      "Failed to unmarshal message",
			Description: err.Error(),
		}
	}

	err = helloWorldMessage.Validate()

	if err != nil {
		return broker.DeadLetter{
			Reason:      "Failed to validate message",
			Description: err.Error(),
		}
	}

	err = application.HandleHelloWorldMessage(&helloWorldMessage)

	if err != nil {
		return broker.DeadLetter{}
	}

	return broker.Complete{}
}
