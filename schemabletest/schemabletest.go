package schemabletest

import (
	"context"
	"testing"

	"github.com/refractionist/schemable"
)

type TestStruct struct {
	ID      int64  `db:"id, PRIMARY KEY, AUTO INCREMENT"`
	ID2     int64  `db:"id_two, PRIMARY KEY"`
	Name    string `db:"name"`
	Num     int    `db:"num"`
	Ignored string
}

var TestSchemer = schemable.Bind[TestStruct]("test_structs")

func assertExists(t *testing.T, ctx context.Context, r *schemable.Recorder[TestStruct]) {
	t.Helper()
	ok, err := r.Exists(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if !ok {
		t.Fatalf("does not exist: %+v", r.Target)
	}
}

func refuteExists(t *testing.T, ctx context.Context, r *schemable.Recorder[TestStruct]) {
	t.Helper()
	ok, err := r.Exists(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if ok {
		t.Fatalf("exists: %+v", r.Target)
	}
}

func recorderErr(t *testing.T, r *schemable.Recorder[TestStruct]) {
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