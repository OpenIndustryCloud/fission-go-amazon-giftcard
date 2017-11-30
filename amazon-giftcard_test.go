package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var (
	server *httptest.Server
	//Test Data TV
	userJson          = ` {"amount": 23}`
	userJsonMalformed = ` {}`
	// ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr = httptest.NewRecorder()
)

func TestHandler(t *testing.T) {
	req, err := http.NewRequest("POST", "", strings.NewReader(userJson))
	reqMalformed, err := http.NewRequest("POST", "", strings.NewReader(userJsonMalformed))
	if err != nil {
		t.Error(err) //Something is wrong while sending request
	}

	//TEST CASES
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"Test Data-1", args{rr, req}, 12},
		{"Test Data-2", args{rr, reqMalformed}, 0},
	}
	for _, tt := range tests {
		// call ServeHTTP method
		// directly and pass Request and ResponseRecorder.
		handler := http.HandlerFunc(Handler)
		handler.ServeHTTP(tt.args.w, tt.args.r)

		// Check the status code is what we expect.
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}
		//check content type
		if ctype := rr.Header().Get("Content-Type"); ctype != "application/json" {
			t.Errorf("content type header does not match: got %v want %v",
				ctype, "application/json")
		}
		// // check the output
		// res, err := ioutil.ReadAll(rr.Body)
		// if err != nil {
		// 	t.Error(err) //Something is wrong while read res
		// }
		// got := TicketResponse{}
		// err = json.Unmarshal(res, &got)

		// if err != nil && got.Audit.TicketID == tt.want {
		// 	t.Errorf("%q. compute weather risk() = %v, want %v", tt.name, got.Audit.TicketID, tt.want)
		// }
	}
}
