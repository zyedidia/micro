# Comment Plugin

The comment plugin provides auto commenting/uncommenting.
The default binding to comment/uncomment a line is `Alt-/`,
but you can easily modify that in your `bindings.json` file:

```json
{
    "Alt-g": "comment.comment"
}
```

You can also execute a command which will do the same thing as
the binding:

```
> comment
```

If you have a selection, the plugin will comment all the lines
selected.

The comment type will be auto detected based on the filetype,
but it is only available for certain filetypes:

* c: `// %s`
* c++: `// %s`
* d: `// %s`
* go: `// %s`
* html: `<!-- %s -->`
* java: `// %s`
* javascript: `// %s`
* julia: `# %s`
* lua: `-- %s`
* perl: `# %s`
* php: `// %s`
* python: `# %s`
* python3: `# %s`
* ruby: `# %s`
* rust: `// %s`
* shell: `# %s`
* swift: `// %s`

If your filetype is not available here, you can simply modify
the `commenttype` option:

```
set commenttype "/* %s */"
```

Or in your `settings.json`:

```json
{
    "*.c": {
        "commenttype": "/* %s */"
    }
}
```

