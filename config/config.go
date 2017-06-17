package config

import (
	"github.com/BurntSushi/toml"
	"strings"
)

// Config holds all configuration data. It will be read in from a config file.
// Elementary values each have a boolean *_set variable to mark a variable as set.
type Config struct {
	TabString     string
	TabString_set bool

	TabWidth     int
	TabWidth_set bool

	AutoTab     bool
	AutoTab_set bool

	TabDetect     bool
	TabDetect_set bool

	FmtCmd      string
	Parent      string
	ExtMap      map[string]string
	SyntaxRules []SyntaxConfig
	FileConfigs map[string]Config
}

// SyntaxConfig defines one syntax coloring specification.
type SyntaxConfig struct {
	Pattern string
	FG      string
	BG      string
	Attrs   []string
}

// Read the configuration file(s).
func Read(paths ...string) Config {
	config := Config{}
	for _, path := range paths {
		// Read from the file.
		meta, err := toml.DecodeFile(path, &config)
		if err != nil {
			continue
		}
		// Mark elementary values as set.
		config.markAsSet(meta, "")
		for ext, cfg := range config.FileConfigs {
			cfg.markAsSet(meta, "fileconfigs."+ext+".")
			config.FileConfigs[ext] = cfg
		}
	}
	return config
}

func (config *Config) markAsSet(meta toml.MetaData, prefix string) {
	for _, key := range meta.Keys() {
		switch strings.ToLower(key.String()) {
		case prefix + "tabstring":
			config.TabString_set = true
		case prefix + "tabwidth":
			config.TabWidth_set = true
		case prefix + "autotab":
			config.AutoTab_set = true
		case prefix + "tabdetect":
			config.TabDetect_set = true
		}
	}
}

// Merge a config into this config.
func (config Config) Merge(other Config) Config {

	// Watch out for nil maps.
	if config.ExtMap == nil {
		config.ExtMap = map[string]string{}
	}
	if other.ExtMap == nil {
		other.ExtMap = map[string]string{}
	}
	if config.FileConfigs == nil {
		config.FileConfigs = map[string]Config{}
	}
	if other.FileConfigs == nil {
		other.FileConfigs = map[string]Config{}
	}

	// Copy map/array values over.
	for ext, ft := range other.ExtMap {
		config.ExtMap[ext] = ft
	}
	for ext, ofc := range other.FileConfigs {
		fc, ok := config.FileConfigs[ext]
		if ok {
			config.FileConfigs[ext] = fc.Merge(ofc)
		} else {
			config.FileConfigs[ext] = ofc
		}
	}
	config.SyntaxRules = append(config.SyntaxRules, other.SyntaxRules...)

	// Set the elementary values.
	if other.AutoTab_set {
		config.AutoTab = other.AutoTab
		config.AutoTab_set = true
	}
	if other.TabWidth_set {
		config.TabWidth = other.TabWidth
		config.TabWidth_set = true
	}
	if other.TabDetect_set {
		config.TabDetect = other.TabDetect
		config.TabDetect_set = true
	}
	if other.TabString_set {
		config.TabString = other.TabString
		config.TabString_set = true
	}
	config.Parent = other.Parent
	if other.FmtCmd != "" {
		config.FmtCmd = other.FmtCmd
	}

	return config
}

func (config Config) MergeParent(level int) Config {
	if level <= 0 {
		return config
	}

	parentConfig, ok := config.FileConfigs[config.Parent]
	if !ok {
		return config
	}

	parentConfig = config.Merge(parentConfig).MergeParent(level - 1)

	return parentConfig.Merge(config)
}

// ForExt tailors a config for a specific file extension.
func (config Config) ForExt(ext string) Config {

	var ok bool
	var ft string
	var newConfig Config

	// Map to another extension (possibly).
	ft, ok = config.ExtMap[ext]
	if ok {
		ext = ft
	}

	// Grab the extension-specific config and merge.
	newConfig, ok = config.FileConfigs[ext]
	if !ok {
		return config
	}
	config = config.Merge(newConfig)

	// Get the parent config and merge.
	config = config.MergeParent(5)

	return config
}
