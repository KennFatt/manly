// Package config loads manly's user configuration.
package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

const defaultYAML = `root: ~/.okf

defaults:
  format: compact
  list:
    recursive: false

display:
  actions: true
  usage: true

analytics:
  enabled: true
  provider: sqlite
`

// Config is the validated runtime configuration for one invocation.
type Config struct {
	Root      string    `koanf:"root"`
	Defaults  Defaults  `koanf:"defaults"`
	Display   Display   `koanf:"display"`
	Analytics Analytics `koanf:"analytics"`
}

// Defaults contains command defaults.
type Defaults struct {
	Format string       `koanf:"format"`
	List   ListDefaults `koanf:"list"`
}

// ListDefaults contains list command defaults.
type ListDefaults struct {
	Recursive bool `koanf:"recursive"`
}

// Display controls optional list and show presentation elements.
type Display struct {
	Actions bool `koanf:"actions"`
	Usage   bool `koanf:"usage"`
}

// Analytics controls local concept-usage recording and reporting.
type Analytics struct {
	Enabled  bool   `koanf:"enabled"`
	Provider string `koanf:"provider"`
}

// FilePath returns the resolved user configuration path for home.
func FilePath(home string) (string, error) {
	absoluteHome, err := resolveHome(home)
	if err != nil {
		return "", err
	}
	return filepath.Join(absoluteHome, ".config", "manly", "config.yml"), nil
}

func resolveHome(home string) (string, error) {
	if strings.TrimSpace(home) == "" {
		return "", errors.New("config: home directory is empty")
	}
	absoluteHome, err := filepath.Abs(home)
	if err != nil {
		return "", fmt.Errorf("config: resolve home directory: %w", err)
	}
	return absoluteHome, nil
}

// Load resolves and loads the user's configuration. A missing configuration
// is bootstrapped with the built-in defaults; an existing file is never
// rewritten.
func Load(home string) (Config, error) {
	absoluteHome, err := resolveHome(home)
	if err != nil {
		return Config{}, err
	}
	path := filepath.Join(absoluteHome, ".config", "manly", "config.yml")
	if err := ensureFile(path); err != nil {
		return Config{}, fmt.Errorf("config %s: %w", path, err)
	}

	k := koanf.New(".")
	if err := setDefaults(k, absoluteHome); err != nil {
		return Config{}, fmt.Errorf("config %s: set defaults: %w", path, err)
	}
	if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
		return Config{}, fmt.Errorf("config %s: load: %w", path, err)
	}

	var result Config
	decoderConfig := &mapstructure.DecoderConfig{
		TagName:          "koanf",
		Result:           &result,
		ErrorUnused:      true,
		WeaklyTypedInput: false,
	}
	if err := k.UnmarshalWithConf("", &result, koanf.UnmarshalConf{DecoderConfig: decoderConfig}); err != nil {
		return Config{}, fmt.Errorf("config %s: decode: %w", path, err)
	}
	if err := validate(&result); err != nil {
		return Config{}, fmt.Errorf("config %s: %w", path, err)
	}
	result.Root = expandHome(result.Root, absoluteHome)
	return result, nil
}

func setDefaults(k *koanf.Koanf, home string) error {
	defaults := map[string]any{
		"root":                    filepath.Join(home, ".okf"),
		"defaults.format":         "compact",
		"defaults.list.recursive": false,
		"display.actions":         true,
		"display.usage":           true,
		"analytics.enabled":       true,
		"analytics.provider":      "sqlite",
	}
	for key, value := range defaults {
		if err := k.Set(key, value); err != nil {
			return err
		}
	}
	return nil
}

func validate(result *Config) error {
	result.Defaults.Format = strings.ToLower(strings.TrimSpace(result.Defaults.Format))
	switch result.Defaults.Format {
	case "compact", "fancy", "json", "markdown":
		// Supported output format.
	default:
		return fmt.Errorf("defaults.format: unsupported format %q (want compact, fancy, json, or markdown)", result.Defaults.Format)
	}
	if strings.TrimSpace(result.Root) == "" {
		return errors.New("root: value must not be empty")
	}
	result.Analytics.Provider = strings.ToLower(strings.TrimSpace(result.Analytics.Provider))
	switch result.Analytics.Provider {
	case "sqlite", "csv":
		// Supported analytics provider.
	default:
		return fmt.Errorf("analytics.provider: unsupported provider %q (want sqlite or csv)", result.Analytics.Provider)
	}
	return nil
}

func expandHome(value, home string) string {
	value = strings.TrimSpace(value)
	if value == "~" {
		return home
	}
	if strings.HasPrefix(value, "~/") || strings.HasPrefix(value, `~\`) {
		return filepath.Join(home, value[2:])
	}
	return value
}

func ensureFile(path string) error {
	if info, err := os.Stat(path); err == nil {
		if info.IsDir() {
			return errors.New("path is a directory")
		}
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	temporary, err := os.CreateTemp(filepath.Dir(path), ".config.yml-*")
	if err != nil {
		return err
	}
	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)
	if err := temporary.Chmod(0o644); err != nil {
		_ = temporary.Close()
		return err
	}
	if _, err := io.WriteString(temporary, defaultYAML); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Link(temporaryName, path); err != nil {
		if errors.Is(err, os.ErrExist) {
			return nil
		}
		return err
	}
	return nil
}
