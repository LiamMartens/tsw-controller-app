package main

import "tsw_controller_app/cabdebugger"

func (a *App) GetCabControlState() (Interop_Cab_ControlState, error) {
	control_state := Interop_Cab_ControlState{
		Name:     a.cab_debugger.State.DrivableActorName,
		Controls: []Interop_Cab_ControlState_Control{},
	}

	a.cab_debugger.State.Controls.ForEach(func(control cabdebugger.CabDebugger_ControlState_Control, key cabdebugger.PropertyName) bool {
		control_state.Controls = append(control_state.Controls, Interop_Cab_ControlState_Control{
			Identifier:             control.Identifier,
			PropertyName:           control.PropertyName,
			CurrentValue:           control.CurrentValue,
			CurrentNormalizedValue: control.CurrentNormalizedValue,
		})
		return true
	})

	return control_state, nil
}

func (a *App) ResetCabControlState() {
	a.cab_debugger.Clear()
}
