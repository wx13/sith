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
		SyntaxRules: defaultSyntaxRules(),
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

func defaultSyntaxRules() map[string]Color {
	sr := map[string]Color{
		"[ \t]+$": {BG: "yellow", Clobber: true},
	}
	return sr
}

func defaultFileConfigs() map[string]Config {
	fc := map[string]Config{}
	fc["code"] = Config{
		SyntaxRules: map[string]Color{
			"'.*?'": {FG: "yellow"},
			`".*?"`: {FG: "yellow"},
		},
	}
	fc["sh"] = Config{
		Parent: "code",
		SyntaxRules: map[string]Color{
			"#.*$": {FG: "cyan"},
		},
	}
	fc["c-style"] = Config{
		Parent: "code",
		SyntaxRules: map[string]Color{
			"//.*$":     {FG: "cyan"},
			`/\*.*?\*/`: {FG: "cyan"},
		},
	}
	fc["c"] = Config{
		Parent: "c-style",
		SyntaxRules: map[string]Color{
			"^#[a-z]*": {FG: "blue"},
		},
	}
	fc["go"] = Config{
		Parent: "code",
		SyntaxRules: map[string]Color{
			"//.*$": {FG: "cyan"},
			"'.*?'": {FG: "red"},
			"`.*?`": {FG: "yellow"},
		},
	}
	fc["md"] = Config{
		SyntaxRules: map[string]Color{
			"^#+.*$": {FG: "green"},
			"^===*$": {FG: "green"},
			"^---*$": {FG: "green"},
		},
	}
	fc["toml"] = Config{
		Parent: "sh",
		SyntaxRules: map[string]Color{
			`\[.*?\]`: {FG: "green"},
		},
	}
	fc["git"] = Config{
		SyntaxRules: map[string]Color{
			"#.*?$": {FG: "cyan"},
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
