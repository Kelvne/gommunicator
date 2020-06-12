package deco

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/kelvne/gommunicator"
)

func decode(data interface{}, incoming interface{}) error {
	baseData := reflect.ValueOf(data)
	incomingType := reflect.TypeOf(incoming)
	structType := incomingType.Elem()

	if incomingType.Kind() != reflect.Ptr {
		return errors.New("invalid incoming object")
	}

	if structType.Kind() == reflect.Slice {
		structType = structType.Elem().Elem()
	}

	if baseData.Kind() == reflect.Ptr {
		baseData = reflect.ValueOf(baseData.Elem())
	}

	if !baseData.IsValid() {
		return nil
	}

	switch baseData.Kind() {
	case reflect.Array, reflect.Slice:
		if incomingType.Elem().Kind() != reflect.Slice && incomingType.Elem().Kind() != reflect.Array {
			return errors.New("trying to parse a slice to a single object")
		}

		incomingValue := reflect.ValueOf(incoming).Elem()

		for i := 0; i < baseData.Len(); i++ {
			marshaled, err := json.Marshal(baseData.Index(i).Interface())
			if err != nil {
				return err
			}
			toParse := reflect.New(structType).Interface()
			err = json.Unmarshal(marshaled, toParse)
			if err != nil {
				return fmt.Errorf("error while unmarshaling value to slice item: %s", err.Error())
			}

			incomingValue.Set(reflect.Append(incomingValue, reflect.ValueOf(toParse)))
			return nil
		}

	case reflect.Struct, reflect.Map:
		incomingValue := reflect.ValueOf(incoming)

		marshaled, err := json.Marshal(data)
		if err != nil {
			return err
		}

		if incomingValue.IsZero() {
			val := reflect.New(structType.Elem())
			incoming = val.Interface()
		}

		err = json.Unmarshal(marshaled, incomingValue.Interface())
		return err
	default:
		return errors.New("cant parse directly this data type, use the receiver channel and decode by yourself")
	}

	return nil
}

// DecodeRequest decodes the data of a request to a incoming struct or slice of
func DecodeRequest(dt *gommunicator.DataTransactionRequest, incoming interface{}) error {
	return decode(dt.Data, incoming)
}

// DecodeResponse decodes the data of a response to a incoming struct or slice of
func DecodeResponse(dt *gommunicator.DataTransactionResponnse, incoming interface{}) error {
	return decode(dt.Data, incoming)
}
