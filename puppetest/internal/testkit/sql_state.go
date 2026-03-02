package testkit

import "sync"

type SQLState struct {
	PingErr error
	ExecErr error

	mu         sync.Mutex
	execStmts  []string
	closeCount int
}

func (s *SQLState) recordExec(query string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.execStmts = append(s.execStmts, query)
}

func (s *SQLState) recordClose() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closeCount++
}

func (s *SQLState) ExecStatements() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	copied := make([]string, len(s.execStmts))
	copy(copied, s.execStmts)
	return copied
}

func (s *SQLState) CloseCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closeCount
}
