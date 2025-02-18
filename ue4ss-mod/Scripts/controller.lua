local Helpers = require("Helpers")
local UEHelpers = require("UEHelpers")
local DirectControlStateClass = require("DirectControlState")
local SyncControlStateClass = require("SyncControlState")
local socket_conn = require("tsw5_gamepad_lua_socket_connection")

local DirectControlState = DirectControlStateClass.New()
local SyncControlState = SyncControlStateClass.New()

-- set socket connection callback
socket_conn.direct_controller_task.set_callback(function(var)
  print("[TSW5GamepadMod] Received message: " .. var .. "\n")

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
      if drivable_actor:GetAddress() ~= DirectControlState.VehicleID then
        DirectControlState:Reset(drivable_actor:GetAddress())
      end

      -- 0 = front, 1 = back
      local train_side = 0
      if player:GetAttachedSeatComponent().SeatSide == 1 then
        train_side = 1
      end
      local control_name = train_side == 0 and string.gsub(command_control, "{SIDE}", "F") or
          string.gsub(command_control, "{SIDE}", "B")
      DirectControlState:SetComponentTargetValue(control_name, command_value)
    end
  end
end)

-- this loop handles applying the dirty control states to the VHID
LoopAsync(100, function()
  -- do nothing if not dirty or is locked
  if not DirectControlState:AnyDirty() or DirectControlState:IsLocked() then
    return false
  end

  ExecuteInGameThread(function()
    print("[TSW5GamepadMod] Running control state update\n")

    local unlock = DirectControlState:ThreadLock()

    local player = UEHelpers.GetPlayer()
    local controller = player.Controller
    if not player:IsValid() or not controller:IsValid() then
      return unlock()
    end

    local drivable_actor = player:GetDrivableActor()
    if not drivable_actor:IsValid() then
      return unlock()
    end

    print("[TSW5GamepadMod] Checking components\n")
    for control_name, control_state in pairs(DirectControlState.Components) do
      if control_state.IsDirty then
        control_state.IsDirty = false

        print("[TSW5GamepadMod] Control name valid check (" .. control_name .. ")\n")
        if drivable_actor[control_name]:IsValid() then
          print("[TSW5GamepadMod] Insert preset if not exists\n")
          Helpers.InsertDirectControlPresetIfNotExists(drivable_actor, control_name)
          print("[TSW5GamepadMod] Update preset target value\n")
          -- the 1972 tube stock DeadmansHandleButton has a weird resetting logic when using VHID presets
          -- so we need to set the pushed state manually (SetPushedState doesn't work with other UPushButtonComponent's though oddly)
          if control_name == "DeadmansHandleButton" then
            drivable_actor[control_name]:SetPushedState(control_state.TargetValue > 0.5, true)
          else
            -- levers are controlled using VHID presets because it's more stable
            Helpers.InsertOrUpdateDirectControlPresetControlIfNotExists(drivable_actor, control_name,
              control_state.TargetValue)
          end

          local preset_name = string.format("DirectControl:%s", control_name)
          print("[TSW5GamepadMod] Apply VHID Preset (" .. preset_name .. ")\n")
          -- Begin interacting?
          drivable_actor.RailVehiclePhysicsComponent:ApplyVHIDPreset(
            drivable_actor.GameplayTasksComponent,
            controller,
            FName(preset_name),
            control_state.TargetValue, -- TargetInputValue
            0.05,                      -- ErrorTolerance
            0.05,                      -- MinMoveTime
            0.05,                      -- MaxMoveTime
            100.0                       -- RateOfChange
          )
          print("[TSW5GamepadMod] Applied VHID Preset (" .. preset_name .. ")\n")
        end
      end
    end

    print("[TSW5GamepadMod] Unlocking thread logic\n")
    return unlock()
  end)
  return false
end)

-- this loop handles sending the current input values to the SC WS server
LoopAsync(100, function()
  if SyncControlState:AnyDirty() then
    for vhid_component_identifier, control_state in pairs(SyncControlState.Components) do
      control_state.IsDirty = false
      local sync_state_message = string.format("%s,%.3f", vhid_component_identifier, control_state.InputValue)
      print("[SC] Forwarding message: " .. sync_state_message .. " \n")
      socket_conn.sync_controller_task.send(sync_state_message)
    end
  end
  return false
end)

-- attach handlers
RegisterHook("/Script/TS2Prototype.VirtualHIDComponent:InputValueChanged", function(self, oldValue, newValue)
  local vhid_component = self:get()
  local changing_controller = vhid_component:GetCurrentlyChangingController()
  local vhid_component_identifier = vhid_component.InputIdentifier.Identifier:ToString()
  -- ignore any None components or controls that aren't being controlled by the current player
  if vhid_component_identifier ~= "None" and changing_controller:GetAddress() == UEHelpers.GetPlayerController():GetAddress() then
    print("[SC] InputValueChanged: " .. vhid_component_identifier .. " " .. newValue.ToFloat .. "\n")
    SyncControlState:SetCurrentInputValue(vhid_component_identifier, newValue.ToFloat)
  end
end)
