package update

import "testing"

func TestUpdateRejectsUnversionedBuildBeforeNetworkAccess(t *testing.T) {
	t.Parallel()
	if _, err := Update(t.Context(), "dev"); err == nil {
		t.Fatal("expected unversioned build error")
	}
}

func TestIsNewer(t *testing.T) {
	t.Parallel()
	tests := []struct {
		latest  string
		current string
		want    bool
	}{
		{latest: "v0.2.0", current: "0.1.0", want: true},
		{latest: "v0.2.0", current: "0.2.0", want: false},
		{latest: "v0.1.9", current: "0.2.0", want: false},
		{latest: "v1.0.0", current: "0.9.9", want: true},
		{latest: "v1.0.0", current: "1.0.0-rc.1", want: true},
		{latest: "v1.0.0-rc.2", current: "1.0.0-rc.1", want: true},
	}
	for _, test := range tests {
		got, err := isNewer(test.latest, test.current)
		if err != nil {
			t.Fatalf("isNewer(%q, %q): %v", test.latest, test.current, err)
		}
		if got != test.want {
			t.Errorf("isNewer(%q, %q) = %v, want %v", test.latest, test.current, got, test.want)
		}
	}
}
