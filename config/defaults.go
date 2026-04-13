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
	em["commit_editmsg"] = "git_commit"
	em["git-rebase-todo"] = "git_rebase"
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
	fc["git_commit"] = Config{
		SyntaxRules: map[string]Color{
			"#.*?$": {FG: "cyan"},
		},
	}
	fc["git_rebase"] = Config{
		Parent: "git_commit",
		SyntaxRules: map[string]Color{
			"^(pick |p )":   {FG: "green"},
			"^(squash |s )": {FG: "yellow"},
			"^(fixup |f )":  {FG: "yellow"},
			"^(drop |d )":   {FG: "red"},
			"^(edit |e )":   {FG: "blue"},
			"^(reword |r )": {FG: "blue"},
			"^(exec |x )":   {FG: "magenta"},
		},
	}

	return fc
}

func CreateConfig() Config {
	cfg := readConfig()
	return defaultConfig().Merge(cfg)
}

func readConfig() Config {
	home := homeDir()
	if home == "" {
		return Config{}
	}
	cfg := Read(
		filepath.Join(home, ".sith.toml"),
		filepath.Join(home, ".sith/config.toml"),
		filepath.Join(home, ".config/sith/config.toml"),
	)
	return cfg
}

func homeDir() string {
	user, err := user.Current()
	if err != nil {
		return ""
	}
	return user.HomeDir
}

// ConfigDir returns the path to the sith config directory (~/.config/sith).
// It creates the directory if it doesn't exist.
func ConfigDir() string {
	home := homeDir()
	if home == "" {
		return ""
	}
	return filepath.Join(home, ".config", "sith")
}
