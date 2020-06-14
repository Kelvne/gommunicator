package gommunicator

import (
	"fmt"

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

	errorHandler func(error)
	mq           *sqs.SQS
	orchestrator *sns.SNS
	dynamo       *dynamodb.DynamoDB
	actions      map[string]ActionHandler
	log          bool
	logger       *Logger
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
		errorHandler: func(err error) {},
		dynamo:       dynamo,
		actions:      make(map[string]ActionHandler),
		log:          true,
		logger:       getLogger(),
	}
}

func (gom *Gommunicator) tryLogInfo(values ...interface{}) {
	if gom.log {
		gom.logger.log(true, values...)
	}
}

func (gom *Gommunicator) tryLogErr(values ...interface{}) {
	if gom.log {
		gom.logger.log(false, values...)
	}
}

func (gom *Gommunicator) onErr(err error) {
	gom.tryLogErr(err)
	gom.onErr(err)
}

// SetLogState sets the log state
func (gom *Gommunicator) SetLogState(state bool) *Gommunicator {
	gom.log = state
	return gom
}

// SetErrorHandler configure the main error handler
func (gom *Gommunicator) SetErrorHandler(errorHandle func(error)) *Gommunicator {
	gom.errorHandler = errorHandle
	return gom
}

// Start start listening to new messages sended to this service's queue URL
func (gom *Gommunicator) Start(maxMessage int64, longPollingTime int64) error {
	// TODO: check for possibility of start
	if false {
		return nil
	}

	gom.tryLogInfo("Gommunicator is running!")
	gom.tryLogInfo(fmt.Sprintf("%s service is waiting for messages...", gom.ServiceName))

	for {
		messageOutput, err := gom.mq.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl:            &gom.ServiceQueueURL,
			AttributeNames:      aws.StringSlice([]string{"All"}),
			WaitTimeSeconds:     aws.Int64(longPollingTime),
			MaxNumberOfMessages: aws.Int64(maxMessage),
		})

		if gom.errorHandler != nil && err != nil {
			gom.onErr(err)
			continue
		}

		if len(messageOutput.Messages) > 0 {
			for _, message := range messageOutput.Messages {
				go func(m *sqs.Message) {
					handleError := gom.handleMessage(m)

					if handleError != nil {
						gom.onErr(handleError)
					}

					return
				}(message)
			}
		}
	}
}
