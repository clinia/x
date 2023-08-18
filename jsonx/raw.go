package jsonx

import "encoding/json"

// Raw returns a json.RawMessage.
// This is useful for testing json payload.
// The function will panic if the input is not valid json.
// It normalize the json by unmarshaling and marshaling it again.
func RawMessage(bytesin []byte) json.RawMessage {
	var v interface{}
	err := json.Unmarshal(bytesin, &v)
	if err != nil {
		panic(err)
	}

	bytesout, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return bytesout
}
