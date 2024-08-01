VERSION = '1.0.0'

local micro = import('micro')
local config = import('micro/config')
local shell = import('micro/shell')

local errors = import('errors')
local fmt = import('fmt')
local regexp = import('regexp')
local runtime = import('runtime')
local strings = import('strings')

---@alias Error { Error: fun(): string } # Goland error

---@class Buffer
---@field Path      string
---@field AbsPath   string
---@field FileType  fun(): string
---@field ReOpen    fun()

---@class BufPane userdata  # Micro BufPane
---@field Buf Buffer

-- luacheck: globals toString
---@param value any
local function toString(value)
  if type(value) == 'string' then
    return value
  elseif type(value) == 'table' then
    return strings.Join(value, ' ')
  end
  return value
end

-- luacheck: globals contains
---@param t table # Table to check.
---@param e any # Element to verify.
---@return boolean # If contains or not
local function contains(t, e)
  for i = 1, #t do
    if t[i] == e then
      return true
    end
  end
  return false
end

-- luacheck: globals Format
---@class Format
---@field cmd        string
---@field args       string|string[]
---@field name       string
---@field bind       string
---@field onSave     boolean
---@field filetypes  string[]
---@field os         string[]
---@field whitelist  boolean
---@field domatch    boolean
---@field callback   fun(buf: Buffer): boolean
local Format = {}

---create a valid formatter
---@param input table
---@return Format?, Error?
---@nodiscard
function Format:new(input)
  ---@type Format
  local f = {}

  if input.cmd == nil or type(input.cmd) ~= 'string' then
    return input, errors.New('Invalid "cmd"')
  elseif input.filetypes == nil or type(input.filetypes) ~= 'table' then
    return input, errors.New('Invalid "filetypes"')
  end

  f.cmd = input.cmd
  f.filetypes = input.filetypes

  if not input.name then
    ---@type string[]
    local cmds = strings.Split(input.cmd, ' ')
    f.name = fmt.Sprintf('%s', cmds[1])
  else
    f.name = input.name
  end

  f.bind = input.bind
  f.args = toString(input.args) or ''
  f.onSave = input.onSave
  f.os = input.os
  f.whitelist = input.whitelist or false
  f.domatch = input.domatch or false
  f.callback = input.callback

  self.__index = self
  return setmetatable(f, self), nil
end

---@return boolean
function Format:hasOS()
  if self.os == nil then
    return true
  end
  local has_os = contains(self.os, runtime.GOOS)
  if (not has_os and self.whitelist) or (has_os and not self.whitelist) then
    return false
  end

  return true
end

---@param buf Buffer
---@param filter fun(f: Format): boolean
---@return boolean
function Format:hasFormat(buf, filter)
  if filter ~= nil and not filter(self) then
    return false
  end

  ---@type string
  local filetype = buf:FileType()
  ---@type string[]
  local filetypes = self.filetypes

  for _, ft in ipairs(filetypes) do
    if self.domatch then
      if regexp.MatchString(ft, buf.AbsPath) then
        return true
      end
    elseif ft == filetype then
      return true
    end
  end
  return false
end

---@param buf Buffer
---@return boolean
function Format:hasCallback(buf)
  if self.callback ~= nil and type(self.callback) == 'function' and not self.callback(buf) then
    return false
  end
  return true
end

---run a formatter on a given file
---@param buf Buffer
---@return Error?
function Format:run(buf)
  ---@type string
  local args = self.args:gsub('%%f', buf.Path)
  ---@type string
  local cmd = fmt.Sprintf('%s %s', self.cmd:gsub('%%f', buf.Path), args)
  -- err: Error?
  local _, err = shell.RunCommand(cmd)

  ---@type string
  if err ~= nil then
    return err
  end
end

---@type Format[]
-- luacheck: globals formatters
local formatters = {}

-- luacheck: globals format
---format a bufpane
---@param bp       BufPane
---@param args     strign[]
---@param filter?  fun(f: Format): boolean
---@return Error?
local function format(bp, args, filter)
  if #formatters < 1 then
    return
  end

  local name = nil
  if #args >= 1 then
    name = args[1]
  end

  ---@type string
  local errs = ''
  for _, f in ipairs(formatters) do
    ---@cast filter fun(f: Format): boolean
    if (name == nil or name == f.name) and f:hasFormat(bp.Buf, filter) and f:hasOS() and f:hasCallback(bp.Buf) then
      local err = f:run(bp.Buf)
      if err ~= nil then
        errs = fmt.Sprintf('%s | %s', errs, f.name)
      end
    end
  end

  bp.Buf:ReOpen()

  if errs ~= '' then
    return micro.InfoBar():Error('💥 Error when using formatters: %s', errs)
  else
    micro.InfoBar():Message(fmt.Sprintf('🎬 File formatted successfully! %s ✨ 🍰 ✨', bp.Buf.Path))
  end
end

---@param buf Buffer
---@return (string[], string[])
local function formatComplete(buf)
  local completions, suggestions = {}, {}

  ---@type string
  local input = buf:GetArg()

  ---@type BufPane
  local bp = micro.CurPane()

  for _, f in ipairs(formatters) do
    -- i: integer
    -- j: integer
    local i, j = f.name:find(input, 1, true)
    if i == 1 and f:hasFormat(bp.Buf) and f:hasOS() and f:hasCallback(bp.Buf) then
      table.insert(suggestions, f.name)
      table.insert(completions, f.name:sub(j + 1))
    end
  end

  table.sort(completions)
  table.sort(suggestions)

  return completions, suggestions
end

-- luacheck: globals makeFormatter
---make a formatter
---@param cmd        string
---@param filetypes  string[]
---@param args       string|string[]
---@param name       string
---@param bind       string
---@param onSave     boolean
---@param os         string[]
---@param whitelist  boolean
---@param domatch    boolean
---@param callback   fun(buf: Buffer): boolean
---@return Error?
function makeFormatter(cmd, filetypes, args, name, bind, onSave, os, whitelist, domatch, callback)
  -- f: Format
  -- err: Error?
  local f, err = Format:new({
    cmd = cmd,
    filetypes = filetypes,
    args = args,
    name = name,
    bind = bind,
    onSave = onSave,
    os = os,
    whitelist = whitelist,
    domatch = domatch,
    callback = callback,
  })
  if err ~= nil then
    return err
  end
  table.insert(formatters, f)

  if f.bind then
    config.TryBindKey(f.bind, 'command:format ' .. f.name, true)
  end
end

-- luacheck: globals setup
---initialize formatters
---@param formats Format[]
function setup(formats)
  ---@type string
  for _, f in ipairs(formats) do
    ---@type Error?
    makeFormatter(f.cmd, f.filetypes, f.args, f.name, f.bind, f.onSave, f.os, f.whitelist, f.domatch, f.callback)
  end
end

-- CALLBACK'S

---runs formatters set to onSave
---@param bp BufPane
function onSave(bp)
  if #formatters < 1 then
    return true
  end

  ---@type Error?
  local err = format(bp, {}, function(f)
    return f.onSave == true
  end)

  if err ~= nil then
    micro.InfoBar():Error(fmt.Sprintf('%v', err))
  else
    micro.InfoBar():Message(fmt.Sprintf('🎬 Saved! %s ✨ 🍰 ✨', bp.Buf.Path))
  end
  return true
end

function init()
  config.AddRuntimeFile('formatter', config.RTHelp, 'help/formatter.md')
  config.MakeCommand('format', format, formatComplete)
  config.TryBindKey('Alt-f', 'command:format', false)
end
