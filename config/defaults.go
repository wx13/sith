package config

import (
	"os/user"
	"path/filepath"
)

func defaultConfig() Config {
	return Config{
		TabString: "\t",
		TabWidth:  4,
		AutoTab:   true,
		TabDetect: true,

		TabString_set: true,
		TabWidth_set:  true,
		AutoTab_set:   true,
		TabDetect_set: true,

		Parent: "",

		ExtMap:      defaultExtMap(),
		FileConfigs: defaultFileConfigs(),
	}
}

func defaultExtMap() map[string]string {
	em := map[string]string{}
	for _, ext := range []string{"csh", "ksh"} {
		em[ext] = "sh"
	}
	for _, ext := range []string{"h", "cc", "C", "c++", "cpp"} {
		em[ext] = "c"
	}
	return em
}

func defaultFileConfigs() map[string]Config {
	fc := map[string]Config{}
	fc["code"] = Config{
		SyntaxRules: []SyntaxConfig{
			{Pattern: "'.*?'", FG: "yellow"},
			{Pattern: `".*?"`, FG: "yellow"},
		},
	}
	fc["sh"] = Config{
		Parent: "code",
		SyntaxRules: []SyntaxConfig{
			{Pattern: "#.*$", FG: "cyan"},
		},
	}
	fc["c"] = Config{
		Parent: "code",
		SyntaxRules: []SyntaxConfig{
			{Pattern: "//.*$", FG: "cyan"},
			{Pattern: `/\*.*?\*/`, FG: "cyan"},
		},
	}
	fc["go"] = Config{
		Parent: "code",
		SyntaxRules: []SyntaxConfig{
			{Pattern: "//.*$", FG: "cyan"},
			{Pattern: "'.*?'", FG: "red"},
			{Pattern: "`.*?`", FG: "yellow"},
		},
	}
	fc["md"] = Config{
		SyntaxRules: []SyntaxConfig{
			{Pattern: "^#+.*$", FG: "green"},
			{Pattern: "^===*$", FG: "green"},
			{Pattern: "^---*$", FG: "green"},
		},
	}
	fc["toml"] = Config{
		Parent: "sh",
		SyntaxRules: []SyntaxConfig{
			{Pattern: `\[.*?\]`, FG: "green"},
		},
	}

	return fc
}

func CreateConfig() Config {
	cfg := readConfig()
	return defaultConfig().Merge(cfg)
}

func readConfig() Config {
	user, err := user.Current()
	if err != nil {
		return Config{}
	}
	home := user.HomeDir
	cfg := Read(
		filepath.Join(home, ".sith.toml"),
		filepath.Join(home, ".sith/config.toml"),
		filepath.Join(home, ".config/sith/config.toml"),
	)
	return cfg
}
