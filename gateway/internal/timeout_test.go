package internal

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetTimeout(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set(grpcTimeoutHeader, "1s")
	timeout := GetTimeout(req.Header, time.Second*5)
	assert.Equal(t, time.Second, timeout)
}

func TestGetTimeoutDefault(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	timeout := GetTimeout(req.Header, time.Second*5)
	assert.Equal(t, time.Second*5, timeout)
}
