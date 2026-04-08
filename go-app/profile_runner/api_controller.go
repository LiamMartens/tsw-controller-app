package profile_runner

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strings"
	"sync"
	"time"
	"tsw_controller_app/logger"
	"tsw_controller_app/math_utils"
	"tsw_controller_app/tswapi"
)

const API_CONTROLLER_QUEUE_BUFFER_SIZE = 32

type FormattedControlName = string

type ApiController_Command struct {
	Controls      string
	InputValue    float64
	Hold          bool
	MaxChangeRate float64
}

type ApiController_ControlStates_Control_InteractingState struct {
	InteractingSince *time.Time
}

type ApiController_ControlStates_Control struct {
	Mutex            sync.RWMutex
	ControlName      FormattedControlName
	InteractingState *ApiController_ControlStates_Control_InteractingState
	TargetCommand    *ApiController_Command
}

type ApiController_ControlStates struct {
	mutex    sync.RWMutex
	controls map[FormattedControlName]*ApiController_ControlStates_Control
}

type ApiController_ControlTargets struct {
	mutex sync.RWMutex
}

type ApiController_ActiveCab struct {
	updatedAt *time.Time
	Front     bool
	Back      bool
}

type ApiController struct {
	API            tswapi.ITSWAPI
	ControlChannel chan ApiController_Command
	ActiveCab      ApiController_ActiveCab
	controlStates  ApiController_ControlStates
}

func (c *ApiController_Command) ToString() string {
	return fmt.Sprintf("api_control_command:%s:%f", c.Controls, c.InputValue)
}

func (s *ApiController_ActiveCab) Update(api tswapi.ITSWAPI) error {
	/* update at most once every 2 seconds */
	now := time.Now()
	should_update := s.updatedAt == nil || now.Sub(*s.updatedAt).Seconds() > 2.0
	if should_update {
		cab, err := api.GetActiveCab()
		s.Front = cab.Front
		s.Back = cab.Back
		s.updatedAt = &now
		return err
	}
	return nil
}

func (c *ApiController_ControlStates_Control) GetTargetCommand() (*ApiController_Command, bool) {
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()
	return c.TargetCommand, c.TargetCommand != nil
}

func (c *ApiController_ControlStates_Control) ClearTargetCommand() {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	c.TargetCommand = nil
}

func (c *ApiController_ControlStates_Control) UpdateTargetCommand(tc ApiController_Command) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	c.TargetCommand = &tc
}

func (c *ApiController_ControlStates_Control) StartInteractingIfNotAlready(api tswapi.ITSWAPI) error {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	if c.InteractingState.InteractingSince == nil {
		logger.Logger.Debug("starting interaction with", "control", c.ControlName)
		if err := api.SetInteracting(c.ControlName, 1.0); err != nil {
			return err
		}
		now := time.Now()
		c.InteractingState.InteractingSince = &now
	}
	return nil
}

func (c *ApiController_ControlStates_Control) StopInteractingIfNotAlready(api tswapi.ITSWAPI) error {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	if c.InteractingState.InteractingSince != nil {
		logger.Logger.Debug("stopping interaction with", "control", c.ControlName)
		if err := api.SetInteracting(c.ControlName, 0.0); err != nil {
			return err
		}
		c.InteractingState.InteractingSince = nil
	}
	return nil
}

func (c *ApiController_ControlStates_Control) SetInputValue(api tswapi.ITSWAPI, value float64) error {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	if err := api.SetInputValue(c.ControlName, value); err != nil {
		return err
	}
	return nil
}

func (c *ApiController) getControlStateRef(control FormattedControlName) *ApiController_ControlStates_Control {
	c.controlStates.mutex.Lock()
	defer c.controlStates.mutex.Unlock()
	if _, has_controlstate := c.controlStates.controls[control]; !has_controlstate {
		c.controlStates.controls[control] = &ApiController_ControlStates_Control{
			ControlName: control,
			InteractingState: &ApiController_ControlStates_Control_InteractingState{
				InteractingSince: nil,
			},
			TargetCommand: nil,
		}
	}
	return c.controlStates.controls[control]
}

func (controller *ApiController) formatControlName(control string) (string, error) {
	re := regexp.MustCompile(`\{SIDE\}|\{SIDE:[^:\}]+:[^:\}]+\}`)

	side_replacement_failed := false
	formatted_control := re.ReplaceAllStringFunc(control, func(match string) string {
		err := controller.ActiveCab.Update(controller.API)
		if err != nil {
			logger.Logger.Error("[App::ApiController] Could not update active cab: %e", err)
		}

		match_parts := strings.Split(match[1:len(match)-1], ":") /* split and remove leading and trailing {} */
		front_value := "F"
		back_value := "B"
		if len(match_parts) == 3 {
			front_value = match_parts[1]
			back_value = match_parts[2]
		}
		if controller.ActiveCab.Back {
			return back_value
		}
		if controller.ActiveCab.Front {
			return front_value
		}

		side_replacement_failed = true
		return match
	})

	if side_replacement_failed {
		return formatted_control, fmt.Errorf("Could not replace side placeholder due to missing active cab")
	}

	return formatted_control, nil
}

