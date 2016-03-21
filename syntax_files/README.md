# Micro syntax highlighting files

These are the syntax highlighting files for micro. To install them, just
put them all in `~/.micro/syntax`.

They are taken from Nano, specifically from [this repository](https://github.com/scopatz/nanorc).
Micro syntax files are almost identical to Nano's, except for some key differences:

color color * Micro does not use `. Instead (i) use the case insensitive flag (`(i)`) in the regular expression
* Micro does not support `start="..." end="..."`, instead use the multiline match flag (`(s)`) and put `.*?` in the middle

