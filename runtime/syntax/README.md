# Syntax Files

Here are micro's syntax files.

Each yaml file specifies how to detect the filetype based on file extension or headers (first line of the file).
Then there are patterns and regions linked to highlight groups which tell micro how to highlight that filetype.

Making your own syntax files is very simple. I recommend you check the file after you are finished with the
[`syntax_checker.go`](./syntax_checker.go) program (located in this directory). Just place your yaml syntax
file in the current directory and run `go run syntax_checker.go` and it will check every file. If there are no
errors it will print `No issues!`.

You can read more about how to write syntax files (and colorschemes) in the [colors](../help/colors.md) documentation.

# Legacy '.micro' filetype

Micro used to use the `.micro` filetype for syntax files which is no longer supported. If you have `.micro`
syntax files that you would like to convert to the new filetype, you can use the [`syntax_converter.go`](./syntax_converter.go) program (also located in this directory):

```
$ go run syntax_converter.go c.micro > c.yaml
```

Most the the syntax files here have been converted using that tool.

Note that the tool isn't perfect and though it is unlikely, you may run into some small issues that you will have to fix manually
(about 4 files from this directory had issues after being converted).

# Micro syntax highlighting files

These are the syntax highlighting files for micro. To install them, just
put all the syntax files in `~/.config/micro/syntax`.

They are taken from Nano, specifically from [this repository](https://github.com/scopatz/nanorc).
Micro syntax files are almost identical to Nano's, except for some key differences:

* Micro does not use `icolor`. Instead, for a case insensitive match, use the case insensitive flag (`i`) in the regular expression
    * For example, `icolor green ".*"` would become `color green "(?i).*"`

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
