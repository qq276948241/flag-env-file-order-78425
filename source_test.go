package viper

import (
	"strings"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPrecedenceThreeLayerConflict pins down the documented resolution
// order (flag > env > config > default) when the same key exists in
// several layers at once.
func TestPrecedenceThreeLayerConflict(t *testing.T) {
	newConflictViper := func(t *testing.T) (*Viper, *pflag.FlagSet) {
		t.Helper()

		v := New()
		v.SetConfigType("yaml")
		require.NoError(t, v.ReadConfig(strings.NewReader("name: from-config\n")))
		v.SetDefault("name", "from-default")

		t.Setenv("VIPER_TEST_NAME", "from-env")
		require.NoError(t, v.BindEnv("name", "VIPER_TEST_NAME"))

		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.String("name", "flag-default", "")
		require.NoError(t, v.BindPFlag("name", fs.Lookup("name")))

		return v, fs
	}

	t.Run("config beats default", func(t *testing.T) {
		v := New()
		v.SetConfigType("yaml")
		require.NoError(t, v.ReadConfig(strings.NewReader("name: from-config\n")))
		v.SetDefault("name", "from-default")

		assert.Equal(t, "from-config", v.GetString("name"))
		assert.Equal(t, SourceConfig, v.GetSource("name"))
	})

	t.Run("env beats config", func(t *testing.T) {
		v := New()
		v.SetConfigType("yaml")
		require.NoError(t, v.ReadConfig(strings.NewReader("name: from-config\n")))
		v.SetDefault("name", "from-default")
		t.Setenv("VIPER_TEST_NAME", "from-env")
		require.NoError(t, v.BindEnv("name", "VIPER_TEST_NAME"))

		assert.Equal(t, "from-env", v.GetString("name"))
		assert.Equal(t, SourceEnv, v.GetSource("name"))
	})

	t.Run("changed flag beats env", func(t *testing.T) {
		v, fs := newConflictViper(t)
		require.NoError(t, fs.Set("name", "from-flag"))

		assert.Equal(t, "from-flag", v.GetString("name"))
		assert.Equal(t, SourceFlag, v.GetSource("name"))
	})

	t.Run("unchanged flag does not shadow env", func(t *testing.T) {
		v, _ := newConflictViper(t)

		assert.Equal(t, "from-env", v.GetString("name"))
		assert.Equal(t, SourceEnv, v.GetSource("name"))
	})

	t.Run("set override beats everything", func(t *testing.T) {
		v, fs := newConflictViper(t)
		require.NoError(t, fs.Set("name", "from-flag"))
		v.Set("name", "from-override")

		assert.Equal(t, "from-override", v.GetString("name"))
		assert.Equal(t, SourceOverride, v.GetSource("name"))
	})

	t.Run("default is the last resort", func(t *testing.T) {
		v := New()
		v.SetDefault("name", "from-default")

		assert.Equal(t, "from-default", v.GetString("name"))
		assert.Equal(t, SourceDefault, v.GetSource("name"))
	})

	t.Run("flag default is reported separately", func(t *testing.T) {
		v := New()
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.String("name", "flag-default", "")
		require.NoError(t, v.BindPFlag("name", fs.Lookup("name")))

		assert.Equal(t, "flag-default", v.GetString("name"))
		assert.Equal(t, SourceFlagDefault, v.GetSource("name"))
	})
}

func TestGetSourceUnsetAndEmptyString(t *testing.T) {
	t.Run("missing key is unset", func(t *testing.T) {
		v := New()

		assert.Equal(t, SourceUnset, v.GetSource("missing"))
		assert.Nil(t, v.Get("missing"))
		assert.False(t, v.IsSet("missing"))
	})

	t.Run("empty string env falls through by default", func(t *testing.T) {
		v := New()
		v.SetDefault("name", "from-default")
		t.Setenv("VIPER_TEST_EMPTY", "")
		require.NoError(t, v.BindEnv("name", "VIPER_TEST_EMPTY"))

		// Long-standing behavior: an empty env var is treated as unset
		// unless AllowEmptyEnv is enabled, so the default wins.
		assert.Equal(t, "from-default", v.GetString("name"))
		assert.Equal(t, SourceDefault, v.GetSource("name"))
	})

	t.Run("empty string env counts with AllowEmptyEnv", func(t *testing.T) {
		v := New()
		v.AllowEmptyEnv(true)
		v.SetDefault("name", "from-default")
		t.Setenv("VIPER_TEST_EMPTY", "")
		require.NoError(t, v.BindEnv("name", "VIPER_TEST_EMPTY"))

		assert.Equal(t, "", v.GetString("name"))
		assert.Equal(t, SourceEnv, v.GetSource("name"))
	})

	t.Run("empty string config value still counts as set", func(t *testing.T) {
		v := New()
		v.SetConfigType("yaml")
		require.NoError(t, v.ReadConfig(strings.NewReader("name: \"\"\n")))
		v.SetDefault("name", "from-default")

		assert.Equal(t, "", v.GetString("name"))
		assert.Equal(t, SourceConfig, v.GetSource("name"))
	})

	t.Run("case insensitive", func(t *testing.T) {
		v := New()
		v.SetDefault("Name", "from-default")

		assert.Equal(t, SourceDefault, v.GetSource("NAME"))
		assert.Equal(t, SourceDefault, v.GetSource("name"))
	})
}

func TestSourceString(t *testing.T) {
	assert.Equal(t, "unset", SourceUnset.String())
	assert.Equal(t, "override", SourceOverride.String())
	assert.Equal(t, "flag", SourceFlag.String())
	assert.Equal(t, "env", SourceEnv.String())
	assert.Equal(t, "config", SourceConfig.String())
	assert.Equal(t, "kvstore", SourceKVStore.String())
	assert.Equal(t, "default", SourceDefault.String())
	assert.Equal(t, "flag-default", SourceFlagDefault.String())
}
