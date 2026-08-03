package version

import "testing"

func TestCurrentPrefersInjectedReleaseVersion(t *testing.T) {
	originalVersion := Version
	Version = "0.2.0"
	t.Cleanup(func() { Version = originalVersion })

	if got, want := Current(), "0.2.0"; got != want {
		t.Fatalf("Current() = %q, want %q", got, want)
	}
}
