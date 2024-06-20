package daemons

import (
	"context"
)

// Daemons are funcs that should be run as goroutines
type Daemon func(context.Context)
