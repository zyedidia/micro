---@alias BufPane userdata

VERSION = "1.0.0"

local config = import("micro/config")

---@param bp BufPane
---@return boolean
function preInsertNewline(bp)
   local buf = bp.Buf

   ---@type string
   local regex = buf.Settings["indent.regex"]
   if not regex then return true end

   ---@type string
   local separators = buf.Settings["indent.separator"]

   ---@type string
   local line = buf:Line(bp.Cursor.Y)

   if separators then
      for pattern in regex:gmatch("[^" .. separators .. "]+") do
         if line:match(pattern) then
            bp:InsertNewline()
            bp:InsertTab()
            return false
         end
      end
   else
      if line:match(regex) then
         bp:InsertNewline()
         bp:InsertTab()
         return false
      end
   end

   return true
end

function preinit()
   config.RegisterCommonOption("indent", "patterns", "")
   config.RegisterCommonOption("indent", "separator", "|")

   config.AddRuntimeFile("indent", config.RTHelp, "help/indent.md")
end
