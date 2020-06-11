package gommunicator

import (
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type dtStatus string

func statusToStr(input string) dtStatus {
	switch input {
	case "IN_PROGRESS":
		return inProgress
	case "ERRORED":
		return errored
	case "COMPLETED":
		return completed
	}

	return nothing
}

const (
	inProgress dtStatus = "IN_PROGRESS"
	completed           = "COMPLETED"
	errored             = "ERRORED"
	nothing             = "NOTHING"
)

type dtDocument struct {
	ID        string   `json:"id"`
	Status    dtStatus `json:"status"`
	Timestamp int64    `json:"timestamp"`
}

func (gom *Gommunicator) checkDT(dtID string) (*dtDocument, error) {
	itemOutput, err := gom.dynamo.GetItem(&dynamodb.GetItemInput{
		ConsistentRead: aws.Bool(true),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(dtID),
			},
		},
		TableName: aws.String(gom.DynamoTable),
	})

	if err != nil {
		return nil, err
	}

	ts, err := strconv.ParseInt(*itemOutput.Item["timestamp"].N, 10, 64)
	if err != nil {
		ts = int64(0)
	}

	return &dtDocument{
		ID:        *itemOutput.Item["id"].S,
		Status:    statusToStr(*itemOutput.Item["status"].S),
		Timestamp: ts,
	}, nil
}

func (gom *Gommunicator) createDT(dtID string) error {
	_, err := gom.dynamo.PutItem(&dynamodb.PutItemInput{
		ConditionExpression: aws.String("attribute_not_exists(id)"),
		Item: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(dtID),
			},
			"status": {
				S: aws.String(string(inProgress)),
			},
			"timestamp": {
				N: aws.String(strconv.FormatInt(time.Now().UnixNano(), 10)),
			},
		},
		TableName: aws.String(gom.DynamoTable),
	})

	return err
}

func (gom *Gommunicator) updateDT(dtID string, status dtStatus) error {
	_, err := gom.dynamo.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String(gom.DynamoTable),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(dtID),
			},
		},
		UpdateExpression: aws.String("set status = :st"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":st": {
				S: aws.String(string(status)),
			},
		},
	})

	return err
}
