package keepalive_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ronelliott/keepalive"
)

func TestExponentialBackoff(t *testing.T) {
	backoff := keepalive.NewExponentialBackoff(time.Second, time.Minute)

	assert.Equal(t, time.Second, backoff(), "First backoff should be 1 second")
	assert.Equal(t, 2*time.Second, backoff(), "Second backoff should be 2 seconds")
	assert.Equal(t, 4*time.Second, backoff(), "Third backoff should be 4 seconds")
	assert.Equal(t, 8*time.Second, backoff(), "Fourth backoff should be 8 seconds")
	assert.Equal(t, 16*time.Second, backoff(), "Fifth backoff should be 16 seconds")
	assert.Equal(t, 32*time.Second, backoff(), "Sixth backoff should be 32 seconds")
	assert.Equal(t, time.Minute, backoff(), "Seventh backoff should be 60 seconds")
	assert.Equal(t, time.Minute, backoff(), "Eighth backoff should be 60 seconds")
	assert.Equal(t, time.Minute, backoff(), "Ninth backoff should be 60 seconds")
	assert.Equal(t, time.Minute, backoff(), "Tenth backoff should be 60 seconds")
}
