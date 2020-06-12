package gommunicator

import (
	"reflect"
	"testing"
)

func fDecode(t *testing.T, err error) {
	t.Fatalf("DecodeRequest failed: %s", err.Error())
}

func fDecodeValue(t *testing.T, expected interface{}) {
	t.Fatalf("DecodeRequest failed, wrong value, %v", expected)
}

type Expected struct {
	Message string `json:"msg"`
}

func TestDecodeRequest(t *testing.T) {
	rq := DataTransactionRequest{
		Data: map[string]interface{}{
			"msg": "hello",
		},
	}

	expected := new(Expected)

	err := DecodeRequest(&rq, expected)
	if err != nil {
		fDecode(t, err)
	}

	if expected.Message != rq.Data.(map[string]interface{})["msg"] {
		fDecodeValue(t, rq.Data.(map[string]interface{})["msg"])
	}

	rq = DataTransactionRequest{}
	var x *string

	err = DecodeRequest(&rq, x)
	if err != nil {
		fDecode(t, err)
	}

	if x != nil {
		fDecodeValue(t, x)
	}

	rq.Data = "oi"

	err = DecodeRequest(&rq, x)
	if err == nil {
		fDecode(t, nil)
	}
}

func TestDecodeSlice(t *testing.T) {
	rq := DataTransactionRequest{
		Data: []map[string]interface{}{
			map[string]interface{}{
				"msg": "hello from slice",
			},
		},
	}

	expectedSl := make([]*Expected, 0)
	err := DecodeRequest(&rq, &expectedSl)
	if err != nil {
		fDecode(t, err)
	}

	returnE := []*Expected{
		&Expected{
			Message: rq.Data.([]map[string]interface{})[0]["msg"].(string),
		},
	}

	if !reflect.DeepEqual(expectedSl, returnE) {
		fDecodeValue(t, rq.Data)
	}
}
