package middleware

import (
	"testing"
	"time"
)

// A4/D-4: legacy login endpoints had no throttling at all.

func TestLoginRateLimiter_BlocksAfterLimit(t *testing.T) {
	rl := NewLoginRateLimiter(3, time.Minute)

	for i := 0; i < 3; i++ {
		if !rl.Allow("1.2.3.4") {
			t.Fatalf("attempt %d should be allowed", i+1)
		}
	}
	if rl.Allow("1.2.3.4") {
		t.Fatal("4th attempt within the window should be blocked")
	}
}

func TestLoginRateLimiter_KeyedPerClient(t *testing.T) {
	rl := NewLoginRateLimiter(1, time.Minute)

	if !rl.Allow("1.1.1.1") {
		t.Fatal("first attempt for client A should be allowed")
	}
	if !rl.Allow("2.2.2.2") {
		t.Fatal("a different client (B) must not be blocked by A's attempts")
	}
	if rl.Allow("1.1.1.1") {
		t.Fatal("second attempt for client A within the window should be blocked")
	}
}

func TestLoginRateLimiter_ResetsAfterWindow(t *testing.T) {
	rl := NewLoginRateLimiter(1, 50*time.Millisecond)

	if !rl.Allow("1.2.3.4") {
		t.Fatal("first attempt should be allowed")
	}
	if rl.Allow("1.2.3.4") {
		t.Fatal("second attempt within the window should be blocked")
	}
	time.Sleep(60 * time.Millisecond)
	if !rl.Allow("1.2.3.4") {
		t.Fatal("attempt after the window has passed should be allowed again")
	}
}
