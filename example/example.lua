---@meta

--- Position in 3D space
---@class Position
---@field X number
---@field Y number
---@field Z number
local Position = {}

--- Distance between two positions
---@param a Position
---@param b Position
---@return number
function Distance(a, b) end

--- Line in 3D space
---@class Line
local Line = {}

--- Length of the line
---@return number
function Line:Length() end

