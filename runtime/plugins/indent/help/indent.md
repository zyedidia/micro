# `indent`

The `indent` plugin serves one specific task: to autoindent a new line if a current line can be matched against at least one pattern.

Here is an example for a language that has colon-based indentation (e.g., Python):
```json
{
   "indent.patterns": ":%s*$"
}
```

If you want to support comments that start with `#`:
```json
{
   "indent.patterns": ":%s*$|:%s*#.*$"
}
```

Note that `|` is the default separator. If it's conflicting with patterns themselves, you may change it too:
```json
{
   "indent.separators": ";",
   "indent.patterns": ":%s*$;:%s*|.*|$"
}
```

Each character in `indent.separators` is an individual separator. If no characters are present, there is no pattern splitting.
