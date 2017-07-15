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
	SyntaxRules map[string]Color
	FileConfigs map[string]Config
}

// Color defines background/forground color and attributes for text.
// The 'Clobber' bool says to color this pattern even if it is nested
// within another.
type Color struct {
	FG      string
	BG      string
	Attrs   []string
	Clobber bool
}

// Dup deep copies a Color struct.
func (color Color) Dup() Color {
	newColor := Color{
		FG:      color.FG,
		BG:      color.BG,
		Attrs:   append([]string{}, color.Attrs...),
		Clobber: color.Clobber,
	}
	return newColor
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

// Dup deep copies a Config struct.
func (config Config) Dup() Config {
	newCfg := Config{
		TabString:     config.TabString,
		TabString_set: config.TabString_set,
		TabWidth:      config.TabWidth,
		TabWidth_set:  config.TabWidth_set,
		AutoTab:       config.AutoTab,
		AutoTab_set:   config.AutoTab_set,
		TabDetect:     config.TabDetect,
		TabDetect_set: config.TabDetect_set,
		FmtCmd:        config.FmtCmd,
		Parent:        config.Parent,
		ExtMap:        map[string]string{},
		FileConfigs:   map[string]Config{},
		SyntaxRules:   map[string]Color{},
	}
	for k, v := range config.ExtMap {
		newCfg.ExtMap[k] = v
	}
	for k, v := range config.FileConfigs {
		newCfg.FileConfigs[k] = v.Dup()
	}
	for k, v := range config.SyntaxRules {
		newCfg.SyntaxRules[k] = v.Dup()
	}
	return newCfg
}

// Merge a config into this config.
func (config Config) Merge(other Config) Config {

	// We don't need to check for nil maps, b/c Dup() will initialize all maps.
	config = config.Dup()
	other = other.Dup()

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
	for pattern, color := range other.SyntaxRules {
		config.SyntaxRules[pattern] = color
	}

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
	if other.Parent != "" {
		config.Parent = other.Parent
	}
	if other.FmtCmd != "" {
		config.FmtCmd = other.FmtCmd
	}

	return config
}

// MergeParent merges the parent config into the child config.
// It does this recursively.
func (config Config) MergeParent(level int) Config {
	if level <= 0 {
		return config
	}

	parentConfig, ok := config.FileConfigs[config.Parent]
	if !ok {
		return config
	}

	parentConfig = config.Merge(parentConfig).MergeParent(level - 1)

	config = parentConfig.Merge(config)
	config.Parent = parentConfig.Parent

	return config
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
