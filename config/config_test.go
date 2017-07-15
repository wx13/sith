package config_test

import (
	"github.com/wx13/sith/config"
	"io/ioutil"
	"os"
	"testing"
)

func writeTempFile(contents string) string {
	b := []byte(contents)
	tmpfile, err := ioutil.TempFile("", "config_test")
	if err != nil {
		panic(err)
	}
	_, err = tmpfile.Write(b)
	if err != nil {
		panic(err)
	}
	return tmpfile.Name()
}

func TestReadLowercase(t *testing.T) {
	contents := "" +
		"autotab = true\n" +
		"tabwidth = 3\n"
	path := writeTempFile(contents)
	defer os.Remove(path)
	cfg := config.Read(path)
	if cfg.AutoTab != true || cfg.TabWidth != 3 {
		t.Error("config file wrong", cfg)
	}
	if cfg.AutoTab_set != true || cfg.TabWidth_set != true {
		t.Error("Should be set", cfg)
	}
	if cfg.TabString_set == true || cfg.TabDetect_set == true {
		t.Error("Should not be set", cfg)
	}
}

func TestMarkAsSet(t *testing.T) {
	contents := "" +
		"tabwidth = 3\n" +
		"[fileconfigs.toml]\n" +
		"  tabwidth = 5\n"
	path := writeTempFile(contents)
	defer os.Remove(path)
	cfg := config.Read(path)

	if cfg.TabWidth_set != true {
		t.Errorf("elementary values should be set: %+v\n", cfg)
	}
	if cfg.FileConfigs["toml"].TabWidth_set != true {
		t.Errorf("fileconfigs values should be set: %+v\n", cfg)
	}

}

func TestReadUppercase(t *testing.T) {
	contents := "" +
		"AutoTab = true\n" +
		"TabWidth = 3\n"
	path := writeTempFile(contents)
	defer os.Remove(path)
	cfg := config.Read(path)
	if cfg.AutoTab != true || cfg.TabWidth != 3 {
		t.Error("config file wrong", cfg)
	}
	if cfg.AutoTab_set != true || cfg.TabWidth_set != true {
		t.Error("Should be set", cfg)
	}
	if cfg.TabString_set == true || cfg.TabDetect_set == true {
		t.Error("Should not be set", cfg)
	}
}

func TestMerge(t *testing.T) {

	cfg := config.Config{
		TabWidth: 3, TabWidth_set: true,
		AutoTab: true, AutoTab_set: true,
		ExtMap: map[string]string{
			"sh":  "sh",
			"csh": "sh",
		},
	}

	cfg2 := config.Config{
		TabWidth: 5, TabWidth_set: true,
		TabDetect: true, TabDetect_set: true,
		ExtMap: map[string]string{
			"csh": "csh",
			"foo": "foo",
		},
	}

	cfg = cfg.Merge(cfg2)

	if cfg.TabWidth != 5 || cfg.TabDetect != true {
		t.Errorf("Should have overwritten: %+v\n", cfg)
	}
	if cfg.AutoTab != true {
		t.Errorf("Should not be overwritten: %+v\n", cfg)
	}
	if cfg.ExtMap["csh"] != "csh" || cfg.ExtMap["foo"] != "foo" ||
		cfg.ExtMap["sh"] != "sh" {
		t.Errorf("Map didn't merge: %+v\n", cfg)
	}

}

func TestForExt(t *testing.T) {

	cfg := config.Config{
		TabWidth: 3, TabWidth_set: true,
		AutoTab: true, AutoTab_set: true,
		ExtMap: map[string]string{
			"sh":  "sh",
			"csh": "sh",
		},
		FileConfigs: map[string]config.Config{
			"sh": config.Config{
				TabWidth: 5, TabWidth_set: true,
			},
		},
	}

	cfg2 := cfg.ForExt("foo")
	if cfg2.TabWidth != cfg.TabWidth {
		t.Error("Should not have changed.", cfg, cfg2)
	}

	cfg2 = cfg.ForExt("sh")
	if cfg2.TabWidth != 5 {
		t.Error("TabWidth should've been overwritten.", cfg, cfg2)
	}
	if cfg2.AutoTab != true {
		t.Error("AutoTab should have remained set.", cfg, cfg2)
	}

}

