module github.com/zyedidia/micro

require (
	github.com/blang/semver v3.5.1+incompatible
	github.com/dustin/go-humanize v1.0.0
	github.com/go-errors/errors v1.0.1
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51
	github.com/mattn/go-isatty v0.0.11
	github.com/mattn/go-runewidth v0.0.7
	github.com/mitchellh/go-homedir v1.1.0
	github.com/sergi/go-diff v1.1.0
	github.com/smacker/go-tree-sitter v0.0.0-20191230102415-949ed041aea3
	github.com/stretchr/testify v1.4.0
	github.com/yuin/gopher-lua v0.0.0-20191220021717-ab39c6098bdb
	github.com/zyedidia/clipboard v0.0.0-20190823154308-241f98e9b197
	github.com/zyedidia/glob v0.0.0-20170209203856-dd4023a66dc3
	github.com/zyedidia/highlight v0.0.0-20170330143449-201131ce5cf5
	github.com/zyedidia/json5 v0.0.0-20200102012142-2da050b1a98d
	github.com/zyedidia/pty v2.0.0+incompatible // indirect
	github.com/zyedidia/tcell v1.4.2
	github.com/zyedidia/terminal v0.0.0-20180726154117-533c623e2415
	golang.org/x/text v0.3.2
	gopkg.in/yaml.v2 v2.2.7
	layeh.com/gopher-luar v1.0.7
)

replace github.com/smacker/go-tree-sitter => github.com/p-e-w/go-tree-sitter v0.0.0-20200125032645-7b3cf93b37eb

go 1.11
