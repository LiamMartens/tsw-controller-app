local Struct_DirectControlState = {}

function Struct_DirectControlState.New()
  local DirectControlState = {}
  DirectControlState.ThreadLocked = false
  DirectControlState.VehicleID = nil
  DirectControlState.Components = {}

  function DirectControlState:IsLocked()
    return self.ThreadLocked
  end

  function DirectControlState:ThreadLock()
    self.ThreadLocked = true
    return function()
      self.ThreadLocked = false
    end
  end

  function DirectControlState:AnyDirty()
    for _, control_state in pairs(self.Components) do
      if control_state.IsDirty then
        return true
      end
    end
    return false
  end

  ---@param vehicle_id number
  function DirectControlState:Reset(vehicle_id)
    self.VehicleID = nil
    self.Components = {}
  end

  ---@param control_name string
  ---@param target_value number
  function DirectControlState:SetComponentTargetValue(control_name, target_value)
    if self.Components[control_name] == nil then
      self.Components[control_name] = {}
    end
    self.Components[control_name].TargetValue = target_value
    self.Components[control_name].IsDirty = true
  end

  return DirectControlState
end

return Struct_DirectControlState
