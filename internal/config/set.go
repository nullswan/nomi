package config

import (
	"fmt"
	"strconv"
	"strings"
)

func SetConfigValue(cfg *Config, key string, value string) error {
	keys := strings.Split(key, ".")
	var err error
	switch keys[0] {
	case "input":
		err = setInputConfigValue(&cfg.Input, keys[1:], value)
	case "output":
		err = setOutputConfigValue(&cfg.Output, keys[1:], value)
	default:
		err = fmt.Errorf("unknown configuration key '%s'", key)
	}
	return err
}

func setInputConfigValue(
	input *InputConfig,
	keys []string,
	value string,
) error {
	if len(keys) != 2 {
		return fmt.Errorf("invalid key for input: %v", keys)
	}
	var enabled *bool
	switch keys[0] {
	case "voice":
		enabled = &input.Voice.Enabled
	default:
		return fmt.Errorf("unknown input key '%s'", keys[0])
	}
	if keys[1] != "enabled" {
		return fmt.Errorf("unknown input property '%s'", keys[1])
	}
	v, err := strconv.ParseBool(value)
	if err != nil {
		return fmt.Errorf(
			"invalid value for input property '%s': %w",
			keys[1],
			err,
		)
	}
	*enabled = v
	return nil
}

func setOutputConfigValue(
	output *OutputConfig,
	keys []string,
	value string,
) error {
	if len(keys) != 2 {
		return fmt.Errorf("invalid key for output: %v", keys)
	}
	switch keys[0] {
	case "sqlite":
		return setOutputDetailConfigValue(&output.Sqlite, keys[1], value)
	default:
		return fmt.Errorf("unknown output key '%s'", keys[0])
	}
}

func setOutputDetailConfigValue(
	odc *OutputDetailConfig,
	key string,
	value string,
) error {
	switch key {
	case "enabled":
		v, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf(
				"invalid value for output property '%s': %w",
				key,
				err,
			)
		}
		odc.Enabled = v
	case "path":
		odc.Path = value
	default:
		return fmt.Errorf("unknown output property '%s'", key)
	}
	return nil
}
