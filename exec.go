package gommunicator

import (
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/google/uuid"
)

// Exec executes an action on the services cluster
// It receives a DataTransactionRequest as parameter and the timeout policy in seconds
func (gom *Gommunicator) Exec(data *DataTransactionRequest, timeout int) (chan<- *DataTransactionResponse, error) {
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
					BinaryValue: []byte(now),
					DataType:    aws.String("Number"),
				},
				"Service": {
					StringValue: aws.String(data.Service),
					BinaryValue: []byte(data.Service),
					DataType:    aws.String("String"),
				},
				"Action": {
					StringValue: aws.String(data.Action),
					BinaryValue: []byte(data.Action),
					DataType:    aws.String("String"),
				},
			},
		},
	)

	if err != nil {
		return nil, err
	}

	go gom.handleResponse(receiver, data, time.Duration(timeout)*time.Second)

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
					BinaryValue: []byte(now),
					DataType:    aws.String("Number"),
				},
				"Service": {
					StringValue: aws.String(request.IncomingService),
					BinaryValue: []byte(request.IncomingService),
					DataType:    aws.String("String"),
				},
			},
		},
	)

	return err
}
