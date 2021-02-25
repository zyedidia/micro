package config

import (
	"testing"

	"github.com/blang/semver"

	"github.com/zyedidia/json5"
)

func TestDependencyResolving(t *testing.T) {
	js := `
[{
  "Name": "Foo",
  "Versions": [{ "Version": "1.0.0" }, { "Version": "1.5.0" },{ "Version": "2.0.0" }]
}, {
  "Name": "Bar",
  "Versions": [{ "Version": "1.0.0", "Require": {"Foo": ">1.0.0 <2.0.0"} }]
}, {
  "Name": "Unresolvable",
  "Versions": [{ "Version": "1.0.0", "Require": {"Foo": "<=1.0.0", "Bar": ">0.0.0"} }]
	}]
`
	var all PluginPackages
	err := json5.Unmarshal([]byte(js), &all)
	if err != nil {
		t.Error(err)
	}
	selected, err := all.Resolve(PluginVersions{}, PluginDependencies{
		&PluginDependency{"Bar", semver.MustParseRange(">=1.0.0")},
	})

	check := func(name, version string) {
		v := selected.find(name)
		expected := semver.MustParse(version)
		if v == nil {
			t.Errorf("Failed to resolve %s", name)
		} else if expected.NE(v.Version) {
			t.Errorf("%s resolved in wrong version %v", name, v)
		}
	}

	if err != nil {
		t.Error(err)
	} else {
		check("Foo", "1.5.0")
		check("Bar", "1.0.0")
	}

	selected, err = all.Resolve(PluginVersions{}, PluginDependencies{
		&PluginDependency{"Unresolvable", semver.MustParseRange(">0.0.0")},
	})
	if err == nil {
		t.Error("Unresolvable package resolved:", selected)
	}
}
