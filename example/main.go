package main

import (
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/novalagung/natskeepalivesubscribe"
	"strings"
)

func main() {

	// get nats url from env var
	natsURL := "nats://localhost:4222"
	subject := "service-example"

	// keep alive subscribe
	natskeepalivesubscribe.KeepAliveSubscribe(natsURL, subject, func(msg *nats.Msg) (interface{}, error) {

		// parse payload
		payload := make(map[string]interface{})
		err := json.Unmarshal(msg.Data, &payload)
		if err != nil {
			return nil, err
		}

		// handle the request
		switch strings.ToUpper(payload["method"].(string)) {
		case "OPTIONS":
			// ...
		case "GET":
			// ...
		case "POST":
			// ...
		case "PATCH":
			// ...
		case "PUT":
			// ...
		case "DELETE":
			// ...
		}

		// error on invalid http method request
		return nil, fmt.Errorf("invalid http method")
	})
}
