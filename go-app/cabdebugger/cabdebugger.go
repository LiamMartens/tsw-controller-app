package cabdebugger

import (
	"context"
	"errors"
	"net"
	"strconv"
	"sync"
	"time"
	"tsw_controller_app/map_utils"
	"tsw_controller_app/tswapi"
	"tsw_controller_app/tswconnector"
)

type PropertyName = string

const CURRENT_DRIVABLE_ACTOR_CONNECTOR_EVENT_NAME = "current_drivable_actor"
const SYNC_CONTROL_VALUE_CONNECTOR_EVENT_NAME = "sync_control_value"

// http://localhost:31270/get/CurrentDrivableActor.Function.IS_GetActiveCab

type CabDebugger_ControlState_Control struct {
	Identifier             string
	PropertyName           PropertyName
	CurrentValue           float64
	CurrentNormalizedValue float64
}

type CabDebugger_ControlState struct {
	Mutex             sync.Mutex
	DrivableActorName string
	Controls          *map_utils.LockMap[PropertyName, CabDebugger_ControlState_Control]
}

type CabDebugger_Config struct {
	TSWAPISubscriptionIDStart int
}

type CabDebugger struct {
	Connector tswconnector.TSWConnector
	TSWAPI    *tswapi.TSWAPI
	Config    CabDebugger_Config
	State     CabDebugger_ControlState
}

var ErrAlreadyLocked = errors.New("already locked error")

func (cd *CabDebugger) updateCurrentDrivableActor(name string) {
	cd.State.Mutex.Lock()
	defer cd.State.Mutex.Unlock()
	should_reset := cd.State.DrivableActorName != name
	cd.State.DrivableActorName = name
	if should_reset {
		cd.State.Controls.Clear()
		cd.TSWAPI.DeleteSubscription(cd.Config.TSWAPISubscriptionIDStart)
	}
}

func (cd *CabDebugger) updateControlStateFromAPI() error {
	if cd.TSWAPI.Enabled() {
		/* try to acquire lock ; if already locked we skip */
		did_lock := cd.State.Mutex.TryLock()
		if !did_lock {
			return ErrAlreadyLocked
		}
		defer cd.State.Mutex.Unlock()

		drivable_actor_result, err := cd.TSWAPI.GetCurrentDrivableActorObjectClass()
		if err != nil {
			return nil
		}

		subscription_result, err := cd.TSWAPI.GetCurrentDrivableActorSubscription(cd.Config.TSWAPISubscriptionIDStart)
		if (err != nil &&
			/* don't do anything further if the comm api key is missing */
			!errors.Is(err, tswapi.ErrMissingCommAPIKey) &&
			/* don't do anything for an OpError */
			!errors.As(err, new(*net.OpError))) ||
			drivable_actor_result != cd.State.DrivableActorName {
			cd.State.DrivableActorName = drivable_actor_result
			cd.State.Controls.Clear()
			cd.TSWAPI.DeleteSubscription(cd.Config.TSWAPISubscriptionIDStart)
			if err := cd.TSWAPI.CreateCurrentDrivableActorSubscription(cd.Config.TSWAPISubscriptionIDStart); err != nil {
				return err
			}
			if subscription_result, err = cd.TSWAPI.GetCurrentDrivableActorSubscription(cd.Config.TSWAPISubscriptionIDStart); err != nil {
				return err
			}
		}

		cd.State.Controls.Clear()
		cd.State.DrivableActorName = subscription_result.ObjectClass
		for property_name, control := range subscription_result.Controls {
			control_state, _ := cd.State.Controls.Get(property_name)
			control_state.Identifier = control.Identifier
			control_state.PropertyName = control.PropertyName
			control_state.CurrentValue = control.CurrentValue
			control_state.CurrentNormalizedValue = control.CurrentNormalizedValue
			cd.State.Controls.Set(property_name, control_state)
		}
	}

	return nil
}

func (cd *CabDebugger) UpdateConfig(config CabDebugger_Config) {
	cd.Config = config
}

func (cd *CabDebugger) Clear() {
	cd.State.Controls.Clear()
}

func (cd *CabDebugger) Start(ctx context.Context) {
	go func() {
		socket_channel, unsubscribe_socket_channel := cd.Connector.Subscribe()
		ticker := time.NewTicker(333 * time.Millisecond)
		slow_ticker := time.NewTicker(5 * time.Second)
		for {
			tick_channel := ticker.C
			if !cd.TSWAPI.CanConnect() {
				tick_channel = slow_ticker.C
			}

			select {
			case msg := <-socket_channel:
				if msg.EventName == CURRENT_DRIVABLE_ACTOR_CONNECTOR_EVENT_NAME && msg.Properties["name"] != cd.State.DrivableActorName {
					go cd.updateCurrentDrivableActor(msg.Properties["name"])
				}
				if msg.EventName == SYNC_CONTROL_VALUE_CONNECTOR_EVENT_NAME {
					control_state, has_control_state := cd.State.Controls.Get(msg.Properties["property"])
					if cd.TSWAPI.Enabled() && cd.TSWAPI.CanConnect() && !has_control_state {
						/* if the API is enabled and connectable - it should drive the existance of the controls */
						continue
					}

					control_state.Identifier = msg.Properties["name"]
					control_state.PropertyName = msg.Properties["property"]
					current_value, _ := strconv.ParseFloat(msg.Properties["value"], 64)
					current_normalized_value, _ := strconv.ParseFloat(msg.Properties["normalized_value"], 64)
					control_state.CurrentValue = current_value
					control_state.CurrentNormalizedValue = current_normalized_value
					cd.State.Controls.Set(msg.Properties["property"], control_state)
				}
			case <-tick_channel:
				go cd.updateControlStateFromAPI()
			case <-ctx.Done():
				ticker.Stop()
				slow_ticker.Stop()
				unsubscribe_socket_channel()
				return
			}
		}
	}()

}

func NewCabDebugger(tswapi *tswapi.TSWAPI, socket_conn tswconnector.TSWConnector, config CabDebugger_Config) *CabDebugger {
	return &CabDebugger{
		Connector: socket_conn,
		TSWAPI:    tswapi,
		Config:    config,
		State: CabDebugger_ControlState{
			Mutex:             sync.Mutex{},
			DrivableActorName: "",
			Controls:          map_utils.NewLockMap[PropertyName, CabDebugger_ControlState_Control](),
		},
	}
}
