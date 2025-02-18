local Helpers = {}

function Helpers.SplitString(str, sep)
  local result = {}
  for part in str:gmatch("([^" .. sep .. "]+)") do
    table.insert(result, part)
  end
  return result
end

---@param vehicle ARailVehicle
---@param control_name string
function Helpers.InsertDirectControlPresetIfNotExists(vehicle, control_name)
  local has_direct_control_preset = false
  local preset_name = string.format("DirectControl:%s", control_name)
  local presets = vehicle.RailVehiclePhysicsComponent.VHIDPresets
  -- this is obviously not ideal to loop over presets twice but it's the only way to properly save into it
  presets:ForEach(function(index, remote)
    local element = remote:get()
    if element.PresetName:ToString() == preset_name then
      has_direct_control_preset = true
    end
  end)
  if not has_direct_control_preset then
    local direct_control_preset = {}
    direct_control_preset.PresetName = FName(preset_name)
    direct_control_preset.DisplayName = FText(preset_name)
    direct_control_preset.Presets = {}
    table.insert(presets, direct_control_preset)
  end
end

---@param vehicle ARailVehicle
---@param control string
---@param target_value float
function Helpers.InsertOrUpdateDirectControlPresetControlIfNotExists(vehicle, control, target_value)
  local preset_name = string.format("DirectControl:%s", control)
  local presets = vehicle.RailVehiclePhysicsComponent.VHIDPresets
  -- this is obviously not ideal to loop over presets twice but it's the only way to properly save into it
  presets:ForEach(function(index, remote)
    local element = remote:get()
    if element.PresetName:ToString() == preset_name then
      local has_control_preset = false
      element.Presets:ForEach(function(index, preset)
        if preset:get().Component.ComponentName:ToString() == control then
          preset:get().TargetInputValue = target_value
          has_control_preset = true
        end
      end)

      if not has_control_preset then
        local new_control_preset = {}
        table.insert(element.Presets, new_control_preset)
        local new_control_preset = remote:get().Presets[remote:get().Presets:GetArrayNum()]
        local component_ref = {}
        component_ref.ComponentName = FName(control)
        new_control_preset.TargetInputValue = target_value
        new_control_preset.Component = component_ref
      end
    end
  end)
end

return Helpers
