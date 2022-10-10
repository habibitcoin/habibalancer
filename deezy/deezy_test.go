package deezy

import "testing"

func TestIsChannelOpen(t *testing.T) {
	got := IsChannelOpen()
	want := true

	if got != want {
		t.Errorf("got %v, wanted %v", got, want)
	}
}

func TestIsNoChannelOpen(t *testing.T) {
	// need to add mock for ListChannels call to return []
	got := IsChannelOpen()
	want := false

	if got != want {
		t.Errorf("got %v, wanted %v", got, want)
	}
}