func TestForExtWithParent(t *testing.T) {

	cfg := config.Config{
		TabWidth: 3, TabWidth_set: true,
		AutoTab: true, AutoTab_set: true,
		ExtMap: map[string]string{
			"sh":  "sh",
			"csh": "csh",
		},
		FileConfigs: map[string]config.Config{
			"sh": config.Config{
				TabWidth: 5, TabWidth_set: true,
				SyntaxRules: map[string]config.Color{
					"abc": {FG: "green"},
				},
			},
			"csh": config.Config{
				Parent: "sh",
				SyntaxRules: map[string]config.Color{
					"def": {FG: "cyan"},
				},
			},
		},
	}

	cfg2 := cfg.ForExt("sh")
	if len(cfg2.SyntaxRules) != 1 {
		t.Errorf("There should be one syntax rule. %+v %+v",
			cfg.SyntaxRules, cfg2.SyntaxRules)
	}

	cfg2 = cfg.ForExt("csh")
	if len(cfg2.SyntaxRules) != 2 {
		t.Errorf("There should be two syntax rules. %+v %+v",
			cfg.SyntaxRules, cfg2.SyntaxRules)
	}

}

func TestForExtAgainstInfLoop(t *testing.T) {

	cfg := config.Config{
		TabWidth: 3, TabWidth_set: true,
		AutoTab: true, AutoTab_set: true,
		ExtMap: map[string]string{
			"sh":  "sh",
			"csh": "csh",
		},
		FileConfigs: map[string]config.Config{
			"sh": config.Config{
				Parent:   "csh",
				TabWidth: 5, TabWidth_set: true,
				SyntaxRules: map[string]config.Color{
					"abc": {FG: "green"},
				},
			},
			"csh": config.Config{
				Parent: "sh",
				SyntaxRules: map[string]config.Color{
					"def": {FG: "cyan"},
				},
			},
		},
	}

	cfg2 := cfg.ForExt("sh")
	if len(cfg2.SyntaxRules) != 2 {
		t.Errorf("There should be two syntax rule. %+v %+v",
			cfg.SyntaxRules, cfg2.SyntaxRules)
	}

	cfg2 = cfg.ForExt("csh")
	if len(cfg2.SyntaxRules) != 2 {
		t.Errorf("There should be two syntax rules. %+v %+v",
			cfg.SyntaxRules, cfg2.SyntaxRules)
	}

}

func TestEmptyConfigFile(t *testing.T) {
	path := writeTempFile("")
	defer os.Remove(path)
	cfg := config.Read(path)
	if cfg.AutoTab != false || cfg.AutoTab_set != false {
		t.Errorf("Wrong defaults: %+v\n", cfg)
	}
}

func TestColorDup(t *testing.T) {
	color := config.Color{
		BG:    "blue",
		FG:    "yellow",
		Attrs: []string{"bold"},
	}
	color2 := color.Dup()
	color.FG = "white"
	color.Attrs[0] = "underline"
	if color2.BG != "blue" || color2.Attrs[0] != "bold" {
		t.Errorf("Color.Dup() did not work: %+v %+v\n", color, color2)
	}
}

func TestConfigDup(t *testing.T) {

	cfg := config.Config{
		TabWidth: 3, TabWidth_set: true,
		AutoTab: true, AutoTab_set: true,
		ExtMap: map[string]string{
			"sh":  "sh",
			"csh": "csh",
		},
		FileConfigs: map[string]config.Config{
			"sh": config.Config{
				Parent:   "csh",
				TabWidth: 5, TabWidth_set: true,
				SyntaxRules: map[string]config.Color{
					"abc": {FG: "green"},
				},
			},
			"csh": config.Config{
				Parent: "sh",
				SyntaxRules: map[string]config.Color{
					"def": {FG: "cyan"},
				},
			},
		},
	}

	cfg2 := cfg.Dup()
	cfg.ForExt("sh")

	if cfg2.FileConfigs["sh"].SyntaxRules["abc"].FG != "green" {
		t.Error("Dup allowed for overwrite.")
	}

}
