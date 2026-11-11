package testutil

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// fakeTB records the first fatal message and stops its goroutine, the way
// testing.T does, so Golden's failure paths can be observed.
type fakeTB struct {
	testing.TB
	msg string
}

func (f *fakeTB) Helper() {}

func (f *fakeTB) Fatal(args ...any) {
	f.msg = fmt.Sprint(args...)
	runtime.Goexit()
}

func (f *fakeTB) Fatalf(format string, args ...any) {
	f.msg = fmt.Sprintf(format, args...)
	runtime.Goexit()
}

// golden runs Golden against a fakeTB and returns its fatal message ("" if it passed).
func golden(name, got string) string {
	f := &fakeTB{}
	done := make(chan struct{})
	go func() {
		defer close(done)
		Golden(f, name, "thing", got)
	}()
	<-done
	return f.msg
}

func TestGolden(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.Mkdir("testdata", 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join("testdata", "x.golden"), []byte("a\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if msg := golden("x.golden", "a\n"); msg != "" {
		t.Errorf("match failed: %q", msg)
	}
	if msg, want := golden("x.golden", "b\n"), "thing drifted from testdata/x.golden:\nb\n"; msg != want {
		t.Errorf("mismatch message = %q, want %q", msg, want)
	}
	if msg := golden("missing.golden", "a\n"); msg == "" {
		t.Error("a missing golden passed")
	}
}

func TestGoldenUpdate(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("UPDATE_GOLDEN", "1")
	if msg := golden("x.golden", "new\n"); msg == "" {
		t.Error("update into a missing testdata dir passed")
	}
	if err := os.Mkdir("testdata", 0o700); err != nil {
		t.Fatal(err)
	}
	if msg := golden("x.golden", "new\n"); msg != "" {
		t.Errorf("update failed: %q", msg)
	}
	b, err := os.ReadFile(filepath.Join("testdata", "x.golden"))
	if err != nil || string(b) != "new\n" {
		t.Errorf("golden after update = %q, %v", b, err)
	}
}
