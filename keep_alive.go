// Package natskeepalivesubscribe definition
package natskeepalivesubscribe

import (
	"encoding/json"
	"fmt"
	"log"
	"runtime/debug"
	"time"

	"github.com/nats-io/nats.go"
)

// Response handler base schema
type Response struct {
	Success      bool
	Data         interface{}
	ErrorMessage string
}

// recoverPossiblePanicError do recover possible panic, transform it into ordinary error
func recoverPossiblePanicError(err *error) {
	if e := recover(); e != nil {
		if err != nil {
			*err = fmt.Errorf("%v", e)
		}
		log.Println("PANIC ERROR (RECOVERED):", e)
		log.Println("STACK TRACE:", string(debug.Stack()))
	}
}

// delay for a while
func delay() {
	time.Sleep(time.Millisecond * 100)
}

// moveOnWithError to the next process
func moveOnWithError(err error) {
	log.Println(err.Error())
	log.Println("re-estabilishing nats subscription now ...")
	delay()
}

// jsonify convert any object into json string bytes
func jsonify(any interface{}) []byte {
	buf, _ := json.Marshal(any)
	return buf
}

// KeepAliveSubscribe perform nats subscription with auto re-estabilish connection in case of failing
func KeepAliveSubscribe(url, subject string, handler func(msg *nats.Msg) (interface{}, error)) {

	// infinity loop (with some delay) enable auto reconnection easier
	for {

		log.Printf("subscribing to nats server with subject '%s' \n", subject)

		// create nats connection object server connection
		conn, err := nats.Connect(url)
		if err != nil {
			moveOnWithError(fmt.Errorf("ERROR: unable to create nats connection object. %s", err.Error()))
			continue
		}

		// subscribe to nats server
		_, err = conn.Subscribe(subject, func(msg *nats.Msg) {

			// gracefully handle incoming request
			var handleErr, err error
			resp := func(innerErr *error) *Response {
				defer recoverPossiblePanicError(innerErr)
				data, err := handler(msg)
				if err != nil {
					return &Response{Success: false, ErrorMessage: err.Error()}
				}
				return &Response{Success: true, Data: data}
			}(&handleErr)

			// generate error response on panic
			if handleErr != nil {
				resp := jsonify(&Response{Success: false, ErrorMessage: handleErr.Error()})
				err := conn.Publish(msg.Reply, resp)
				if err != nil {
					log.Println("ERROR: unable publish response", err)
					return
				}
			}

			// generate success response
			err = conn.Publish(msg.Reply, jsonify(resp))
			if err != nil {
				log.Println("ERROR: unable publish response", err)
				return
			}
		})
		if err != nil {
			moveOnWithError(fmt.Errorf("ERROR: unable to subscribe to nats server with subject '%s'. %s", subject, err.Error()))
			continue
		}

		log.Println("connected")

		// watch nats connection state
		chSubIsClosed := make(chan bool)
		go func(chSubIsClosed chan bool, conn *nats.Conn) {
			// check nats conn every 0.1 second
			for conn.IsConnected() {
				delay()
			}
			// if it's not connected, then close the current conn
			conn.Close()
			// next, send signal to chSubIsClosed
			chSubIsClosed <- true
		}(chSubIsClosed, conn)

		// go to next loop for re-estabilishing subscription in case current conn is closed
		<-chSubIsClosed
		moveOnWithError(fmt.Errorf("ERROR: nats subscription is closed"))
	}
}
