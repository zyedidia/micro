# Status

The status plugin provides some functions for modifying the status line.

Using the `statusformatl` and `statusformatr` options, the exact contents
of the status line can be modified. Please see the documentation for
those options (`> help options`) for more information.

This plugin provides the three functions that can be used in the status
line format:

* `status.branch`: returns the name of the current git branch.
* `status.hash`: returns the hash of the current git commit.
* `status.paste`: returns "" if the paste option is disabled and "PASTE"
   if it is enabled.
