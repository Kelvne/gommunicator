package gommunicator

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/google/uuid"
)

// Exec executes an action on the services cluster
// It receives a DataTransactionRequest as parameter and the timeout policy in seconds
func (gom *Gommunicator) Exec(data *DataTransactionRequest, timeout int) (<-chan *DataTransactionResponse, error) {
	receiver := make(chan *DataTransactionResponse)

	dedupUUID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	data.DedupID = dedupUUID.String()

	bytesMessage, err := json.Marshal(&data)
	if err != nil {
		return nil, err
	}

	message := string(bytesMessage)

	now := string(time.Now().UnixNano())

	_, err = gom.orchestrator.Publish(
		&sns.PublishInput{
			TopicArn: aws.String(gom.SNSTopicARN),
			Message:  aws.String(message),
			MessageAttributes: map[string]*sns.MessageAttributeValue{
				"Timestamp": {
					StringValue: aws.String(now),
					DataType:    aws.String("String"),
				},
				"Service": {
					StringValue: aws.String(data.Service),
					DataType:    aws.String("String"),
				},
				"Action": {
					StringValue: aws.String(data.Action),
					DataType:    aws.String("String"),
				},
			},
		},
	)

	if err != nil {
		gom.errorHandler(err)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)

	go func(c context.Context, r chan *DataTransactionResponse) {
		for {
			select {
			case <-c.Done():
				r <- nil
				close(r)
				return
			}
		}
	}(ctx, receiver)

	registerCallback(
		*data.ActionID,
		func(response *DataTransactionResponse) error {
			receiver <- response
			cancel()
			return ctx.Err()
		},
	)

	return receiver, nil
}

// Respond reponds a DataTransactionRequest
func (gom *Gommunicator) Respond(request *DataTransactionRequest, payload interface{}) error {
	dt := FromRequest(request)
	dt.data = payload
	dt.action = nil

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

	now := string(time.Now().UnixNano())

	_, err = gom.orchestrator.Publish(
		&sns.PublishInput{
			TopicArn: aws.String(gom.SNSTopicARN),
			Message:  aws.String(message),
			MessageAttributes: map[string]*sns.MessageAttributeValue{
				"Timestamp": {
					StringValue: aws.String(now),
					DataType:    aws.String("String"),
				},
				"Service": {
					StringValue: aws.String(request.IncomingService),
					DataType:    aws.String("String"),
				},
			},
		},
	)

	return err
}
