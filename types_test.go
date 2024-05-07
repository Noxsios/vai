package vai

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// helloWorldWorkflow is a simple workflow that prints "Hello World!"
// do not make changes to this variable
var helloWorldWorkflow = Workflow{"default": {Step{CMD: "echo 'Hello World!'"}}}

func TestWorkflowFind(t *testing.T) {
	task, err := helloWorldWorkflow.Find(DefaultTaskName)
	require.NoError(t, err)

	require.Len(t, task, 1)
	require.Equal(t, "echo 'Hello World!'", task[0].CMD)

	task, err = helloWorldWorkflow.Find("foo")
	require.Error(t, err)
	require.Nil(t, task)
	require.EqualError(t, err, `task "foo" not found`)
}

func TestWorkflowSchemaGen(t *testing.T) {
	schema := WorkFlowSchema()

	require.NotNil(t, schema)

	b, err := json.Marshal(schema)
	require.NoError(t, err)

	current, err := os.ReadFile("vai.schema.json")
	require.NoError(t, err)

	require.JSONEq(t, string(current), string(b))
}
