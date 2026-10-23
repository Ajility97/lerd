package cli

import (
	"strings"
	"testing"

	"github.com/geodro/lerd/internal/serviceops"
)

type s3ImportCalls struct {
	started, created []string
	mirrored         string
}

// stubS3ImportSeams stands in for the remote server and lerd's own S3 service.
func stubS3ImportSeams(t *testing.T, buckets []string, objects int) *s3ImportCalls {
	t.Helper()
	calls := &s3ImportCalls{}
	prevBuckets, prevCount, prevMirror := s3RemoteBuckets, s3RemoteObjectCount, s3MirrorInto
	prevRun, prevBucket := ensureServiceRunning, createS3Bucket
	s3RemoteBuckets = func(serviceops.S3Remote) ([]string, error) { return buckets, nil }
	s3RemoteObjectCount = func(serviceops.S3Remote, string) (int, error) { return objects, nil }
	s3MirrorInto = func(_ serviceops.S3Remote, from, to string, progress func(int)) error {
		calls.mirrored = from + "->" + to
		if progress != nil {
			progress(objects)
		}
		return nil
	}
	ensureServiceRunning = func(svc string) error {
		calls.started = append(calls.started, svc)
		return nil
	}
	createS3Bucket = func(name string) (bool, error) {
		calls.created = append(calls.created, name)
		return true, nil
	}
	t.Cleanup(func() {
		s3RemoteBuckets, s3RemoteObjectCount, s3MirrorInto = prevBuckets, prevCount, prevMirror
		ensureServiceRunning, createS3Bucket = prevRun, prevBucket
	})
	return calls
}

func TestS3Import_MirrorsIntoASanitisedLerdBucket(t *testing.T) {
	calls := stubS3ImportSeams(t, []string{"local", "other"}, 3)
	if err := runS3Import(serviceops.S3Remote{Endpoint: "127.0.0.1:29000"}, "local", "My_Shop"); err != nil {
		t.Fatal(err)
	}
	if strings.Join(calls.started, ",") != "rustfs" || strings.Join(calls.created, ",") != "my-shop" || calls.mirrored != "local->my-shop" {
		t.Errorf("calls = %+v", calls)
	}
}

func TestS3Import_TakesTheOnlyBucketWhenNoneIsNamed(t *testing.T) {
	calls := stubS3ImportSeams(t, []string{"uploads"}, 1)
	if err := runS3Import(serviceops.S3Remote{}, "", "shop"); err != nil {
		t.Fatal(err)
	}
	if calls.mirrored != "uploads->shop" {
		t.Errorf("mirrored %q", calls.mirrored)
	}
}

func TestS3Import_EmptyBucketTouchesNothing(t *testing.T) {
	calls := stubS3ImportSeams(t, []string{"local"}, 0)
	if err := runS3Import(serviceops.S3Remote{}, "local", "shop"); err != nil {
		t.Fatal(err)
	}
	if len(calls.created) != 0 || calls.mirrored != "" {
		t.Errorf("an empty bucket must not create or copy anything: %+v", calls)
	}
}

func TestS3Import_Refusals(t *testing.T) {
	for name, c := range map[string]struct {
		buckets  []string
		from, to string
		want     string
	}{
		"no destination":      {[]string{"local"}, "local", "", "--to"},
		"unknown bucket":      {[]string{"a", "b"}, "c", "shop", "a, b"},
		"several, none named": {[]string{"a", "b"}, "", "shop", "--bucket"},
		"server has none":     {nil, "", "shop", "no buckets"},
	} {
		t.Run(name, func(t *testing.T) {
			calls := stubS3ImportSeams(t, c.buckets, 1)
			err := runS3Import(serviceops.S3Remote{Endpoint: "x"}, c.from, c.to)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("err = %v, want one mentioning %q", err, c.want)
			}
			if calls.mirrored != "" {
				t.Errorf("nothing should be copied on a refusal")
			}
		})
	}
}
