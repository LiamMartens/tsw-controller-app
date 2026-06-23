package profile_runner

func (pc *ProfileRunnerAssignmentCall) IsSameAction(other *ProfileRunnerAssignmentCall) bool {
	if pc.ActionSequencerAction != nil && other.ActionSequencerAction != nil {
		return pc.ActionSequencerAction.Keys == other.ActionSequencerAction.Keys
	}
	if pc.VirtualAction != nil && other.VirtualAction != nil {
		return pc.VirtualAction.Control == other.VirtualAction.Control && pc.VirtualAction.Value == other.VirtualAction.Value
	}
	if pc.DirectControlCommand != nil && other.DirectControlCommand != nil {
		return pc.DirectControlCommand.ToSocketMessage().ToString() == other.DirectControlCommand.ToSocketMessage().ToString()
	}
	if pc.ApiControlCommand != nil && other.ApiControlCommand != nil {
		return pc.ApiControlCommand.Controls == other.ApiControlCommand.Controls && pc.ApiControlCommand.InputValue == other.ApiControlCommand.InputValue
	}
	return false
}
