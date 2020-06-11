package gommunicator

import (
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type receivedData struct {
	Party string `json:"party"`
}

func TestHandleMessage(t *testing.T) {
	gom := Gommunicator{}

	receiver := make(chan *receivedData, 1)
	message := sqs.Message{
		Body: aws.String(`{"party": "out of sight"}`),
	}

	err := gom.handleRawMessage(&message, receiver)
	if err != nil {
		t.Fatal(err)
	}

	timedOut := make(chan bool)

	go func(tt *testing.T) {
		time.Sleep(5 * time.Second)
		timedOut <- true
	}(t)

Loop:
	for {
		select {
		case response := <-receiver:
			expected := receivedData{
				Party: "out of sight",
			}
			if !reflect.DeepEqual(*response, expected) {
				t.Fatal("wrong response.")
			}
			break Loop
		case timed := <-timedOut:
			if timed {
				t.Fatal("timed out.")
			}
		}
	}
}
