package gommunicator

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/google/uuid"
)

func getRequest(action, service, dtID string, incomingService string, payload interface{}, timeout int) (*DataTransactionRequest, error) {
	actionUUID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	actionID := actionUUID.String()

	if timeout == 0 {
		timeout = 5
	}

	return &DataTransactionRequest{
		ID:              dtID,
		ActionID:        &actionID,
		Data:            payload,
		Action:          action,
		Service:         service,
		IncomingService: incomingService,
		Timeout:         timeout,
	}, nil
}

// ExecInput input settings for an action execution
// Timeout is not required, if omitted default timeout will be set to 5 seconds
type ExecInput struct {
	DataTransactionID string
	Action            string
	Service           string
	Payload           interface{}
	Timeout           int
}

// Exec executes an action on the services cluster
func (gom *Gommunicator) Exec(input *ExecInput) (<-chan *DataTransactionResponse, error) {
	// Generate request
	request, err := getRequest(input.Action, input.Service, input.DataTransactionID, gom.ServiceName, input.Payload, input.Timeout)
	if err != nil {
		return nil, err
	}

	// Create receiver channel
	// This is the channel that will receive a possible action's response
	receiver := make(chan *DataTransactionResponse)

	// Generate UUID to prevent duplicates
	// This is necessary because Standard SQS may deliver duplicated messages
	dedupUUID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	request.DedupID = dedupUUID.String()

	// Marshal request to JSON string
	bytesMessage, err := json.Marshal(&request)
	if err != nil {
		return nil, err
	}

	message := string(bytesMessage)

	// Publish SNS message to Orchestrator Topic
	_, err = gom.orchestrator.Publish(
		&sns.PublishInput{
			TopicArn: aws.String(gom.SNSTopicARN),
			Message:  aws.String(message),
			MessageAttributes: map[string]*sns.MessageAttributeValue{
				"Service": {
					StringValue: aws.String(input.Service),
					DataType:    aws.String("String"),
				},
				"Action": {
					StringValue: aws.String(input.Action),
					DataType:    aws.String("String"),
				},
			},
		},
	)

	if err != nil {
		gom.onErr(err)
		return nil, err
	}

	// Creates a new context related to the action req/resp
	// This context has a timeout of expected timeout + 1 second
	// When the context is closed, the request is timed out, by closing the listener goroutine
	// 	and no further response to this action will be handled
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(input.Timeout)*time.Second)

	go func(c context.Context, r chan *DataTransactionResponse, actionID string) {
		for {
			select {
			case <-c.Done():
				r <- nil
				close(r)
				deleteCallback(actionID)
				return
			}
		}
	}(ctx, receiver, *request.ActionID)

	// Register a new response callback
	// This is the callback that will run when a response is received
	registerCallback(
		*request.ActionID,
		func(response *DataTransactionResponse) error {
			receiver <- response
			cancel()
			return nil
		},
	)

	return receiver, nil
}

// Respond sends a response to a DataTransactionRequest
func (gom *Gommunicator) Respond(request *DataTransactionRequest, payload interface{}) error {
	dt := FromRequest(request)
	dt.data = payload

	response := dt.Success("")

	dedupUUID, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	response.DedupID = dedupUUID.String()

	bytesMessage, err := json.Marshal(&response)
	if err != nil {
		return err
	}

	message := string(bytesMessage)

	_, err = gom.orchestrator.Publish(
		&sns.PublishInput{
			TopicArn: aws.String(gom.SNSTopicARN),
			Message:  aws.String(message),
			MessageAttributes: map[string]*sns.MessageAttributeValue{
				"Service": {
					StringValue: aws.String(request.IncomingService),
					DataType:    aws.String("String"),
				},
			},
		},
	)

	return err
}

// RespondError sends a response to a DataTransactionRequest
func (gom *Gommunicator) RespondError(request *DataTransactionRequest, mapErr MapErr) error {
	dt := FromRequest(request)
	response := dt.FailFromMapErr(mapErr)

	dedupUUID, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	response.DedupID = dedupUUID.String()

	bytesMessage, err := json.Marshal(&response)
	if err != nil {
		return err
	}

	message := string(bytesMessage)

	_, err = gom.orchestrator.Publish(
		&sns.PublishInput{
			TopicArn: aws.String(gom.SNSTopicARN),
			Message:  aws.String(message),
			MessageAttributes: map[string]*sns.MessageAttributeValue{
				"Service": {
					StringValue: aws.String(request.IncomingService),
					DataType:    aws.String("String"),
				},
			},
		},
	)

	return err
}
