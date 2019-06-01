# Terminfo parser

This terminfo parser was written by the authors of [tcell](https://github.com/gdamore/tcell). We are using it here
to compile the terminal database if the terminal entry is not found in set of precompiled terminals.

The source for `mkinfo.go` is adapted from tcell's `mkinfo` tool to be more of a library.
