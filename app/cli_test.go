package app

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/cybersamx/teapot/model"
	"github.com/kylelemons/godebug/pretty"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type targetVars struct {
	// Targets of all data types supported by BindFlagsToCommand.
	str        string
	strReader  string
	strEnv     string
	strFlag    string
	num        int
	b          bool
	single     map[string]string
	singleJSON map[string]string
	multi      map[string]string
	list       []string
	timeout    time.Duration
}

func newCommand(t *testing.T, targets *targetVars, reader io.Reader) *cobra.Command {
	bindings := []model.FlagBinding{
		{
			Usage:   "str with default value",
			Flag:    "str",
			Target:  &targets.str,
			Default: "default-str",
		},
		{
			Usage:   "str to be overridden by reader",
			Flag:    "str-reader",
			Target:  &targets.strReader,
			Default: "reader-str",
		},
		{
			Usage:   "str to be overridden by env variable",
			Flag:    "str-env",
			Target:  &targets.strEnv,
			Default: "env-str",
		},
		{
			Usage:   "str to be overridden by flag",
			Flag:    "str-flag",
			Target:  &targets.strFlag,
			Default: "flag-str",
		},
		{
			Usage:   "test num",
			Flag:    "num",
			Target:  &targets.num,
			Default: 123,
		},
		{
			Usage:   "test b",
			Flag:    "nested.long-long-name",
			Target:  &targets.b,
			Default: false,
		},
		{
			Usage:   "test single",
			Flag:    "single",
			Target:  &targets.single,
			Default: map[string]string{"key": "default"},
		},
		{
			Usage:   "test single (in json format)",
			Flag:    "single-json",
			Target:  &targets.singleJSON,
			Default: map[string]string{"key": "defaultJson"},
		},
		{
			Usage:   "test multi",
			Flag:    "multi",
			Target:  &targets.multi,
			Default: map[string]string{"map.key1": "map.default1", "map.key2": "map.default2"},
		},
		{
			Usage:   "string slice of names",
			Flag:    "list",
			Target:  &targets.list,
			Default: []string{"list.default1", "list.default2", "list.default3"},
		},
		{
			Usage:   "time duration",
			Flag:    "timeout",
			Target:  &targets.timeout,
			Default: time.Hour + 2*time.Minute + 3*time.Second,
		},
	}

	v := NewViper()
	cmd := cobra.Command{
		Use: "app_test",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	if reader != nil {
		v.SetConfigType("yaml")
		err := v.ReadConfig(reader)
		require.NoError(t, err)
	}

	err := BindFlagsToCommand(v, cmd.Flags(), bindings)
	require.NoError(t, err)

	return &cmd
}

func TestCLI_CommandWithDefaults(t *testing.T) {
	targets := targetVars{}

	cmd := newCommand(t, &targets, nil)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.NoError(t, err)

	want := targetVars{
		str:        "default-str",
		strReader:  "reader-str",
		strEnv:     "env-str",
		strFlag:    "flag-str",
		num:        123,
		b:          false,
		single:     map[string]string{"key": "default"},
		singleJSON: map[string]string{"key": "defaultJson"},
		multi:      map[string]string{"map.key1": "map.default1", "map.key2": "map.default2"},
		list:       []string{"list.default1", "list.default2", "list.default3"},
		timeout:    time.Hour + 2*time.Minute + 3*time.Second,
	}

	diff := pretty.Compare(want, targets)
	assert.Emptyf(t, diff, "want: %+v, got: %+v", want, targets)
}

func TestCLI_CommandWithReader(t *testing.T) {
	targets := targetVars{}

	yamlConfig := []byte(`
str-reader: changed-by-reader
num: 654
nested:
  long-long-name: true
single:
  key: reader
single-json:
  key: readerJson
multi:
  map.key1: map.reader1
  map.key2: map.reader2
list:
  - list.reader1
  - list.reader2
  - list.reader3
timeout: 4h5m6s
`)
	reader := bytes.NewBuffer(yamlConfig)

	cmd := newCommand(t, &targets, reader)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.NoError(t, err)

	want := targetVars{
		str:        "default-str",
		strReader:  "changed-by-reader",
		strEnv:     "env-str",
		strFlag:    "flag-str",
		num:        654,
		b:          true,
		single:     map[string]string{"key": "reader"},
		singleJSON: map[string]string{"key": "readerJson"},
		multi:      map[string]string{"map.key1": "map.reader1", "map.key2": "map.reader2"},
		list:       []string{"list.reader1", "list.reader2", "list.reader3"},
		timeout:    4*time.Hour + 5*time.Minute + 6*time.Second,
	}

	diff := pretty.Compare(want, targets)
	assert.Emptyf(t, diff, "want: %+v, got: %+v", want, targets)
}

func TestCLI_CommandWithEnvs(t *testing.T) {
	t.Setenv("AX_STR_ENV", "changed-by-env")
	t.Setenv("AX_NUM", "456")
	t.Setenv("AX_NESTED_LONG_LONG_NAME", "true")
	t.Setenv("AX_SINGLE", "key=env")                 // Using key=value encoding
	t.Setenv("AX_SINGLE_JSON", `{"key": "envJson"}`) // Using json encoding
	t.Setenv("AX_MULTI", `{"map.key1": "map.env1", "map.key2": "map.env2"}`)
	t.Setenv("AX_LIST", `"list,env1","list,env2","list,env3"`) // Using more obscure values with a comma
	t.Setenv("AX_TIMEOUT", "7h8m9s")

	targets := targetVars{}

	cmd := newCommand(t, &targets, nil)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.NoError(t, err)

	want := targetVars{
		str:        "default-str",
		strReader:  "reader-str",
		strEnv:     "changed-by-env",
		strFlag:    "flag-str",
		num:        456,
		b:          true,
		single:     map[string]string{"key": "env"},
		singleJSON: map[string]string{"key": "envJson"},
		multi:      map[string]string{"map.key1": "map.env1", "map.key2": "map.env2"},
		list:       []string{"list,env1", "list,env2", "list,env3"},
		timeout:    7*time.Hour + 8*time.Minute + 9*time.Second,
	}

	diff := pretty.Compare(want, targets)
	assert.Emptyf(t, diff, "want=%+v, actual=%+v", want, targets)
}

func TestCLI_CommandWithPFlags(t *testing.T) {
	targets := targetVars{}

	cmd := newCommand(t, &targets, nil)
	cmd.SetArgs([]string{
		"--str-flag=changed-by-flag",
		"--num=321",
		"--nested.long-long-name",
		"--single",
		"key=flag",
		"--single-json",
		"key=flagJson",
		"--multi",
		"map.key1=map.flag1",
		"--multi",
		"map.key2=map.flag2",
		"--list=list.flag1",
		"--list=list.flag2",
		"--list=list.flag3",
		"--timeout=9h8m7s",
	})
	err := cmd.Execute()
	require.NoError(t, err)

	want := targetVars{
		str:        "default-str",
		strReader:  "reader-str",
		strEnv:     "env-str",
		strFlag:    "changed-by-flag",
		num:        321,
		b:          true,
		single:     map[string]string{"key": "flag"},
		singleJSON: map[string]string{"key": "flagJson"},
		multi:      map[string]string{"map.key1": "map.flag1", "map.key2": "map.flag2"},
		list:       []string{"list.flag1", "list.flag2", "list.flag3"},
		timeout:    9*time.Hour + 8*time.Minute + 7*time.Second,
	}

	diff := pretty.Compare(want, targets)
	assert.Emptyf(t, diff, "want=%+v, actual=%+v", want, targets)
}

func TestCLI_CommandWithOverride(t *testing.T) {
	t.Setenv("AX_STR_ENV", "changed-by-env")

	targets := targetVars{}

	yamlConfig := []byte("str-reader: changed-by-reader")
	reader := bytes.NewBuffer(yamlConfig)

	cmd := newCommand(t, &targets, reader)

	cmd.SetArgs([]string{"--str-flag=changed-by-flag"})
	err := cmd.Execute()
	require.NoError(t, err)

	want := targetVars{
		str:        "default-str",
		strReader:  "changed-by-reader",
		strEnv:     "changed-by-env",
		strFlag:    "changed-by-flag",
		num:        123,
		b:          false,
		single:     map[string]string{"key": "default"},
		singleJSON: map[string]string{"key": "defaultJson"},
		multi:      map[string]string{"map.key1": "map.default1", "map.key2": "map.default2"},
		list:       []string{"list.default1", "list.default2", "list.default3"},
		timeout:    time.Hour + 2*time.Minute + 3*time.Second,
	}

	diff := pretty.Compare(want, targets)
	assert.Emptyf(t, diff, "want: %+v, got: %+v", want, targets)
}
