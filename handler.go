package gommunicator

import (
	"encoding/json"
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

func (gom *Gommunicator) handleMessage(message *sqs.Message) error {
	isRequest := message.Attributes["Action"] != nil

	var request *DataTransactionRequest
	var response *DataTransactionResponse

	var errDyn error
	var dedupID string

	if isRequest {
		err := json.Unmarshal([]byte(*message.Body), request)
		if err != nil {
			_, err = gom.mq.DeleteMessage(&sqs.DeleteMessageInput{
				QueueUrl:      aws.String(gom.ServiceQueueURL),
				ReceiptHandle: message.ReceiptHandle,
			})
			return err
		}
		dedupID = request.DedupID
	} else {
		err := json.Unmarshal([]byte(*message.Body), response)
		if err != nil {
			_, err = gom.mq.DeleteMessage(&sqs.DeleteMessageInput{
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
