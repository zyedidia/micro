
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

local curFileType = ""
local snippets = {}

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
			curSnip = { code = "" }
			for snipName in line:gmatch("%s(%a+)") do
				snippets[snipName] = curSnip
			end
		else
			curSnip.code = curSnip.code .. "\n" .. line:match("^\t(.*)$")
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
	EnsureSnippets()
	local curSn = snippets[CursorWord()]
	if curSn then
		messenger:Message(curSn.code)
	end
end
MakeCommand("foo", "snippet.foo", 0)