package gommunicator

import (
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
)

// Gommunicator is the main wrapper for connecting to the services group
type Gommunicator struct {
	// The name of the service
	ServiceName string
	// The URL which represents the SQS queue URL related to this service
	ServiceQueueURL string
	// The SNS Topic that will receive incoming messages
	SNSTopicARN string
	// The table that hold transactions statuses
	DynamoTable string

	errorHandler  func(error)
	mq            *sqs.SQS
	orchestrator  *sns.SNS
	closeWildcard chan bool
	dynamo        *dynamodb.DynamoDB
}

// NewGommunicator returns a new Gommunicator using the SQS as mq using the provided AWS IAM Account ID and secret
func NewGommunicator(sqs *sqs.SQS, sns *sns.SNS, dynamo *dynamodb.DynamoDB, serviceQueueURL, serviceName, dynamoTable, snsTopicArn string) *Gommunicator {
	return &Gommunicator{
		ServiceName:     serviceName,
		ServiceQueueURL: serviceQueueURL,
		SNSTopicARN:     snsTopicArn,
		DynamoTable:     dynamoTable,

		mq:           sqs,
		orchestrator: sns,
		errorHandler: func(err error) {
			if err != nil {
				getLogger().Error(err.Error())
			}
		},
		dynamo: dynamo,
	}
}

// SetErrorHandler configure the main error handler
func (gom *Gommunicator) SetErrorHandler(errorHandle func(error)) *Gommunicator {
	gom.errorHandler = errorHandle
	return gom
}

// Start start listening to new messages sended to this service's queue URL
func (gom *Gommunicator) Start(maxMessage int64, receiver chan<- *DataTransactionRequest) error {
	// TODO: check for possibility of start
	if false {
		return nil
	}

	for {
		select {
		case <-gom.closeWildcard:
			close(receiver)
			return nil
		default:
			var wg sync.WaitGroup

			messageOutput, err := gom.mq.ReceiveMessage(&sqs.ReceiveMessageInput{
				QueueUrl:            &gom.ServiceQueueURL,
				AttributeNames:      aws.StringSlice([]string{"All"}),
				WaitTimeSeconds:     aws.Int64(0),
				MaxNumberOfMessages: aws.Int64(maxMessage),
			})

			if gom.errorHandler != nil && err != nil {
				gom.errorHandler(err)
			}

			wg.Add(len(messageOutput.Messages))

			for _, message := range messageOutput.Messages {
				go func(m *sqs.Message, w *sync.WaitGroup) {
					handleError := gom.handleMessage(m, receiver)

					if handleError != nil {
						gom.errorHandler(handleError)
					}

					w.Done()
				}(message, &wg)
			}

			wg.Wait()
		}
	}
}

// StartChan start listening to new messages async and returns a receiver channel
// ps. not safe for verifying startage
func (gom *Gommunicator) StartChan(maxMessages int64) <-chan *DataTransactionRequest {
	receiver := make(chan *DataTransactionRequest, maxMessages)
	go gom.Start(maxMessages, receiver)
	return receiver
}

// CloseChan closes the goroutine and the channel created by StartChan
func (gom *Gommunicator) CloseChan() {
	close(gom.closeWildcard)
}
