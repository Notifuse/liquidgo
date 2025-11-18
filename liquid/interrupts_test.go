package liquid

import "testing"

func TestInterrupt(t *testing.T) {
	interrupt := NewInterrupt("test message")
	if interrupt.Message != "test message" {
		t.Errorf("Expected message 'test message', got '%s'", interrupt.Message)
	}

	emptyInterrupt := NewInterrupt("")
	if emptyInterrupt.Message != "interrupt" {
		t.Errorf("Expected default message 'interrupt', got '%s'", emptyInterrupt.Message)
	}
}

func TestBreakInterrupt(t *testing.T) {
	brk := NewBreakInterrupt()
	if brk.Message != "break" {
		t.Errorf("Expected message 'break', got '%s'", brk.Message)
	}
}

func TestContinueInterrupt(t *testing.T) {
	cont := NewContinueInterrupt()
	if cont.Message != "continue" {
		t.Errorf("Expected message 'continue', got '%s'", cont.Message)
	}
}
