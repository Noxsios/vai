package vai

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStepOpteration(t *testing.T) {
	s1 := Step{CMD: "echo hello"}
	s2 := Step{Uses: "other"}
	bad := Step{}

	require.Equal(t, OperationRun, s1.Operation())
	require.Equal(t, OperationUses, s2.Operation())
	require.Equal(t, OperationUnknown, bad.Operation())
}
