package daemons

import (
	"context"
)

// Daemon is a func that should be run as goroutines
type Daemon func(context.Context)
