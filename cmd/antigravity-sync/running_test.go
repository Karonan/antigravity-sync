package main

import "testing"

// Smoke test: must not panic or hang whether or not Antigravity is open.
func TestAntigravityRunning(t *testing.T) {
	t.Logf("antigravityRunning() = %v", antigravityRunning())
}
