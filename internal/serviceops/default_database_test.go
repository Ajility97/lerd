package serviceops

import "testing"

func TestConnectionDatabase(t *testing.T) {
	cases := map[string]string{
		"postgresql://postgres:lerd@127.0.0.1:{{host_port}}/lerd": "lerd",
		"mysql://root:lerd@127.0.0.1:3306/lerd":                   "lerd",
		"postgresql://postgres:lerd@127.0.0.1:5432/app?sslmode=x": "app",
		"mongodb://root:lerd@127.0.0.1:27017/?authSource=admin":   "",
		"http://lerd:lerd@127.0.0.1:8123":                         "",
		"redis://127.0.0.1:6379":                                  "",
		"":                                                        "",
	}
	for in, want := range cases {
		if got := connectionDatabase(in); got != want {
			t.Errorf("connectionDatabase(%q) = %q, want %q", in, got, want)
		}
	}
}

// A data directory initialised before the image was told to create the
// database never gets it, so a dependent pointed there would crash-loop.
func TestEnsureDefaultDatabase_CreatesWhatTheConnectionURLNames(t *testing.T) {
	origURL, origCreate := connectionURLFn, createDatabaseFn
	t.Cleanup(func() { connectionURLFn, createDatabaseFn = origURL, origCreate })

	connectionURLFn = func(string) string { return "postgresql://postgres:lerd@127.0.0.1:5432/lerd" }
	var got []string
	createDatabaseFn = func(svc, name string) (bool, error) {
		got = append(got, svc+"/"+name)
		return true, nil
	}
	if err := ensureDefaultDatabase("postgres"); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != "postgres/lerd" {
		t.Fatalf("created %v, want [postgres/lerd]", got)
	}

	got = nil
	connectionURLFn = func(string) string { return "redis://127.0.0.1:6379" }
	if err := ensureDefaultDatabase("redis"); err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("a URL naming no database created %v", got)
	}
}
