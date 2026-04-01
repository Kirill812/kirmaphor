package queries

import "errors"

// ErrNotFound is returned when a database operation affects no rows.
var ErrNotFound = errors.New("not found")
