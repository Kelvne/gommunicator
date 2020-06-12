package gommunicator

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func handlerErr(id, action, incoming, service string) string {
	return formatWithID(
		id,
		"data transaction id",
		fmt.Sprintf("Data transaction request errored on %s.%s from %s", action, service, incoming),
	)
}

func handlerRequestSuccess(id, action, incoming, service string) string {
	return formatWithID(
		id,
		"data transaction id",
		fmt.Sprintf("Data transaction request received on %s.%s from %s", action, service, incoming),
	)
}

func handlerSuccessResponse(id, action string) string {
	return formatWithID(
		id,
		"data transaction id",
		fmt.Sprintf("Data transaction success internal response read from action %s", action),
	)
}

func (gom *Gommunicator) handleDuplicated(dupID string) (*dtDocument, error) {
	// check for dynamodb request state
	dt, err := gom.checkDT(dupID)

	if err == nil && dt == nil {
		gom.createDT(dupID)
	}

	return dt, err
}

func (gom *Gommunicator) deleteMessage(message *sqs.Message) error {
	_, err := gom.mq.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      aws.String(gom.ServiceQueueURL),
		ReceiptHandle: message.ReceiptHandle,
	})
	return err
}

func (gom *Gommunicator) handleMessage(message *sqs.Message) error {
	var raw map[string]interface{}

	err := json.Unmarshal([]byte(*message.Body), &raw)
	if err != nil {
		gom.deleteMessage(message)
		return err
	}

	rawMessage, hasMsg := raw["Message"].(string)
	_, isRequest := raw["MessageAttributes"].(map[string]interface{})["Action"]

	if !hasMsg {
		gom.deleteMessage(message)
		return errors.New("empty message")
	}

	request := new(DataTransactionRequest)
	response := new(DataTransactionResponse)

	var errDyn error
	var dedupID string

	if isRequest {
		err := json.Unmarshal([]byte(rawMessage), request)
		if err != nil {
			gom.mq.DeleteMessage(&sqs.DeleteMessageInput{
				QueueUrl:      aws.String(gom.ServiceQueueURL),
				ReceiptHandle: message.ReceiptHandle,
			})
			return err
		}
		dedupID = request.DedupID
	} else {
		err := json.Unmarshal([]byte(rawMessage), response)
		if err != nil {
			gom.mq.DeleteMessage(&sqs.DeleteMessageInput{
				QueueUrl:      aws.String(gom.ServiceQueueURL),
				ReceiptHandle: message.ReceiptHandle,
			})
			return err
		}
		dedupID = response.DedupID
	}

	_, errDyn = gom.handleDuplicated(dedupID)

	if errDyn == nil {
		if isRequest {
			err := gom.CallAction(request)

			if err != nil {
				gom.updateDT(dedupID, errored)
				gom.tryLogErr(handlerErr(request.ID, request.Action, request.Service, request.IncomingService))
			} else {
				gom.updateDT(dedupID, completed)
				gom.tryLogInfo(handlerRequestSuccess(request.ID, request.Action, request.Service, request.IncomingService))
			}
		} else {
			err := callCallback(response)
			if err == nil {
				gom.tryLogInfo(handlerSuccessResponse(response.ID, response.Action))
			}

			_, errD := gom.mq.DeleteMessage(&sqs.DeleteMessageInput{
				QueueUrl:      aws.String(gom.ServiceQueueURL),
				ReceiptHandle: message.ReceiptHandle,
			})
			if err != nil || errD != nil {
				gom.updateDT(request.DedupID, errored)
			} else {
				gom.updateDT(request.DedupID, completed)
			}
		}
	} else {
		gom.updateDT(dedupID, errored)
		gom.tryLogErr(errDyn)
	}

	gom.mq.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      aws.String(gom.ServiceQueueURL),
		ReceiptHandle: message.ReceiptHandle,
	})

	return nil
}
