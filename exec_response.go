package gommunicator

import (
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func setTimeout(duration time.Duration, callback func()) {
	go func(d time.Duration, cb func()) {
		time.Sleep(duration)
		cb()
	}(duration, callback)
}

func (gom *Gommunicator) handleResponse(
	receiver chan<- *DataTransactionResponse,
	request *DataTransactionRequest,
	timeout time.Duration,
) {
	end := make(chan bool)

	setTimeout(timeout, func() {
		close(end)
	})

	for {
		select {
		case <-end:
			return
		default:
			messageOutput, err := gom.mq.ReceiveMessage(&sqs.ReceiveMessageInput{
				QueueUrl:            &gom.ServiceQueueURL,
				AttributeNames:      aws.StringSlice([]string{"All"}),
				WaitTimeSeconds:     aws.Int64(0),
				MaxNumberOfMessages: aws.Int64(1),
			})

			lenOfMessages := len(messageOutput.Messages)

			if gom.errorHandler != nil && err != nil {
				gom.errorHandler(err)
			}

			if lenOfMessages > 0 {
				for _, message := range messageOutput.Messages {
					var response DataTransactionResponse

					dt, err := gom.checkDT(request.DedupID)
					if err != nil {
						gom.createDT(request.DedupID)
						err := json.Unmarshal([]byte(*message.Body), &response)
						if err != nil {
							close(receiver)
						}

						receiver <- &response
						close(receiver)
						gom.updateDT(request.DedupID, completed)
					}

					if dt.Status == inProgress || dt.Status == completed {
						return
					}

					if dt.Status == errored || dt.Status == nothing {
						gom.updateDT(request.DedupID, inProgress)
						err := json.Unmarshal([]byte(*message.Body), &response)
						if err != nil {
							close(receiver)
						}

						receiver <- &response
						close(receiver)
						gom.updateDT(request.DedupID, completed)
					}
				}

				close(end)
			}
		}
	}
}
