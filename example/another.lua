---@meta

--- Represents a location
---@class Loc
---@field X integer
---@field Y integer
local Loc = {}

--- Represents a cursor
---@class Cursor
---@field Loc Loc
local Cursor = {}

--- Moves the cursor up
function Cursor:Up() end

--- Moves the cursor down
function Cursor:Down() end

