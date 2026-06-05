package database

import (
	"context"
	"time"
)

const (
	// DefaultQueryTimeout is the default timeout for individual database queries.
	DefaultQueryTimeout = 5 * time.Second

	// LongQueryTimeout is for complex reporting queries.
	LongQueryTimeout = 15 * time.Second

	// TransactionTimeout is for operations involving multiple queries in a transaction.
	TransactionTimeout = 10 * time.Second
)

// QueryContext returns a context with the default query timeout.
func QueryContext(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, DefaultQueryTimeout)
}

// LongQueryContext returns a context with extended timeout for reporting.
func LongQueryContext(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, LongQueryTimeout)
}

// TxContext returns a context with transaction timeout.
func TxContext(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, TransactionTimeout)
}
