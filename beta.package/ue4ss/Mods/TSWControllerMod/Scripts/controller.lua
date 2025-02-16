local Helpers = require("Helpers")
local UEHelpers = require("UEHelpers")
local socket_conn = require("tsw5_gamepad_lua_socket_connection")

local ControlState = {}
ControlState.ThreadLocked = false
ControlState.VehicleID = nil
ControlState.IsDirty = false
ControlState.Components = {}

function ControlState:IsLocked()
  return self.ThreadLocked
end

function ControlState:ThreadLock()
  self.ThreadLocked = true
  return function()
    self.ThreadLocked = false
  end
end

-- run this action at 5fps
LoopAsync(200, function()
  -- do nothing if not dirty or is locked
  if not ControlState.IsDirty or ControlState:IsLocked() then
    return false
  end

  ExecuteInGameThread(function()
    local unlock = ControlState:ThreadLock()

    local player = UEHelpers.GetPlayer()
    local controller = player.Controller
    if not player:IsValid() or not controller:IsValid() then
      return unlock()
    end

    local drivable_actor = player:GetDrivableActor()
    if not drivable_actor:IsValid() then
      return unlock()
    end

    ControlState.IsDirty = false
    for control_name, control_state in pairs(ControlState.Components) do
      if control_state.IsDirty then
        control_state.IsDirty = false

        Helpers.InsertDirectControlPresetIfNotExists(drivable_actor, control_name)

        if drivable_actor[control_name]:IsValid() then
          -- the 1972 tube stock DeadmansHandleButton has a weird resetting logic when using VHID presets
          -- so we need to set the pushed state manually (SetPushedState doesn't work with other UPushButtonComponent's though oddly)
          if control_name == "DeadmansHandleButton" then
            drivable_actor[control_name]:SetPushedState(control_state.TargetValue > 0.5, true)
          else
            -- levers are controlled using VHID presets because it's more stable
            Helpers.InsertOrUpdateDirectControlPresetControlIfNotExists(drivable_actor, control_name, control_state.TargetValue)
          end

          drivable_actor.RailVehiclePhysicsComponent:ApplyVHIDPreset(
            drivable_actor.GameplayTasksComponent,
            controller,
            FName(string.format("DirectControl:%s", control_name)),
            control_state.TargetValue,       -- TargetInputValue
            0.0,       -- ErrorTolerance
            0.0,       -- MinMoveTime
            0.05,       -- MaxMoveTime
            100.0 -- RateOfChange
          )
        end
      end
    end

    return unlock()
  end)
  return false
end)

socket_conn.set_callback(function(var)
  print("Received message: " .. var)

  local command_split = Helpers.SplitString(var, ",")
  --- only respond to direct control commands
  if command_split[1] ~= "direct_control" then
    return
  end

  local command_control = command_split[2]
  local command_value = tonumber(command_split[3])

  local player = UEHelpers.GetPlayer()
  local controller = player.Controller
  if player:IsValid() and controller:IsValid() then
    local drivable_actor = player:GetDrivableActor()
    if drivable_actor:IsValid() then
      -- reset component state if vehicle changed
      if drivable_actor:GetAddress() ~= ControlState.VehicleID then
        ControlState.VehicleID = drivable_actor:GetAddress()
        ControlState.Components = {}
      end

      -- 0 = front, 1 = back
      local train_side = 0
      if player:GetAttachedSeatComponent().SeatSide == 1 then
        train_side = 1
      end
      local control_name = train_side == 0 and string.gsub(command_control, "{SIDE}", "F") or
          string.gsub(command_control, "{SIDE}", "B")
      if ControlState.Components[control_name] == nil then
        ControlState.Components[control_name] = {}
      end
      ControlState.Components[control_name].TargetValue = command_value
      ControlState.Components[control_name].IsDirty = true
      ControlState.IsDirty = true
    end
  end
end)

-- RegisterHook("/Script/TS2Prototype.VirtualHIDComponent:InputValueChanged", function(self, oldValue, newValue)
--   local vhid_component = self:get()
--   local vhid_component_identifier = vhid_component.InputIdentifier.Identifier:ToString()
--   print("InputValueChanged:" .. vhid_component_identifier .. ":" .. newValue.ToFloat .. "\n")
-- end)

-- RegisterHook("/Script/TS2Prototype.VirtualHIDComponent:OutputValueChanged", function(self, oldValue, newValue)
--   print("OutputValueChanged", newValue.ToFloat)
-- end)

-- RegisterHook("/Script/TS2Prototype.TS2PrototypePlayerController:EndUsingVHIDComponent", function(self, component)
--   print("EndUsingVHIDComponent")
-- end)

-- RegisterHook("/Script/TS2Prototype.TS2PrototypePlayerController:BeginDraggingVHIDComponent", function(self, component)
--   print("BeginDraggingVHIDComponent")
-- end)

-- RegisterHook("/Script/TS2Prototype.TS2PrototypePlayerController:ToggleVHIDComponent", function(self)
--   print("ToggleVHIDComponent")
-- end)

-- RegisterHook("/Script/TS2Prototype.TS2PrototypePlayerController:NotifyBeginInteraction", function(self)
--   print("NotifyBeginInteraction")
-- end)

-- RegisterHook("/Script/TS2Prototype.TS2PrototypePlayerController:NotifyEndInteraction", function(self)
--   print("NotifyEndInteraction")
-- end)

-- RegisterHook("/Script/TS2Prototype.TS2PrototypePlayerController:SetDragDeltaVHIDComponent", function(self, component, newValue)
--   print("SetDragDeltaVHIDComponent", newValue.ToFloat)
-- end)

-- RegisterHook("/Script/TS2Prototype.TS2PrototypePlayerController:BeginIncreasingVHIDComponent", function(self)
--   print("BeginIncreasingVHIDComponent")
-- end)

-- RegisterHook("/Script/TS2Prototype.TS2PrototypePlayerController:BeginDecreasingVHIDComponent", function(self)
--   print("BeginDecreasingVHIDComponent")
-- end)

-- RegisterHook("/Script/TS2Prototype.VirtualHIDComponent:BeginIncreaseDigital", function(self)
--   print("BeginIncreaseDigital")
-- end)
