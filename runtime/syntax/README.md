# Micro syntax highlighting files

These are the syntax highlighting files for micro. To install them, just
put all the syntax files in `~/.config/micro/syntax`.

They are taken from Nano, specifically from [this repository](https://github.com/scopatz/nanorc).
Micro syntax files are almost identical to Nano's, except for some key differences:

* Micro does not use `icolor`. Instead, for a case insensitive match, use the case insensitive flag (`i`) in the regular expression
    * For example, `icolor green ".*"` would become `color green (i) ".*"`

# Using with colorschemes

Not all of these files have been converted to use micro's colorscheme feature. Most of them just hardcode the colors, which
can be problematic depending on the colorscheme you use.

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

Because the nano syntax files I have modified are distributed under the GNU GPLv3 license, these files are also distributed
under that license. See [LICENSE](LICENSE).
