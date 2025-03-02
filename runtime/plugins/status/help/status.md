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
