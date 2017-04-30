# utf8.lua
one-file pure-lua 5.1 regex library

This library _is_ the simple way to add utf8 support into your application.

Some examples from http://www.lua.org/manual/5.1/manual.html#5.4 :
```Lua
local utf8 = require "utf8"

local s = "hello world from Lua"
for w in string.gmatch(s, "%a+") do
    print(w)
end
--[[
hello
world
from
Lua
]]--

s = "Привет, мир, от Lua"
for w in utf8.gmatch(s, "[^%p%d%s%c]+") do
    print(w)
end
--[[
Привет
мир
от
Lua
]]--

local t = {}
s = "from=world, to=Lua"
for k, v in string.gmatch(s, "(%w+)=(%w+)") do
    t[k] = v
end
for k,v in pairs(t) do
    print(k,v)
end
--[[
to	Lua
from	world
]]--

t = {}
s = "从=世界, 到=Lua"
for k, v in utf8.gmatch(s, "([^%p%s%c]+)=([^%p%s%c]+)") do
    t[k] = v
end
for k,v in pairs(t) do
    print(k,v)
end
--[[
到	Lua
从	世界
]]--

local x = string.gsub("hello world", "(%w+)", "%1 %1")
print(x)
--hello hello world world

x = utf8.gsub("Ahoj světe", "([^%p%s%c]+)", "%1 %1")
print(x)
--Ahoj Ahoj světe světe

x = string.gsub("hello world", "%w+", "%0 %0", 1)
print(x)
--hello hello world

x = utf8.gsub("Ahoj světe", "[^%p%s%c]+", "%0 %0", 1)
print(x)
--Ahoj Ahoj světe

x = string.gsub("hello world from Lua", "(%w+)%s*(%w+)", "%2 %1")
print(x)
--world hello Lua from

x = utf8.gsub("γεια κόσμο από Lua", "([^%p%s%c]+)%s*([^%p%s%c]+)", "%2 %1")
print(x)
--κόσμο γεια Lua από
```
Notice, there are some classes that can work only with latin(ASCII) symbols,
for details see: https://github.com/Stepets/utf8.lua/blob/master/utf8.lua#L470

Of course you can do this trick:
```Lua
for k,v in pairs(utf8) do
        string[k] = v
end
```
But this can lead to very strange errors. You were warned.

A little bit more interesting examples:
```Lua
local utf8 = require 'utf8'
for k,v in pairs(utf8) do
        string[k] = v
end

local str = "пыщпыщ ололоо я водитель нло"
print(str:find("(.л.+)н"))
-- 8	26	ололоо я водитель 

print(str:gsub("ло+", "보라"))
-- пыщпыщ о보라보라 я водитель н보라	3

print(str:match("^п[лопыщ ]*я"))
-- пыщпыщ ололоо я
```