func (controller *ApiController) ProcessPendingControlState(ctx context.Context, control FormattedControlName) error {
	controlstate := controller.getControlStateRef(control)
	targetcmd, has_targetcmd := controlstate.GetTargetCommand()
	if !has_targetcmd {
		return nil
	}

	/* we're just silently ignoring this error here and starting from a default of 0.0f on failure which is acceptable in most cases */
	current_value := 0.0
	is_button, _ := controller.API.GetIsButton(control)
	current_value, _ = controller.API.GetInputValue(control)
	target_value_diff := math.Abs(current_value - targetcmd.InputValue)

	if !targetcmd.Hold && math_utils.IsWithinMarginOfError(current_value, targetcmd.InputValue) {
		controlstate.ClearTargetCommand()
		if !is_button {
			controlstate.StopInteractingIfNotAlready(controller.API)
		}
		return nil
	}

	if is_button {
		/**
		* buttons handle the interacting state differently than lever-like controls; so we handle them separately
		 */
		if targetcmd.InputValue > 0.5 {
			if err := controlstate.StartInteractingIfNotAlready(controller.API); err != nil {
				return err
			}
			if err := controlstate.SetInputValue(controller.API, 1.0); err != nil {
				return err
			}
		} else {
			if err := controlstate.SetInputValue(controller.API, 0.0); err != nil {
				return err
			}
			if err := controlstate.StopInteractingIfNotAlready(controller.API); err != nil {
				return err
			}
		}
	} else {
		if err := controlstate.StartInteractingIfNotAlready(controller.API); err != nil {
			return err
		}
		if target_value_diff <= targetcmd.MaxChangeRate {
			/* if less than max change rate; change as-is */
			return controlstate.SetInputValue(controller.API, targetcmd.InputValue)
		} else {
			/* if not generate steps to reach the target value */
			num_steps := int(math.Ceil(target_value_diff / targetcmd.MaxChangeRate))
			for step := 1; step <= num_steps; step++ {
				set_value := current_value
				if current_value < targetcmd.InputValue {
					set_value = math.Min(current_value+(float64(step)*targetcmd.MaxChangeRate), targetcmd.InputValue)
				} else {
					set_value = math.Max(current_value-(float64(step)*targetcmd.MaxChangeRate), targetcmd.InputValue)
				}
				if err := controlstate.SetInputValue(controller.API, set_value); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

/*
* Iterates over the current pending interacting states and processes any target commands
 */
func (controller *ApiController) ProcessPendingControlStates(ctx context.Context) error {
	controller.controlStates.mutex.RLock()
	defer controller.controlStates.mutex.RUnlock()
	for control, controlstate := range controller.controlStates.controls {
		if _, has_target_command := controlstate.GetTargetCommand(); has_target_command {
			/*
				create a copy of the control name and then defer spawning the go routine
				to early release the controlstates lock itself
			*/
			controlname := fmt.Sprintf("%s", control)
			go controller.ProcessPendingControlState(ctx, controlname)
		}
	}
	return nil
}

/*
* Processes the incoming control command:
* - Assigns incoming command as the target command for further processing
 */
func (controller *ApiController) ProcessControlCommand(ctx context.Context, command ApiController_Command) error {
	control, err := controller.formatControlName(command.Controls)

	if err != nil {
		return err
	}

	controller.getControlStateRef(control).UpdateTargetCommand(ApiController_Command{
		Controls:      control,
		InputValue:    command.InputValue,
		Hold:          command.Hold,
		MaxChangeRate: command.MaxChangeRate,
	})

	return nil
}

func (controller *ApiController) Run(ctx context.Context) func() {
	ctx_with_cancel, cancel := context.WithCancel(ctx)
	ticker := time.NewTicker(1000 / 15 * time.Millisecond)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx_with_cancel.Done():
				return
			case <-ticker.C:
				controller.ProcessPendingControlStates(ctx)
			case command := <-controller.ControlChannel:
				controller.ProcessControlCommand(ctx, command)
			}
		}
	}()

	return cancel
}

func NewAPIController(twapi tswapi.ITSWAPI) *ApiController {
	controller := ApiController{
		API:            twapi,
		ControlChannel: make(chan ApiController_Command, API_CONTROLLER_QUEUE_BUFFER_SIZE),
		ActiveCab: ApiController_ActiveCab{
			updatedAt: nil,
			Front:     false,
			Back:      false,
		},
		controlStates: ApiController_ControlStates{
			mutex:    sync.RWMutex{},
			controls: map[FormattedControlName]*ApiController_ControlStates_Control{},
		},
	}
	return &controller
}
