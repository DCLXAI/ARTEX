package server

import (
	"context"
	"testing"
)

// newTestServer builds a Server under cancellable context and registers Close so
// restored task runtimes cannot outlive the test TempDir (which would fail
// testing's RemoveAll cleanup with "directory not empty").
func newTestServer(t *testing.T, m *Manager) *Server {
	t.Helper()
	td := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	s := New(ctx, m, td, td, td)
	t.Cleanup(s.Close)
	return s
}

// newTestServerDirs is like newTestServer but lets the caller choose skill/data/key dirs
// (some suites share one TempDir across all three).
func newTestServerDirs(t *testing.T, m *Manager, skillDir, dataDir, keyDir string) *Server {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	s := New(ctx, m, skillDir, dataDir, keyDir)
	t.Cleanup(s.Close)
	return s
}
