package memory_test

import (
	"testing"

	"airshift/openmos/internal/repository"
	"airshift/openmos/internal/repository/memory"
	"airshift/openmos/internal/repository/repotest"
)

func TestContract(t *testing.T) {
	repotest.Run(t, func(*testing.T) repository.Repository { return memory.New() })
}
