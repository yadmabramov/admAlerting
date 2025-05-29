package main

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {

	_, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	go func() {
		main()
	}()

	time.Sleep(50 * time.Millisecond)

	resp, err := http.Get("http://localhost:8080/update/gauge/test/1")
	if resp != nil {
		defer resp.Body.Close() // Закрываем тело ответа
	}
	assert.NoError(t, err)
}
