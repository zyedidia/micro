# About the snippet plugin

This plugin is designed to work with simple VIM snippets like the once from 
[here](https://github.com/honza/vim-snippets/tree/master/snippets)
The plugin itself doesn't provide any snippets. To add snippets for a language
place a file containing the snippets at `~/.config/micro/plugins/snippet/[filetype].snippets`

# Commands

The plugin providies the following commands:

* `snippetinsert`: with an optional parameter to specify which snippet should be inserted. If the parameter is absent, the word right before the cursor will be used for this. (`Alt+S`)
* `snippetnext`: proceeds to the next placeholder (`Alt+W`)
* `snippetcancel`: removes all the current snippet (`Alt+D`)
* `snippetaccept`: finishes the snippet editing for the current snippet. (`Alt+A`)

# Snippet Files

The snippet files have a very simple syntax:
* lines starting with `#` are ignored
* a line starts with `snippet` starts a new snippet.
* after `snippet` you can add one or more shortcuts for the snippets, like `snippet aFunc bFunc` (at least one shortcut is required)
* every line of code within the snippet must start with a tab (`\t`)
* a snippet can have multiple placeholders which are indicated by `${num[:name]}` where num is a numeric value. Placeholders with the same number share the same value. You also have the option to give a placeholder a name / description / default value.

Plugins can provide snippet files they just need to publish them as a runtime file with type `snippets`.
See the plugins help for additional details.

Sample for go:

```
# creates a function that prints a value
snippet fmtfunc
	func ${0:name}() {
		fmt.Println("${0} prints:", ${1:value})
	}
```