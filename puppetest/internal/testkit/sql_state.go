package testkit

import (
	"database/sql/driver"
	"slices"
	"sync"
)

type RecordedQuery struct {
	Statement string
	Args      []driver.Value
}

type SQLState struct {
	PingErr  error
	ExecErr  error
	QueryErr error
	// One canned result reused by every QueryContext call, not sequenced per call.
	QueryCols []string
	QueryRows [][]driver.Value

	mu         sync.Mutex
	execCalls  []RecordedQuery
	queryCalls []RecordedQuery
	closeCount int
}

func (s *SQLState) recordExec(query string, args []driver.Value) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.execCalls = append(s.execCalls, RecordedQuery{Statement: query, Args: slices.Clone(args)})
}

func (s *SQLState) recordQuery(query string, args []driver.Value) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.queryCalls = append(s.queryCalls, RecordedQuery{Statement: query, Args: slices.Clone(args)})
}

func (s *SQLState) recordClose() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closeCount++
}

func (s *SQLState) ExecStatements() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	statements := make([]string, len(s.execCalls))
	for index, call := range s.execCalls {
		statements[index] = call.Statement
	}
	return statements
}

func (s *SQLState) ExecCalls() []RecordedQuery {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.execCalls)
}

func (s *SQLState) QueryCalls() []RecordedQuery {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.queryCalls)
}

func (s *SQLState) CloseCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closeCount
}
