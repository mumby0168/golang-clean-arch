package application

import "sbprocessor/application/contracts"

func HandleHelloWorldMessage(msg *contracts.HelloWorldMessage) error {

	println(msg.Message)

	return nil
}
