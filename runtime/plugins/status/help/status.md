# Status

The status plugin provides some functions for modifying the status line.

Using the `statusformatl` and `statusformatr` options, the exact contents
of the status line can be modified. Please see the documentation for
those options (`> help options`) for more information.

This plugin provides functions that can be used in the status line format:

* `status.branch`: returns the name of the current git branch in the repository
   where the file is located.
* `status.hash`: returns the hash of the current git commit in the repository
   where the file is located.
* `status.paste`: returns "" if the paste option is disabled and "PASTE"
   if it is enabled.
* `status.lines`: returns the number of lines in the buffer.
* `status.vcol`: returns the visual column number of the cursor.
* `status.bytes`: returns the number of bytes in the current buffer.
* `status.size`: returns the size of the current buffer in a human-readable
   format.
* `status.icon`: returns a Nerd Font icon representing the current file type
  of the buffer. ⚠️ **Requires a Nerd Font installed in your terminal** to display correctly.

### Overriding default icons

The icons used by `status.icon` can be customized in your
`~/.config/micro/settings.json` file. You can override the default icon
for any filetype by adding an entry in the `status.icons` option.
For example:

```json
{
  "status.icons": "go=,lua=,typescript=,ruby=,unkwown="
}
```
