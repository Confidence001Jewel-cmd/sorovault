package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStartRefresherLoopNoOpWhenDisabled(t *testing.T) {
	t.Parallel()

	called := false
	stop := startRefresherLoop(context.Background(), 0, func(context.Context) error {
		called = true
		return nil
	})
	stop()
	assert.False(t, called)
}

func TestStartRefresherLoopCancelsAndDrains(t *testing.T) {
	t.Parallel()

	started := make(chan struct{})
	cancelled := make(chan struct{})
	finish := make(chan struct{})
	stop := startRefresherLoop(context.Background(), time.Minute, func(ctx context.Context) error {
		close(started)
		<-ctx.Done()
		close(cancelled)
		<-finish
		return nil
	})
	<-started

	stopped := make(chan struct{})
	go func() {
		stop()
		close(stopped)
	}()

	select {
	case <-cancelled:
	case <-time.After(time.Second):
		t.Fatal("stop did not cancel the refresher")
	}
	select {
	case <-stopped:
		t.Fatal("stop returned before the in-flight refresher finished")
	default:
	}

	close(finish)
	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("stop did not return after the refresher drained")
	}
	assert.NotPanics(t, stop, "calling stop again is harmless")
}
