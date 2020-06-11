package gommunicator

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/aws/aws-sdk-go/service/sqs"
)

func isChannel(data interface{}) (is bool, channelType reflect.Type) {
	channelType = reflect.TypeOf(data)
	is = channelType.Kind() == reflect.Chan
	return
}

func typeOfChannelEntity(channel interface{}) (reflect.Type, error) {
	is, channelType := isChannel(channel)

	if !is {
		return nil, errors.New("receiver is not a channel")
	}

	elemType := channelType.Elem()

	return elemType, nil
}

func unmarshalToEntityType(incomingData string, entityType reflect.Type) (*reflect.Value, error) {
	data := reflect.New(entityType).Interface()

	err := json.Unmarshal([]byte(incomingData), &data)
	if err != nil {
		return nil, fmt.Errorf("not convertible to type %s. json: %s, err: %ss", entityType.Name(), incomingData, err.Error())
	}

	valueOf := reflect.ValueOf(data)
	return &valueOf, nil
}

func sendDataToChannel(data reflect.Value, receiver interface{}) error {
	if is, _ := isChannel(receiver); !is {
		return errors.New("receiver is not a channel")
	}

	channel := reflect.ValueOf(receiver)
	channel.Send(data)
	return nil
}

func (gom *Gommunicator) handleRawMessage(message *sqs.Message, receiver interface{}) error {
	elemType, err := typeOfChannelEntity(receiver)
	if err != nil {
		return err
	}

	data, err := unmarshalToEntityType(*message.Body, elemType)
	if err != nil {
		return err
	}

	sendDataToChannel(data.Elem(), receiver)

	return nil
}

func (gom *Gommunicator) handleMessage(message *sqs.Message, receiver chan<- *DataTransactionRequest) error {
	var request DataTransactionRequest

	err := json.Unmarshal([]byte(*message.Body), &request)
	if err != nil {
		return err
	}

	dt, err := gom.checkDT(request.DedupID)
	if err != nil {
		gom.createDT(request.DedupID)

		receiver <- &request

		gom.updateDT(request.DedupID, completed)
	}

	if dt.Status == inProgress || dt.Status == completed {
		return nil
	}

	if dt.Status == errored || dt.Status == nothing {
		gom.updateDT(request.DedupID, inProgress)

		receiver <- &request

		gom.updateDT(request.DedupID, completed)
	}

	return nil
}
