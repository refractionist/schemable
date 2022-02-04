package schemable

import (
	"context"
	"database/sql"
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"
)

// WhereFunc modifies a basic select operation to add conditions.
//
// Technically, conditions are not limited to adding where clauses. It will receive
// a select statement with the 'SELECT ... FROM tablename' portion composed already.
type WhereFunc func(query sq.SelectBuilder) sq.SelectBuilder

// DeleteFunc modifies a basic delete operation to add conditions.
//
// Technically, conditions are not limited to adding where clauses. It will receive
// a select statement with the 'SELECT ... FROM tablename' portion composed already.
type DeleteFunc func(query sq.DeleteBuilder) sq.DeleteBuilder

// WithClient returns a modified variant of the given context with an embedded
// client.
func WithClient(ctx context.Context, c Client) context.Context {
	return context.WithValue(ctx, clientKey, c)
}

// ClientFrom fetches the embedded client from the given context.
func ClientFrom(ctx context.Context) Client {
	c, _ := ctx.Value(clientKey).(Client)
	return c
}

// WithTransaction begins a new transaction with a *DBClient in the given
// context, returning a new context with the *TxnClient.
func WithTransaction(ctx context.Context, opts *sql.TxOptions) (context.Context, *TxnClient, error) {
	c, ok := ClientFrom(ctx).(*DBClient)
	if !ok || c == nil {
		return ctx, nil, errors.New("no *schemable.DBClient in context.")
	}

	t, err := c.Begin(ctx, opts)
	return WithClient(ctx, t), t, err
}

// Targets returns a slice of target records from the given Recorders.
func Targets[T any](recs []*Recorder[T]) []*T {
	targets := make([]*T, len(recs))
	for i, r := range recs {
		targets[i] = r.Target
	}
	return targets
}

// WithDBDuration wraps the context with the db execution duration based on
// the given start time.
func WithDBDuration(ctx context.Context, start time.Time) context.Context {
	return context.WithValue(ctx, dbDurKey, time.Since(start))
}

// DBDurationFrom extracts db execution duration from the given context.
func DBDurationFrom(ctx context.Context) time.Duration {
	return ctx.Value(dbDurKey).(time.Duration)
}

// QueryLogger is a wrapper for a type that logs SQL queries.
type QueryLogger interface {
	LogQuery(ctx context.Context, q string, args []any)
}

type noLogger struct{}

func (l *noLogger) LogQuery(ctx context.Context, q string, args []any) {
}

var nilLogger = &noLogger{}


var ErrNoClient = errors.New("no client in context")

type key int

var (
	clientKey = key(1)
	dbDurKey = key(3)
)
