package cleanup

import (
	"errors"
	"testing"
)

// withHistory stubs the per-image layer sizes, newest layer first the way podman
// history lists them. An image missing from the map fails like an unreadable one.
func withHistory(t *testing.T, sizes map[string][]int64) {
	t.Helper()
	historySizes = func(id string) ([]int64, error) {
		if s, ok := sizes[id]; ok {
			return s, nil
		}
		return nil, errors.New("no history for " + id)
	}
}

// withProtected protects refs and keeps the deep tier off the real config and catalog.
func withProtected(t *testing.T, refs ...string) {
	t.Helper()
	prot := map[string]bool{}
	for _, r := range refs {
		prot[r] = true
	}
	serviceRepos = func() (map[string]bool, error) { return map[string]bool{}, nil }
	protectedImages = func() (map[string]bool, error) { return prot, nil }
	t.Cleanup(func() {
		serviceRepos = realServiceRepos
		protectedImages = realProtectedImages
	})
}

func targetBytes(p Plan) map[string]int64 {
	got := map[string]int64{}
	for _, tg := range p.Targets {
		got[tg.ID] = tg.Bytes
	}
	return got
}

// A removed PHP version leaves its FPM image and base. podman credits the image
// only its top layer, so the plan counts layers itself and lists the base too.
func TestInspect_DeepCreditsLayersOfAnUnusedFPMImageAndItsBase(t *testing.T) {
	withImages(t, []image{
		{ID: "fpm86", Names: []string{"localhost/lerd-php86-fpm:local"}, Size: 1200, SharedSize: 1199, Labels: map[string]string{"dev.lerd.fpm.containerfile-hash": "h"}},
		{ID: "base86", Names: []string{"ghcr.io/lerd-env/lerd-php86-fpm-base:x"}, Size: 550, SharedSize: 549},
		{ID: "fpm85", Names: []string{"localhost/lerd-php85-fpm:local"}, Size: 1200, Containers: 1, Labels: map[string]string{"dev.lerd.fpm.containerfile-hash": "h"}},
	}, map[string][]string{
		"fpm86":  {"alpine", "php86", "chromium", "top86"},
		"base86": {"alpine", "php86"},
		"fpm85":  {"alpine", "php85", "chromium85", "top85"},
	})
	withProtected(t, "localhost/lerd-php85-fpm:local")
	withHistory(t, map[string][]int64{
		"fpm86":  {1, 0, 650, 540, 9},
		"base86": {0, 540, 9},
	})

	p, err := Inspect(ScopeDeep)
	if err != nil {
		t.Fatal(err)
	}
	got := targetBytes(p)
	// The shared alpine layer stays for php85; the base is credited its own
	// layers and the FPM image what it adds on top.
	if got["ghcr.io/lerd-env/lerd-php86-fpm-base:x"] != 540 || got["localhost/lerd-php86-fpm:local"] != 651 {
		t.Fatalf("want base 540 and fpm 651, got %+v", p.Targets)
	}
	if p.ReclaimBytes() != 1191 {
		t.Errorf("ReclaimBytes = %d, want 1191", p.ReclaimBytes())
	}
}

// The base of an FPM image that stays is still what that image is built on.
func TestInspect_DeepKeepsTheBaseOfAProtectedFPMImage(t *testing.T) {
	withImages(t, []image{
		{ID: "fpm86", Names: []string{"localhost/lerd-php86-fpm:local"}, Size: 1200, Labels: map[string]string{"dev.lerd.fpm.containerfile-hash": "h"}},
		{ID: "base86", Names: []string{"ghcr.io/lerd-env/lerd-php86-fpm-base:x"}, Size: 550},
	}, map[string][]string{
		"fpm86":  {"alpine", "php86", "chromium"},
		"base86": {"alpine", "php86"},
	})
	withProtected(t, "localhost/lerd-php86-fpm:local")

	p, err := Inspect(ScopeDeep)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Targets) != 0 {
		t.Fatalf("a protected FPM image keeps its base, got %+v", p.Targets)
	}
}

// Below the deep tier an FPM image is never reaped, so its base stays too.
func TestInspect_SafeKeepsTheBaseOfAnUnusedFPMImage(t *testing.T) {
	withImages(t, []image{
		{ID: "fpm86", Names: []string{"localhost/lerd-php86-fpm:local"}, Size: 1200, Labels: map[string]string{"dev.lerd.fpm.containerfile-hash": "h"}},
		{ID: "base86", Names: []string{"ghcr.io/lerd-env/lerd-php86-fpm-base:x"}, Size: 550},
	}, map[string][]string{
		"fpm86":  {"alpine", "php86", "chromium"},
		"base86": {"alpine", "php86"},
	})
	withProtected(t)

	p, err := Inspect(ScopeManaged)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Targets) != 0 {
		t.Fatalf("the managed tier never reaps an FPM image or its base, got %+v", p.Targets)
	}
}

// An image whose layer sizes can't be read keeps podman's estimate rather than
// reading as free.
func TestInspect_KeepsTheEstimateWhenLayerSizesAreUnreadable(t *testing.T) {
	withImages(t, []image{
		{ID: "py313", Names: []string{"docker.io/library/python:3.13-alpine"}, Size: 45},
	}, map[string][]string{"py313": {"Q1", "Q2"}})
	withProtected(t)

	p, err := Inspect(ScopeDeep)
	if err != nil {
		t.Fatal(err)
	}
	if got := targetBytes(p)["docker.io/library/python:3.13-alpine"]; got != 45 {
		t.Fatalf("want the 45 byte estimate kept, got %d", got)
	}
}

func TestLayerSizesFrom(t *testing.T) {
	sizes, ok := layerSizesFrom([]string{"a", "b"}, []int64{0, 7, 0, 3})
	if !ok || sizes["a"] != 3 || sizes["b"] != 7 {
		t.Fatalf("want a=3 b=7, got %v %v", sizes, ok)
	}
	// A zero-sized layer can't be told from an empty history step, so the
	// count no longer lines up and the sizes are refused.
	if _, ok := layerSizesFrom([]string{"a", "b", "c"}, []int64{7, 0, 3}); ok {
		t.Fatal("a history that doesn't line up with the layers must be refused")
	}
}
