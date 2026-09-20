package viper

import "strings"

// Source identifies which layer of the precedence chain produced the
// effective value for a key.
//
// Viper resolves a key by checking the following layers, in order, and
// returning the first non-nil value found:
//
//  1. SourceOverride     — explicit Set() calls
//  2. SourceFlag         — bound pflags that have been changed
//  3. SourceEnv          — environment variables (bound or automatic)
//  4. SourceConfig       — the config file
//  5. SourceKVStore      — the remote key/value store
//  6. SourceDefault      — values registered via SetDefault
//  7. SourceFlagDefault  — the default value of a bound pflag
//
// If no layer yields a value, the source is SourceUnset.
type Source int

const (
	// SourceUnset means no layer provided a value for the key.
	SourceUnset Source = iota
	// SourceOverride means the value came from an explicit Set() call.
	SourceOverride
	// SourceFlag means the value came from a bound pflag that was changed.
	SourceFlag
	// SourceEnv means the value came from an environment variable.
	SourceEnv
	// SourceConfig means the value came from the config file.
	SourceConfig
	// SourceKVStore means the value came from the remote key/value store.
	SourceKVStore
	// SourceDefault means the value came from SetDefault.
	SourceDefault
	// SourceFlagDefault means the value came from a bound pflag's default.
	SourceFlagDefault
)

// String returns a human-readable name for the source, suitable for logs.
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
		return "unset"
	}
}

// GetSource returns which layer of the precedence chain currently provides
// the effective value for key, or SourceUnset if the key has no value.
//
// GetSource is case-insensitive for a key and follows the same resolution
// rules as Get: an empty-string environment variable only counts as set
// when AllowEmptyEnv is enabled, and shadowed nested paths resolve the
// same way Get does.
func GetSource(key string) Source { return v.GetSource(key) }

// GetSource returns which layer of the precedence chain currently provides
// the effective value for key, or SourceUnset if the key has no value.
//
// GetSource is case-insensitive for a key and follows the same resolution
// rules as Get: an empty-string environment variable only counts as set
// when AllowEmptyEnv is enabled, and shadowed nested paths resolve the
// same way Get does.
func (v *Viper) GetSource(key string) Source {
	lcaseKey := strings.ToLower(key)
	_, source := v.findWithSource(lcaseKey, true)
	return source
}
