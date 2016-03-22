# Micro syntax highlighting files

These are the syntax highlighting files for micro. To install them, just
put all the syntax files in `~/.micro/syntax`.

They are taken from Nano, specifically from [this repository](https://github.com/scopatz/nanorc).
Micro syntax files are almost identical to Nano's, except for some key differences:

* Micro does not use `icolor`. Instead, for a case insensitive match, use the case insensitive flag (`i`) in the regular expression
    * For example, `icolor green ".*"` would become `color green (i) ".*"`
* Micro does not support `start="..." end="..."`. Instead use the `s` flag to match newlines and put `.*?` in the middle
    * For example `color green start="hello" end="world"` would become `color green (s) "hello.*?world"`
