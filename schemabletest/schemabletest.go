package schemabletest

import (
	"context"
	"testing"

	"github.com/refractionist/schemable"
)

type ComicTitle struct {
	ID      int64  `db:"id, PRIMARY KEY, AUTO INCREMENT"`
	ID2     int64  `db:"id_two, PRIMARY KEY"`
	Name    string `db:"name"`
	Volume  int    `db:"volume"`
	Ignored string
}

var ComicTitles = schemable.Bind[ComicTitle]("comic_titles")

func assertExists(t *testing.T, ctx context.Context, r *schemable.Recorder[ComicTitle]) {
	t.Helper()
	ok, err := r.Exists(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if !ok {
		t.Fatalf("does not exist: %+v", r.Target)
	}
}

func refuteExists(t *testing.T, ctx context.Context, r *schemable.Recorder[ComicTitle]) {
	t.Helper()
	ok, err := r.Exists(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if ok {
		t.Fatalf("exists: %+v", r.Target)
	}
}

func recorderErr(t *testing.T, r *schemable.Recorder[ComicTitle]) {
	t.Helper()
	t.Errorf("invalid %T record", r.Target)
	t.Logf("recorder: %+v", r)
	t.Logf("target: %+v", r.Target)
}

func QueryLogger(t *testing.T) schemable.QueryLogger {
	return &testLogger{t: t}
}

type testLogger struct {
	t *testing.T
}

func (l *testLogger) LogQuery(q string, args []any) {
	l.t.Logf("SQL: %s; ARGS: %+v", q, args)
}