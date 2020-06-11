package gommunicator

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
)

func topicArn(service string, outputs []*sns.ListTopicsOutput) *string {
	for _, output := range outputs {
		for _, topic := range output.Topics {
			arn := *topic.TopicArn

			match, err := regexp.Match(fmt.Sprintf("CLUSTER_%s", service), []byte(arn))
			if err != nil {
				continue
			}
			if match {
				return topic.TopicArn
			}
		}
	}

	return nil
}

// Exec executes an action on the services cluster
// It receives a DataTransactionRequest as parameter and the timeout policy in seconds
func (gom *Gommunicator) Exec(data *DataTransactionRequest, timeout int) (chan<- *DataTransactionResponse, error) {
	receiver := make(chan *DataTransactionResponse)

	outputs := make([]*sns.ListTopicsOutput, 0)
	err := gom.orchestrator.ListTopicsPages(
		&sns.ListTopicsInput{},
		func(output *sns.ListTopicsOutput, lastPage bool) bool {
			outputs = append(outputs, output)
			return !lastPage
		},
	)

	if err != nil {
		return receiver, nil
	}

	arn := topicArn(data.Service, outputs)

	bytesMessage, err := json.Marshal(&data)
	if err != nil {
		return receiver, nil
	}

	message := string(bytesMessage)

	now := string(time.Now().UnixNano())

	gom.orchestrator.Publish(
		&sns.PublishInput{
			TopicArn: arn,
			Message:  aws.String(message),
			MessageAttributes: map[string]*sns.MessageAttributeValue{
				"Timestamp": {
					StringValue: aws.String(now),
					BinaryValue: []byte(now),
					DataType:    aws.String("Number"),
				},
			},
		},
	)

	go gom.handleResponse(receiver, data, time.Duration(timeout)*time.Second)

	return receiver, nil
}
