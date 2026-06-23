package controller_mgr

type ControllerManager_Control_ChangeEvent struct {
	Device       IControllerManager_Device
	Controller   IControllerManager_Controller
	Control      IControllerManager_Controller_Control
	ControlName  string
	ControlState ControllerManager_Controller_ControlState
}
