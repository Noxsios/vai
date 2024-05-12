package vai

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// helloWorldWorkflow is a simple workflow that prints "Hello World!"
// do not make changes to this variable within tests
var helloWorldWorkflow = Workflow{"default": {Step{CMD: "echo 'Hello World!'"}}, "a-task": {Step{CMD: "echo 'task a'"}}, "task-b": {Step{CMD: "echo 'task b'"}}}

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

func TestOrderedTaskNames(t *testing.T) {
	names := helloWorldWorkflow.OrderedTaskNames()
	expected := []string{"default", "a-task", "task-b"}
	require.ElementsMatch(t, expected, names)

	wf := Workflow{"foo": nil, "bar": nil, "baz": nil, "default": nil}
	names = wf.OrderedTaskNames()
	expected = []string{"default", "bar", "baz", "foo"}
	require.ElementsMatch(t, expected, names)

	wf["default"] = nil

	names = wf.OrderedTaskNames()
	expected = []string{"default", "bar", "baz", "foo"}
	require.ElementsMatch(t, expected, names)
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
