package viper

import (
	"strings"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPrecedenceThreeLayerConflict pins down the documented precedence
// (override > flag > env > config > default) when the same key exists in
// multiple layers at once.
func TestPrecedenceThreeLayerConflict(t *testing.T) {
	v := New()

	flagSet := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flagSet.String("name", "flag-default", "")
	require.NoError(t, flagSet.Parse(nil))
	require.NoError(t, v.BindPFlags(flagSet))

	v.SetDefault("name", "from-default")
	v.SetConfigType("yaml")
	require.NoError(t, v.ReadConfig(strings.NewReader("name: from-config\n")))
	t.Setenv("VIPER_TEST_NAME", "from-env")
	require.NoError(t, v.BindEnv("name", "VIPER_TEST_NAME"))

	// env beats config and default
	assert.Equal(t, "from-env", v.GetString("name"))
	assert.Equal(t, SourceEnv, v.GetSource("name"))

	// a changed flag beats env
	require.NoError(t, flagSet.Set("name", "from-flag"))
	assert.Equal(t, "from-flag", v.GetString("name"))
	assert.Equal(t, SourceFlag, v.GetSource("name"))

	// Set() beats everything
	v.Set("name", "from-override")
	assert.Equal(t, "from-override", v.GetString("name"))
	assert.Equal(t, SourceOverride, v.GetSource("name"))
}

func TestGetSourcePerLayer(t *testing.T) {
	v := New()

	flagSet := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flagSet.String("changed", "flag-default", "")
	flagSet.String("unchanged", "flag-default", "")
	require.NoError(t, flagSet.Parse(nil))
	require.NoError(t, flagSet.Set("changed", "from-flag"))
	require.NoError(t, v.BindPFlags(flagSet))

	v.SetDefault("def", "from-default")
	v.SetConfigType("yaml")
	require.NoError(t, v.ReadConfig(strings.NewReader("conf: from-config\n")))
	t.Setenv("VIPER_TEST_ENV", "from-env")
	require.NoError(t, v.BindEnv("env", "VIPER_TEST_ENV"))
	v.Set("ovr", "from-override")

	cases := map[string]Source{
		"ovr":       SourceOverride,
		"changed":   SourceFlag,
		"env":       SourceEnv,
		"conf":      SourceConfig,
		"def":       SourceDefault,
		"unchanged": SourceFlagDefault,
	}
	for key, want := range cases {
		assert.Equalf(t, want, v.GetSource(key), "key %q", key)
	}
}

func TestGetSourceEdgeCases(t *testing.T) {
	v := New()

	// missing key resolves to nothing
	assert.Nil(t, v.Get("missing"))
	assert.Equal(t, SourceNone, v.GetSource("missing"))

	// an empty-string env var is ignored by default (existing semantics)
	v.SetDefault("empty", "from-default")
	t.Setenv("VIPER_TEST_EMPTY", "")
	require.NoError(t, v.BindEnv("empty", "VIPER_TEST_EMPTY"))
	assert.Equal(t, "from-default", v.GetString("empty"))
	assert.Equal(t, SourceDefault, v.GetSource("empty"))

	// ...but counts as a real value once AllowEmptyEnv is enabled
	v.AllowEmptyEnv(true)
	assert.Equal(t, "", v.GetString("empty"))
	assert.Equal(t, SourceEnv, v.GetSource("empty"))

	// an unset env var does not shadow lower layers
	v.SetDefault("fallback", "from-default")
	require.NoError(t, v.BindEnv("fallback", "VIPER_TEST_NOT_SET"))
	assert.Equal(t, "from-default", v.GetString("fallback"))
	assert.Equal(t, SourceDefault, v.GetSource("fallback"))

	// GetSource is case-insensitive, like Get
	assert.Equal(t, SourceDefault, v.GetSource("FALLBACK"))
}

func TestSourceString(t *testing.T) {
	assert.Equal(t, "override", SourceOverride.String())
	assert.Equal(t, "flag", SourceFlag.String())
	assert.Equal(t, "env", SourceEnv.String())
	assert.Equal(t, "config", SourceConfig.String())
	assert.Equal(t, "kvstore", SourceKVStore.String())
	assert.Equal(t, "default", SourceDefault.String())
	assert.Equal(t, "flag-default", SourceFlagDefault.String())
	assert.Equal(t, "none", SourceNone.String())
}
