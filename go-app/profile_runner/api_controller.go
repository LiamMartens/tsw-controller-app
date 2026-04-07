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

type ApiController_Command struct {
	Controls      string
	InputValue    float64
	Hold          bool
	MaxChangeRate float64
}

type ApiController_Interacting_Control struct {
	Cancel        context.CancelFunc
	Timer         *time.Timer
	TargetCommand *ApiController_Command
}

type ApiController_Interacting struct {
	mutex    sync.RWMutex
	controls map[string]ApiController_Interacting_Control
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
	interacting    ApiController_Interacting
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

func (controller *ApiController) CancelStopInteractingTimer(ctx context.Context, control string) {
	controller.interacting.mutex.Lock()
	defer controller.interacting.mutex.Unlock()
	if interacting, is_interacting := controller.interacting.controls[control]; is_interacting {
		/* already interacting; reset timer */
		interacting.Cancel()
	}
}

func (controller *ApiController) StartInteractingIfNotAlready(ctx context.Context, control string) error {
	controller.interacting.mutex.Lock()
	defer controller.interacting.mutex.Unlock()

	if interacting, is_interacting := controller.interacting.controls[control]; is_interacting {
		/* already interacting; reset timer */
		interacting.Timer.Reset(time.Second * 1)
		return nil
	}

	/* start interaction if not already */
	err := controller.API.SetInteracting(control, 1.0)
	if err != nil {
		logger.Logger.Error("could not start interacting", "control", control, "error", err)
		return err
	}

	logger.Logger.Debug("started interacting with", "control", control)
	childctx, childctxcancel := context.WithCancel(ctx)
	stop_interacting_timer := time.NewTimer(time.Second * 1)
	controller.interacting.controls[control] = ApiController_Interacting_Control{
		Cancel: childctxcancel,
		Timer:  stop_interacting_timer,
	}

	/* start go routine which will stop the interaction */
	go func() {
		defer stop_interacting_timer.Stop()
		select {
		case <-childctx.Done():
			return
		case <-stop_interacting_timer.C:
			if controller.interacting.controls[control].TargetCommand != nil {
				logger.Logger.Debug("skipping stop interaction", "control", control)
				stop_interacting_timer.Reset(time.Second * 1)
			} else if err := controller.API.SetInteracting(control, 0.0); err != nil {
				logger.Logger.Debug("could not stop interacting with", "control", control)
				stop_interacting_timer.Reset(time.Second * 1)
			} else {
				logger.Logger.Debug("stopped interacting with", "control", control)
				controller.interacting.mutex.Lock()
				delete(controller.interacting.controls, control)
				controller.interacting.mutex.Unlock()
			}
		}
	}()
	return nil
}

func (controller *ApiController) UpdateControlValue(ctx context.Context, control string, value float64) error {
	if err := controller.API.SetInputValue(control, value); err != nil {
		logger.Logger.Error("could not update value", "error", err)
		return err
	}

	return nil
}

func (controller *ApiController) getControlTargetCommand(control string) *ApiController_Command {
	controller.interacting.mutex.Lock()
	defer controller.interacting.mutex.Unlock()

	controlstate, has_controlstate := controller.interacting.controls[control]
	if !has_controlstate || controlstate.TargetCommand == nil {
		return nil
	}
	return controlstate.TargetCommand
}

func (controller *ApiController) clearControlTargetCommand(command ApiController_Command) {
	controller.interacting.mutex.Lock()
	defer controller.interacting.mutex.Unlock()

	controlstate, has_controlstate := controller.interacting.controls[command.Controls]
	/* clear target command if the same and not hold */
	if has_controlstate &&
		controlstate.TargetCommand != nil &&
		!controlstate.TargetCommand.Hold &&
		controlstate.TargetCommand.ToString() == command.ToString() {
		controlstate.TargetCommand = nil
		controller.interacting.controls[command.Controls] = controlstate
	}
}

func (controller *ApiController) ProcessPendingControlCommand(ctx context.Context, control string) error {
	command_ptr := controller.getControlTargetCommand(control)
	if command_ptr == nil {
		return nil
	}
	command := *command_ptr
	/* defer clear for this command */
	defer controller.clearControlTargetCommand(command)

	/* we're just silently ignoring this error here and starting from a default of 0.0f on failure which is acceptable in most cases */
	current_value := 0.0
	is_button, _ := controller.API.GetIsButton(command.Controls)
	current_value, _ = controller.API.GetInputValue(command.Controls)
	target_value_diff := math.Abs(current_value - command.InputValue)

	/* no-op if the value is already the same and it's not being "held" */
	if !command.Hold && math_utils.IsWithinMarginOfError(current_value, command.InputValue) {
		return nil
	}

	if is_button {
		/* buttons handle interacting separately */
		controller.CancelStopInteractingTimer(ctx, command.Controls)
		if command.InputValue > 0.5 {
			return controller.API.SetInteracting(command.Controls, 1.0)
		}
		return controller.API.SetInteracting(command.Controls, 0.0)
	} else if target_value_diff <= command.MaxChangeRate {
		/* if less than max change rate; change as-is */
		if err := controller.StartInteractingIfNotAlready(ctx, command.Controls); err != nil {
			return err
		}
		logger.Logger.Info("updating input value", "value", command.InputValue)
		return controller.UpdateControlValue(ctx, command.Controls, command.InputValue)
	} else {
		/* if not generate steps to reach the target value */
		num_steps := int(math.Ceil(target_value_diff / command.MaxChangeRate))
		for step := 1; step <= num_steps; step++ {
			if err := controller.StartInteractingIfNotAlready(ctx, command.Controls); err != nil {
				return err
			}
			set_value := current_value
			if current_value < command.InputValue {
				set_value = math.Min(current_value+(float64(step)*command.MaxChangeRate), command.InputValue)
			} else {
				set_value = math.Max(current_value-(float64(step)*command.MaxChangeRate), command.InputValue)
			}
			if err := controller.UpdateControlValue(ctx, command.Controls, set_value); err != nil {
				return err
			}
		}
	}

	return nil
}

/*
* Iterates over the current pending interacting states and processes any target commands
 */
func (controller *ApiController) ProcessPendingControlStates(ctx context.Context) error {
	controller.interacting.mutex.Lock()
	defer controller.interacting.mutex.Unlock()
	for control, controlstate := range controller.interacting.controls {
		if controlstate.TargetCommand != nil {
			go controller.ProcessPendingControlCommand(ctx, control)
		}
	}
	return nil
}

/*
* Processes the incoming control command:
* - Starts interacting with the control from an API pespective
* - Assigns incoming command as the target command
 */
func (controller *ApiController) ProcessControlCommand(ctx context.Context, command ApiController_Command) error {
	control, err := controller.formatControlName(command.Controls)

	if err != nil {
		return err
	}

	if err := controller.StartInteractingIfNotAlready(ctx, control); err != nil {
		return err
	}

	controller.interacting.mutex.Lock()
	defer controller.interacting.mutex.Unlock()
	controlstate := controller.interacting.controls[control]
	controlstate.TargetCommand = &ApiController_Command{
		Controls:      control,
		InputValue:    command.InputValue,
		Hold:          command.Hold,
		MaxChangeRate: command.MaxChangeRate,
	}
	controller.interacting.controls[control] = controlstate
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
		interacting: ApiController_Interacting{
			mutex:    sync.RWMutex{},
			controls: map[string]ApiController_Interacting_Control{},
		},
	}
	return &controller
}
