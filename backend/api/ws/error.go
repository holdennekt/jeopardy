package ws

import (
	"encoding/json"
)

const ERROR Event = "error"

type errorMessage struct {
	Event   Event        `json:"event"`
	Payload errorPayload `json:"payload"`
}

type errorPayload struct {
	Error string `json:"error"`
}

func NewErrorMessage(err error) errorMessage {
	return errorMessage{
		Event: ERROR,
		Payload: errorPayload{
			Error: err.Error(),
		},
	}
}

func (em errorMessage) ToMessage() Message {
	rawMessage, _ := json.Marshal(em.Payload)
	return Message{
		Event:   em.Event,
		Payload: rawMessage,
	}
}
