local Struct_ControlState = {}

function Struct_ControlState.New()
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

  return ControlState
end

return Struct_ControlState
