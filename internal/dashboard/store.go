//go:build dashboard

package dashboard

import "github.com/ernesto2108/anvil/internal/instrumentation"

// Store is the read/write interface the dashboard package depends on.
// *store.SQLiteStore satisfies this interface.
type Store interface {
	// WriteEvent persists an instrumentation event to the underlying store.
	WriteEvent(ev instrumentation.Event) error
	// Close releases any resources held by the store.
	Close() error
}
