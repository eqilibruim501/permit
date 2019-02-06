package permit

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin/json"

	"github.com/crusttech/permit/internal/context"
)

type (
	httpClientMock struct {
		do func(req *http.Request) (*http.Response, error)
	}
)

func (m httpClientMock) Do(req *http.Request) (*http.Response, error) {
	return m.do(req)
}

func makeHttpClientMock(code int, p *Permit) httpClient {
	return httpClientMock{
		do: func(req *http.Request) (*http.Response, error) {
			j, _ := json.Marshal(p)

			return &http.Response{
				StatusCode: code,
				// Send response to be tested
				Body: ioutil.NopCloser(bytes.NewBuffer(j)),
				// Must be set to non-nil value or it panics
				Header: make(http.Header),
			}, nil
		},
	}
}

func assert(t *testing.T, ok bool, format string, args ...interface{}) bool {
	if !ok {
		t.Fatalf(format, args...)
	}
	return ok
}

func TestCheckWithClient(t *testing.T) {
	var (
		key    = "teCYbMI8vSvi8hKF3Jb23jyeEmI7xbybWSYJXv8TDBQqIfBhGWYuPguBsfhNGaPU"
		domain = "example.tld"
		p      *Permit
		err    error
	)

	p, err = CheckWithClient(nil, nil, Permit{Key: "key"})
	assert(t, err != nil, "expecting error for invalid key length")
	assert(t, p == nil, "expecting permit to be nil")

	// test permit
	tp := Permit{Key: key, Domain: domain}

	p, err = CheckWithClient(
		context.Background(),
		makeHttpClientMock(http.StatusOK, &Permit{Key: key, Domain: domain, Valid: true}),
		tp,
	)

	assert(t, err == nil, "unexpected error: %v", err)
	assert(t, p != nil, "not expecting nil for permit")
	assert(t, p.Key == key, "permit key does not match")
	assert(t, p.Domain == domain, "permit domain does not match")

	p, err = CheckWithClient(context.Background(), makeHttpClientMock(http.StatusBadRequest, nil), tp)
	assert(t, err != nil, "expecting error on bad request")

	p, err = CheckWithClient(context.Background(), makeHttpClientMock(http.StatusNotFound, nil), tp)
	assert(t, err != nil, "expecting error when not found")

	p, err = CheckWithClient(context.Background(), makeHttpClientMock(http.StatusInternalServerError, nil), tp)
	assert(t, err != nil, "expecting error on bad request")

	p, err = CheckWithClient(context.Background(), makeHttpClientMock(http.StatusUnauthorized, nil), tp)
	assert(t, err != nil, "expecting error on bad request")

}
