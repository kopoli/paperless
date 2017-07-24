package jsend_test

import (
	"net/http"

	"github.com/gamegos/jsend"
)

func handler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"id":   1,
		"name": "foo",
	}

	jsend.Wrap(w).
		Data(data).
		Status(201).
		Send()
}

func Example() {
	http.ListenAndServe(":8080", http.HandlerFunc(handler))

	/*
		HTTP/1.1 201 Created
		Content-Type: application/json

		{
		  "status": "success",
		  "data": {
		    "id": 1,
		    "name": "foo"
		  }
		}
	*/
}
