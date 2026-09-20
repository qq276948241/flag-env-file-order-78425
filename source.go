package viper

import "strings"

// Source identifies which layer of the precedence chain produced the
// final value for a key.
//
// The precedence chain is (highest wins):
//
//  1. SourceOverride  - explicit Set() calls
//  2. SourceFlag      - bound pflags that have been changed
//  3. SourceEnv       - environment variables (bound or automatic)
//  4. SourceConfig    - the config file
//  5. SourceKVStore   - the remote key/value store
//  6. SourceDefault   - values registered via SetDefault
//  7. SourceFlagDefault - the default value of a bound, unchanged flag
//
// SourceNone means the key resolves to no value at all.
type Source int

const (
	SourceNone Source = iota
	SourceDefault
	SourceFlagDefault
	SourceKVStore
	SourceConfig
	SourceEnv
	SourceFlag
	SourceOverride
)

// String returns a human-readable name for the source, suitable for
// debug logging.
func (s Source) String() string {
	switch s {
	case SourceOverride:
		return "override"
	case SourceFlag:
		return "flag"
	case SourceEnv:
		return "env"
	case SourceConfig:
		return "config"
	case SourceKVStore:
		return "kvstore"
	case SourceDefault:
		return "default"
	case SourceFlagDefault:
		return "flag-default"
	default:
		return "none"
	}
}

// GetSource reports which layer of the precedence chain supplies the
// final value for key, using the same lookup rules as Get.
// It returns SourceNone when the key resolves to no value.
//
// GetSource is case-insensitive for a key.
func GetSource(key string) Source { return v.GetSource(key) }

// GetSource reports which layer of the precedence chain supplies the
// final value for key, using the same lookup rules as Get.
// It returns SourceNone when the key resolves to no value.
//
// GetSource is case-insensitive for a key.
func (v *Viper) GetSource(key string) Source {
	lcaseKey := strings.ToLower(key)
	val, source := v.findWithSource(lcaseKey, true)
	if val == nil {
		return SourceNone
	}
	return source
}
