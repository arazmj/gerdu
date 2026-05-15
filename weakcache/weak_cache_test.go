package weakcache

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"runtime"
	"strconv"
	"testing"
)

func TestWeakCache(t *testing.T) {
	cache := NewWeakCache()

	c := 500
	for i := 0; i < c; i++ {
		itoa := strconv.Itoa(i)
		cache.Put(itoa, itoa)
	}

	for i := 0; i < c; i++ {
		itoa := strconv.Itoa(i)
		_, ok := cache.Get(itoa)
		if !ok {
			t.Errorf("Expected %d before GC", i)
		}
	}

	// ouch
	for i := 1; i < 10; i++ {
		runtime.Gosched()
		runtime.GC()
	}

	for i := 0; i < c; i++ {
		itoa := strconv.Itoa(i)
		_, ok := cache.Get(itoa)
		if ok {
			t.Errorf("Not expected %d after GC", i)
		}
	}

}

func TestWeakCache_MissDeleteAndSnapshotLifecycle(t *testing.T) {
	cache := NewWeakCache()
	if value, ok := cache.Get("missing"); ok || value != "" {
		t.Fatalf("expected miss, got %q %t", value, ok)
	}
	if !cache.Delete("missing") {
		t.Fatal("delete reports completion even when the key is absent")
	}

	cache.Put("a", "1")
	cache.Put("b", "2")
	snapshot, err := cache.Snapshot()
	if err != nil {
		t.Fatalf("snapshot failed: %v", err)
	}
	sink := &recordingSnapshotSink{}
	if err := snapshot.Persist(sink); err != nil {
		t.Fatalf("persist failed: %v", err)
	}
	if !sink.closed || sink.canceled {
		t.Fatalf("expected clean close without cancel, closed=%t canceled=%t", sink.closed, sink.canceled)
	}
	snapshot.Release()

	var payload map[string]string
	if err := json.Unmarshal(sink.Bytes(), &payload); err != nil {
		t.Fatalf("snapshot payload is not json: %v", err)
	}
	if payload["a"] != "1" || payload["b"] != "2" {
		t.Fatalf("unexpected snapshot payload: %#v", payload)
	}

	restored := NewWeakCache()
	if err := restored.Restore(io.NopCloser(bytes.NewReader(sink.Bytes()))); err != nil {
		t.Fatalf("restore failed: %v", err)
	}
	if value, ok := restored.Get("a"); !ok || value != "1" {
		t.Fatalf("expected restored value, got %q %t", value, ok)
	}
	if err := restored.Restore(io.NopCloser(bytes.NewBufferString("not-json"))); err == nil {
		t.Fatal("expected invalid restore payload to fail")
	}

	failing := &recordingSnapshotSink{failWrite: true}
	if err := snapshot.Persist(failing); err == nil || !failing.canceled {
		t.Fatalf("expected failing persist to cancel, err=%v canceled=%t", err, failing.canceled)
	}
}

type recordingSnapshotSink struct {
	bytes.Buffer
	closed    bool
	canceled  bool
	failWrite bool
	failClose bool
}

func (s *recordingSnapshotSink) ID() string { return "test" }

func (s *recordingSnapshotSink) Close() error {
	s.closed = true
	if s.failClose {
		return errors.New("close failed")
	}
	return nil
}

func (s *recordingSnapshotSink) Cancel() error {
	s.canceled = true
	return nil
}

func (s *recordingSnapshotSink) Write(p []byte) (int, error) {
	if s.failWrite {
		return 0, errors.New("write failed")
	}
	return s.Buffer.Write(p)
}
