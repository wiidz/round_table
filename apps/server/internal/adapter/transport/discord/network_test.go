package discord

import (
	"errors"
	"testing"
)

func TestRetrySend_succeedsFirstTry(t *testing.T) {
	calls := 0
	err := retrySend(func() error {
		calls++
		return nil
	})
	if err != nil || calls != 1 {
		t.Fatalf("calls=%d err=%v", calls, err)
	}
}

func TestRetrySend_succeedsAfterFailures(t *testing.T) {
	calls := 0
	err := retrySend(func() error {
		calls++
		if calls < 3 {
			return errors.New("EOF")
		}
		return nil
	})
	if err != nil || calls != 3 {
		t.Fatalf("calls=%d err=%v", calls, err)
	}
}

func TestRetrySend_exhaustsRetries(t *testing.T) {
	calls := 0
	err := retrySend(func() error {
		calls++
		return errors.New("EOF")
	})
	if err == nil || calls != maxSendRetries {
		t.Fatalf("calls=%d err=%v", calls, err)
	}
}
