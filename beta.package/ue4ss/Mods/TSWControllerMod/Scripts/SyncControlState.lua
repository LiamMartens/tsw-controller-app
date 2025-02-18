local Struct_SyncControlState = {}

function Struct_SyncControlState.New()
  local SyncControlState = {}
  SyncControlState.Components = {}

  function SyncControlState:AnyDirty()
    for _, control_state in pairs(self.Components) do
      if control_state.IsDirty then
        return true
      end
    end
    return false
  end

  ---@param control string
  ---@param value number
  function SyncControlState:SetCurrentInputValue(control, value)
    if self.Components[control] == nil then
      self.Components[control] = {}
    end
    self.Components[control].InputValue = value
    self.Components[control].IsDirty = true
  end

  return SyncControlState
end

return Struct_SyncControlState
