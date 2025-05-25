# Syntax Files

Here are micro's syntax files.

Each yaml file specifies how to detect the filetype based on file extension or header (first line of the line).
In addition, a signature can be provided to help resolving ambiguities when multiple matching filetypes are detected.
Then there are patterns and regions linked to highlight groups which tell micro how to highlight that filetype.

You can read more about how to write syntax files (and colorschemes) in the [colors](../help/colors.md) documentation.

# Legacy '.micro' filetype

Micro used to use the `.micro` filetype for syntax files which is no longer supported. If you have `.micro`
syntax files that you would like to convert to the new filetype, you can use the [`syntax_converter.go`](./syntax_converter.go) program (also located in this directory):

```
$ go run syntax_converter.go c.micro > c.yaml
```

Most of the syntax files here have been converted using that tool.

Note that the tool isn't perfect and though it is unlikely, you may run into some small issues that you will have to fix manually
(about 4 files from this directory had issues after being converted).

# Micro syntax highlighting files

These are the syntax highlighting files for micro. To install them, just
put all the syntax files in `~/.config/micro/syntax`.

They are taken from Nano, specifically from [this repository](https://github.com/scopatz/nanorc).
Micro syntax files are almost identical to Nano's, except for some key differences:

* Micro does not use `icolor`. Instead, for a case insensitive match, use the case insensitive flag (`i`) in the regular expression
    * For example, `icolor green ".*"` would become `color green "(?i).*"`

# Incompatibilities with older versions of micro

With PR [#3458](https://github.com/zyedidia/micro/pull/3458) resp. commit
[a9b513a](https://github.com/zyedidia/micro/commit/a9b513a28adaaa7782505dc1e284e1a0132cb66f)
empty `rules: []` definitions are removed from all syntax files, since
`rules` are no longer mandatory.
Unfortunately they are mandatory for `micro` versions up to and including `v2.0.14`.

To use newer syntax definitions from this repository with older `micro` versions
you have to add these `rules: []` to all regions not including `rules` already.
Otherwise you need to use syntax definitions before the above mentioned PR
for example from version [v2.0.14](https://github.com/zyedidia/micro/tree/v2.0.14).

# Using with colorschemes

Not all of these files have been converted to use micro's colorscheme feature. Most of them just hardcode the colors, which can be problematic depending on the colorscheme you use.

Here is a list of the files that have been converted to properly use colorschemes:

* vi
* go
* c
* d
* markdown
* html
* lua
* swift
* rust
* java
* javascript
* pascal
* python
* ruby
* sh
* git
* tex
* solidity

# License

See [LICENSE](LICENSE).
