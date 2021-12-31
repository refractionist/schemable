package schemable

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
)

// Client represents a DB or Transaction client that runs queries with a
// squirrel builder.
type Client interface {
	Exec(ctx context.Context, q string, args ...any) (sql.Result, error)
	Query(ctx context.Context, q string, args ...any) (*sql.Rows, error)
	QueryRow(ctx context.Context, q string, args ...any) *sql.Row
	Builder() *sq.StatementBuilderType
}

// QueryLogger is a wrapper for a type that logs SQL queries.
type QueryLogger interface {
	LogQuery(q string, args []any)
}

// DBClient is a wrapper around an *sql.DB instance, a QueryLogger, and a
// squirrel query builder.
type DBClient struct {
	db      *sql.DB
	logger  QueryLogger
	builder *sq.StatementBuilderType
}

// New initiates a new database connection with the given connection string
// options, returning a DBClient. See database/sql#Open.
func New(driver, conn string) (*DBClient, error) {
	db, err := sql.Open(driver, conn)
	if err != nil {
		return nil, err
	}

	builder := sq.StatementBuilder.RunWith(db)
	c := &DBClient{db: db, logger: nilLogger, builder: &builder}
	c.SetLogger(nil)
	return c, nil
}

// SetLogger sets the given logger, or resetting it to a no-op logger if nil.
func (c *DBClient) SetLogger(l QueryLogger) {
	if l == nil {
		c.logger = nilLogger
	} else {
		c.logger = l
	}
}

// Builder is the squirrel query builder for this db connection.
func (c *DBClient) Builder() *sq.StatementBuilderType {
	return c.builder
}

// Exec executes a query without returning any rows. The args are for any
// placeholder parameters in the query.
func (c *DBClient) Exec(ctx context.Context, q string, args ...any) (sql.Result, error) {
	c.logger.LogQuery(q, args)
	return c.db.ExecContext(ctx, q, args...)
}

// Query executes a query that returns rows, typically a SELECT. The args are
// for any placeholder parameters in the query.
func (c *DBClient) Query(ctx context.Context, q string, args ...any) (*sql.Rows, error) {
	c.logger.LogQuery(q, args)
	return c.db.QueryContext(ctx, q, args...)
}

// QueryRow executes a query that is expected to return at most one row.
// QueryRowContext always returns a non-nil value. Errors are deferred until
// Row's Scan method is called. If the query selects no rows, the *Row's Scan
// will return ErrNoRows. Otherwise, the *Row's Scan scans the first selected
// row and discards the rest.
func (c *DBClient) QueryRow(ctx context.Context, q string, args ...any) *sql.Row {
	c.logger.LogQuery(q, args)
	return c.db.QueryRowContext(ctx, q, args...)
}

// Close closes the database and prevents new queries from starting. Close then
// waits for all queries that have started processing on the server to finish.
func (c *DBClient) Close() error {
	return c.db.Close()
}

// Ping verifies the connection to the database is still alive.
func (c *DBClient) Ping(ctx context.Context) error {
	return c.db.PingContext(ctx)
}

// DBClient is a wrapper around an *sql.Tx instance, a QueryLogger, and a
// squirrel query builder.
type TxnClient struct {
	tx      *sql.Tx
	logger  QueryLogger
	builder *sq.StatementBuilderType
}

// Begin starts a transaction. See database/sql#DB.BeginTx.
func (c *DBClient) Begin(ctx context.Context, opts *sql.TxOptions) (*TxnClient, error) {
	tx, err := c.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	builder := sq.StatementBuilder.RunWith(tx)
	return &TxnClient{tx: tx, logger: c.logger, builder: &builder}, nil
}

// Commit commits the transaction.
func (c *TxnClient) Commit() error {
	return c.tx.Commit()
}

// Rollback aborts the transaction.
func (c *TxnClient) Rollback() error {
	return c.tx.Rollback()
}

// Exec executes a query without returning any rows. The args are for any
// placeholder parameters in the query.
func (c *TxnClient) Exec(ctx context.Context, q string, args ...any) (sql.Result, error) {
	c.logger.LogQuery(q, args)
	return c.tx.ExecContext(ctx, q, args...)
}

// Query executes a query that returns rows, typically a SELECT. The args are
// for any placeholder parameters in the query.
func (c *TxnClient) Query(ctx context.Context, q string, args ...any) (*sql.Rows, error) {
	c.logger.LogQuery(q, args)
	return c.tx.QueryContext(ctx, q, args...)
}

// QueryRow executes a query that is expected to return at most one row.
// QueryRowContext always returns a non-nil value. Errors are deferred until
// Row's Scan method is called. If the query selects no rows, the *Row's Scan
// will return ErrNoRows. Otherwise, the *Row's Scan scans the first selected
// row and discards the rest.
func (c *TxnClient) QueryRow(ctx context.Context, q string, args ...any) *sql.Row {
	c.logger.LogQuery(q, args)
	return c.tx.QueryRowContext(ctx, q, args...)
}

// Builder is the squirrel query builder for this db connection.
func (c *TxnClient) Builder() *sq.StatementBuilderType {
	return c.builder
}

type noLogger struct{}

func (l *noLogger) LogQuery(q string, args []any) {
}

var nilLogger = &noLogger{}
