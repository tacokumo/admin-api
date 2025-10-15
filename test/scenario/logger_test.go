package scenario_test

import (
	"log/slog"

	"github.com/testcontainers/testcontainers-go/log"
)

type SlogForTestContainers struct{}

// Printf implements log.Logger.
func (s *SlogForTestContainers) Printf(format string, v ...any) {
	slog.Info(format, v...)
}

var _ log.Logger = (*SlogForTestContainers)(nil)
