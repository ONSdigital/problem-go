package problem_test

import (
	"encoding/json"
	"log"
	"net/http"
	"testing"

	problem "github.com/ONSdigital/problem-go"
)

func ExampleWriteResponse() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength == 0 {
			problem.WriteResponse(problem.Details{
				Type:   "https://example.com/help#bad-body",
				Title:  "Problem parsing request body",
				Status: http.StatusBadRequest,
				Detail: "No content was received for the request body. Please check your request and try again",
			}, w)
			return
		}

		// Otherwise other processing ...
	})
}

// Use a global for the content written as the ResponseWriter interface won't
// let us write directly to the underlying mock
var written []byte
var status int

type MockResponseWriter struct {
	Written []byte
	header  http.Header
}

func NewMockResponseWriter() MockResponseWriter {
	written = []byte{}
	status = -1
	return MockResponseWriter{
		header: http.Header{},
	}
}

func (m MockResponseWriter) Header() http.Header {
	return m.header
}

func (m MockResponseWriter) Write(b []byte) (int, error) {
	written = append(written, b...)
	return len(b), nil
}

func (m MockResponseWriter) WriteHeader(s int) {
	log.Println("Write status ", s)
	status = s
}

func TestWriteResponse(t *testing.T) {

	rw := NewMockResponseWriter()

	problem.WriteResponse(problem.Details{
		Type:   "http://localhost:9999/help",
		Title:  "Problem parsing request body",
		Status: http.StatusBadRequest,
		Detail: "The cat got stuck",
	}, rw)

	assertStringEquals("content-type header", rw.Header().Get("Content-type"), "application/problem+json", t)
	assertStringEquals("content-language header", rw.Header().Get("Content-Language"), "en", t)

	var np problem.Details
	if err := json.Unmarshal(written, &np); err != nil {
		t.Fatal("Failed to unmarshal written content")
	}

	assertStringEquals("title", np.Title, "Problem parsing request body", t)
	assertStringEquals("type", np.Type, "http://localhost:9999/help", t)
	assertStringEquals("detail", np.Detail, "The cat got stuck", t)

	assertIntEquals("status", status, http.StatusBadRequest, t)
	assertIntEquals("problem status", np.Status, http.StatusBadRequest, t)

}

func TestWriteResponseMissingStatus(t *testing.T) {
	rw := NewMockResponseWriter()
	problem.WriteResponse(problem.Details{}, rw)
	assertIntEquals("status", status, http.StatusInternalServerError, t)
}

func assertIntEquals(thing string, got, expected int, t *testing.T) {
	if got != expected {
		t.Errorf("Expected %s equal, got '%d', expected '%d'", thing, got, expected)
	}
}

func assertStringEquals(thing, got, expected string, t *testing.T) {
	if got != expected {
		t.Errorf("Expected %s equal, got '%s', expected '%s'", thing, got, expected)
	}
}
