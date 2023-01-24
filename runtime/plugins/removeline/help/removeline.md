# Remove Line Plugin

The removeline plugin provides current line removal functionality.
The default binding to remove a line is `Ctrl-Delete`. You can easily modify that in your `bindings.json`
file:

```json
{
    "Ctrl-Delete": "removeline.removeline"
}
```

You can also execute a command which will do the same thing as
the binding:

```
> removeline
```

The plugin will override your selection upon execution.
