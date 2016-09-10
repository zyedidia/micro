local curFileType = ""
local snippets = {}

local Snippet = {}
Snippet.__index = Snippet

function Snippet.new()
	local self = setmetatable({}, Snippet)
	self.code = ""
    return self
end

function Snippet.AddCodeLine(self, line)
	if self.code ~= "" then
		self.code = self.code .. "\n"
	end
	self.code = self.code .. line
end

function Snippet.Prepare(self)
	if not self.placeholders then
		self.placeholders = {}
		self.locations = {}
		local pattern = "${(%d+):?([^}]*)}"
		while true do
  			local num, value = self.code:match(pattern)
  			if not num then
    			break
  			end
  			local idx = tonumber(self.code:find(pattern))
  			self.code = self.code:gsub(pattern, "", 1)

  			local p = self.placeholders[num]
  			if not p then
  				p = {}
  				self.placeholders[num] = p
  			end
  			self.locations[#self.locations+1] = {
	  			idx = idx,
	  			ph = p
	  		}

  			if value then
  				p.value = value
  			end
		end
	end
end

function Snippet.newInstance(self)
	self:Prepare()
	-- todo
	return self
end

function Snippet.str(self)
	local res = self.code
	for i = #self.locations, 1, -1 do
		local loc = self.locations[i]
		res = res:sub(0, loc.idx-1) .. loc.ph.value .. res:sub(loc.idx)
	end
	return res
end

local function CursorWord()
	local c = CurView().Cursor
	local x = c.X-1 -- start one rune before the cursor
	local result = ""
	while x >= 0 do
		local r = RuneStr(c:RuneUnder(x))
		if IsWordChar(r) then
			result = r .. result
		else
			break
		end
		x = x-1
	end

	return result
end

local function ReadSnippets(filetype)
	local snippets = {}
	local filename = JoinPaths(configDir, "plugins", "snippet", filetype .. ".snippet")

	-- first test if the file exists
	local f = io.open(filename, "r")
	if f then
		f:close()
	else
		return snippets
	end

	
	local curSnip = nil

	for line in io.lines(filename) do
		if string.match(line,"^#") then
			-- comment
		elseif line:match("^snippet") then
			curSnip = Snippet.new()
			for snipName in line:gmatch("%s(%a+)") do
				snippets[snipName] = curSnip
			end
		else
			curSnip:AddCodeLine(line:match("^\t(.*)$"))
		end
	end
	return snippets
end

local function EnsureSnippets()
	local filetype = GetOption("filetype")
	if curFileType ~= filetype then
		snippets = ReadSnippets(filetype)
		curFileType = filetype
	end
end

function foo()
	local v = CurView()
	local c = v.Cursor
	local xy = {X=c.X, Y=c.Y}
	local name = CursorWord()

	EnsureSnippets()
	local curSn = snippets[name]
	if curSn then
		curSn = curSn:newInstance()

		c:SetSelectionStart({X = xy.X - name:len(), Y = xy.Y})
		c:SetSelectionEnd(xy)
		c:DeleteSelection()
		c:ResetSelection()
		v.Buf:insert(xy, curSn:str())
	end
end

MakeCommand("foo", "snippet.foo", 0)