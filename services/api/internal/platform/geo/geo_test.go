package geo

import "testing"

func TestDistanceMeters_SamePoint(t *testing.T) {
	d := DistanceMeters(-6.2, 106.8, -6.2, 106.8)
	if d != 0 {
		t.Fatalf("expected 0 for identical points, got %f", d)
	}
}

func TestDistanceMeters_KnownDistance(t *testing.T) {
	// Monas (-6.1754, 106.8272) to Bundaran HI (-6.1953, 106.8229),
	// real-world distance is roughly 2.3km.
	d := DistanceMeters(-6.1754, 106.8272, -6.1953, 106.8229)
	if d < 2000 || d > 2600 {
		t.Fatalf("expected ~2.3km, got %.0fm", d)
	}
}

func TestDistanceMeters_Symmetric(t *testing.T) {
	a := DistanceMeters(-6.2, 106.8, -6.21, 106.83)
	b := DistanceMeters(-6.21, 106.83, -6.2, 106.8)
	if a != b {
		t.Fatalf("distance should be symmetric: %f != %f", a, b)
	}
}
