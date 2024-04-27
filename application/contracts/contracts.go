package contracts

import (
	validation "github.com/go-ozzo/ozzo-validation"
)

type HelloWorldMessage struct {
	Message string `json:"message"`
}

func (m *HelloWorldMessage) Validate() error {

	return validation.ValidateStruct(m,
		validation.Field(&m.Message, validation.Required),
	)

}
