local section = "coverage"

-- check if string ends with another string
function string.endsWith(String,End)
   return End=='' or string.sub(String,-string.len(End))==End
end
-- check if string starts with another string
function string.startsWith(String,Start)
   return Start=='' or string.sub(String,1, string.len(Start))==Start
end

-- parses one line of coverprofile file
local function parseLine(line)
  return { 
    File = line:match("^(.+):%d+.%d+,%d+.%d+ %d+ %d+$"),
    Start  = {
      Line = tonumber(line:match("^.+:(%d+).%d+,%d+.%d+ %d+ %d+$")),
      Col = tonumber(line:match("^.+:%d+.(%d+),%d+.%d+ %d+ %d+$"))
    },
    End = {
      Line = tonumber(line:match("^.+:%d+.%d+,(%d+).%d+ %d+ %d+$")),
      Col = tonumber(line:match("^.+:%d+.%d+,%d+.(%d+) %d+ %d+$"))
    },
    NumStmt = tonumber(line:match("^.+:%d+.%d+,%d+.%d+ (%d+) %d+$")),
    Count = tonumber(line:match("^.+:%d+.%d+,%d+.%d+ %d+ (%d+)$"))
  }
end

-- read the coverprofile file
local function readProfile(filename)
  local result = nil
  for line in io.lines(filename) do
    if result == nil then
      result = { Mode = line:match("mode: (.+)"), Blocks = {} }      
    else 
      result.Blocks[#result.Blocks + 1] = parseLine(line)
    end
  end
  return result
end

-- show the coverprofile in the gutter
local function showProfile(file)
  local profile = readProfile(file)
  if profile == nil then
    messenger:Error("Failed to read profile file: "..file)
    return
  end

  local bufPath = CurView().Buf:AbsPath()
  local lines = {}

  for p, block in ipairs(profile.Blocks) do 
    if bufPath:endsWith(block.File) then
      for i=block.Start.Line, block.End.Line do 
        if lines[i] == nil then
          lines[i] = { total = 0, checked = 0 }
        end
        local l = lines[i]
        l.total = l.total + 1
        if block.Count > 0 then
          l.checked = l.checked + 1
        end
      end
    end
  end

  for i, line in pairs(lines) do
    if line.total == line.checked then
      CurView():GutterMessage(section, i, "Fully covered", 0)
    elseif line.checked == 0 then 
      CurView():GutterMessage(section, i, "Not covered", 2)
    else
      local perc = math.Floor((line.checked / line.total) * 10000) / 100
      CurView():GutterMessage(section, i, tostring(perc) .. "% covered", 1)
    end
  end
end

-- get the OS path seperator
local function getPathSep()
   local sep = "/"

    if OS == "windows" then
      sep = "\\"
    end
    return sep
end 

-- get the directory of a given file
local function getDirectory(file)
  local sep = getPathSep()
  local p = string.find(file, sep, 1)
  local lastIndex = p
  while p do
    p = string.find(file, sep, p + 1)
    if p then
      lastIndex = p
    end
  end
  return file:sub(1, lastIndex-1)
end

-- get the package name of the current file
local function getPackage()
    local gp = os.getenv("GOPATH")
    if gp == nil then
      return nil
    end
    local sep = getPathSep()

    if gp:endsWith(sep) then
      gp = gp.sub(-1)
    end

    gp = gp .. sep .. "src" .. sep

    local bufPath = getDirectory(CurView().Buf:AbsPath())
    if bufPath:startsWith(gp) then
      bufPath = bufPath:sub(gp:len()+1)
      return bufPath
    end
    return nil
end

-- gets called when go-test finshed
function onExitGoTest(output, tmpFile)
    showProfile(tmpFile)
    os.remove(tmpFile)
end

-- calls go-test as job
function Cover()
    CurView():ClearGutterMessages(section)
    local package = getPackage()
    if package == nil then
      messenger:Error("unable to determine go package name")
    else
      local tmpFile = os.tmpname()
      local cmd = "go test -coverprofile="..tmpFile.." "..package
      JobStart(cmd, "", "", "gocover.onExitGoTest", tmpFile)
    end
end

MakeCommand("gocover", "gocover.Cover")