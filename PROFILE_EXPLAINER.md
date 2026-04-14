# 🎮 Controller Configuration Format

This document describes the structure and semantics of the configuration system used to map game controllers (e.g., joysticks, gamepads) to controls in Train Sim World and Train Simulator Classic. It is designed to be flexible, extensible, and friendly to both analog and digital input devices.

---

## 📦 Overview

Each control on a game controller can be assigned an **action**. Assignments describe _when_ and _how_ the actions are triggered based on the input. Actions describe _what_ happens when triggered.

All assignments conform to a top-level enum `ControllerProfileControlAssignment`, which contains the following variants:

| Type | Description |
|------|-------------|
| `Momentary` | Single-press controls that trigger on threshold crossing |
| `Toggle` | On/off state controls that alternate between actions |
| `Linear` | Multi-threshold controls for fine-grained behavior |
| `DirectControl` | Direct value mapping to the game |
| `SyncControl` | Synchronized control via keypresses |
| `ApiControl` | HTTP API-based value mapping |
| `VirtualAction` | Virtual controls actions |

Each assignment type has a specific use case and behavior, described below.

---

## 🧩 Assignment Types

### 🔘 Momentary

Used for buttons that act while held.

```json
{
  "type": "momentary",
  "threshold": 0.9,
  "match": "exceeds",
  "action_activate": { ... },
  "action_deactivate": { ... }
}
```

- **Triggers** when input value crosses `threshold`.
- **Deactivates** when input falls below `threshold`. (optional - by default if the `action_activate` defines a keystroke to be held; it will be released automatically when releasing the gamepad control)
- Ideal for **press-and-hold** style controls.

#### Properties

| Property | Description | Required |
|----------|-------------|----------|
| `threshold` | The threshold value that triggers the action | ✅ Yes |
| `match` | How to interpret the threshold. Defaults to `"exceeds"` where the action is executed when the value exceeds the threshold. Can also be set to `"equals"` for an exact comparison | No |
| `action_activate` | The action to execute when the threshold is exceeded | ✅ Yes |
| `action_deactivate` | The action to execute when the threshold is no longer exceeded. Defaults to releasing the previously activated key(s) | No |
| `conditions` | Optional conditions that must be met for the assignment to execute | No |
| `rail_class_information` | Optional list of rail classes this assignment applies to | No |

#### Examples

##### Simple Key Press

```json
{
  "type": "momentary",
  "threshold": 0.9,
  "action_activate": {
    "keys": "h"
  }
}
```

##### Key Press with Timing

```json
{
  "type": "momentary",
  "threshold": 0.9,
  "action_activate": {
    "keys": "w",
    "press_time": 0.2,
    "wait_time": 0.2
  }
}
```

##### Direct Control with API Fallback

```json
{
  "type": "momentary",
  "threshold": 0.9,
  "action_activate": {
    "controls": "Throttle_{SIDE}",
    "value": 1.0,
    "hold": true,
    "enable_api_fallback": true
  },
  "action_deactivate": {
    "controls": "Throttle_{SIDE}",
    "value": 0.0,
    "enable_api_fallback": true
  }
}
```

##### Conditional Momentary

Momentary assignments can be conditioned on other control values:

```json
{
  "type": "momentary",
  "threshold": 0.9,
  "conditions": [
    {
      "control": "mylever",
      "operator": "gte",
      "value": 0.5
    }
  ],
  "action_activate": {
    "keys": "h"
  }
}
```

In the above example, the assignment will only execute if `mylever` exceeds 0.5.

#### Supported Action Types

Momentary assignments support the following action types:

- **KeysAction**: Key presses with optional timing
- **DirectControlAction**: Direct control with optional `hold`, `notify`, `enable_api_fallback`, `use_normalized`, `max_change_rate`
- **APIControlAction**: API control with optional `hold`, `max_change_rate`
- **VirtualAction**: Virtual control actions

### 🔁 Toggle

Used for toggle switches that alternate between two states.

```json
{
  "type": "toggle",
  "threshold": 0.9,
  "match": "exceeds",
  "action_activate": { ... },
  "action_deactivate": { ... }
}
```

- **First activation** (when threshold is exceeded and no prior call exists) runs `action_activate`.
- **Second activation** (when threshold is exceeded and the previous call was also above threshold with the same action) runs `action_deactivate`.
- **Key release** (when below threshold and previous call was above threshold) releases key actions.
- Useful for switches like headlights, engine start, etc.

#### Properties

| Property | Description | Required |
|----------|-------------|----------|
| `threshold` | The threshold value that triggers the toggle action | ✅ Yes |
| `match` | How to interpret the threshold. Defaults to `"exceeds"` where the action is executed when the value exceeds the threshold. Can also be set to `"equals"` for an exact comparison | No |
| `action_activate` | The action to execute on the first toggle (when transitioning from off to on) | ✅ Yes |
| `action_deactivate` | The action to execute on the second toggle (when transitioning from on to off) | ✅ Yes |
| `conditions` | Optional conditions that must be met for the assignment to execute | No |
| `rail_class_information` | Optional list of rail classes this assignment applies to | No |

#### Toggle State Machine

The Toggle assignment maintains internal state to track whether it has been activated. The behavior follows this logic:

1. **Initial State (No prior call)**: When the threshold is exceeded, `action_activate` is executed.
2. **Toggled State (Previous call was above threshold)**: When the threshold is exceeded again, `action_deactivate` is executed.
3. **Below Threshold**: When the input falls below the threshold:
   - If the previous action was a key press, it is released.
   - If the previous action was a direct control or API control, no action is taken (the value is not automatically reverted).

#### Examples

##### Simple Key Toggle

```json
{
  "type": "toggle",
  "threshold": 0.9,
  "action_activate": {
    "keys": "x"
  },
  "action_deactivate": {
    "keys": "shift+x"
  }
}
```

Pressing the button once triggers `x`, pressing again triggers `shift+x`.

##### Direct Control Toggle

```json
{
  "type": "toggle",
  "threshold": 0.9,
  "action_activate": {
    "controls": "MarkerLight_R",
    "value": 1
  },
  "action_deactivate": {
    "controls": "MarkerLight_R",
    "value": 0.5
  }
}
```

Pressing the button once sets the marker light to 1, pressing again sets it to 0.5.

##### Toggle with Timing

```json
{
  "type": "toggle",
  "threshold": 0.9,
  "action_activate": {
    "keys": "w",
    "press_time": 0.2,
    "wait_time": 0.2
  },
  "action_deactivate": {
    "keys": "s",
    "press_time": 0.2,
    "wait_time": 0.2
  }
}
```

##### Conditional Toggle

Toggle assignments can be conditioned on other control values:

```json
{
  "type": "toggle",
  "threshold": 0.9,
  "conditions": [
    {
      "control": "mylever",
      "operator": "gte",
      "value": 0.5
    }
  ],
  "action_activate": {
    "keys": "h"
  },
  "action_deactivate": {
    "keys": "shift+h"
  }
}
```

In the above example, the toggle will only execute if `mylever` exceeds 0.5.

##### Toggle with Equals Match

```json
{
  "type": "toggle",
  "threshold": 0.5,
  "match": "equals",
  "action_activate": {
    "keys": "a"
  },
  "action_deactivate": {
    "keys": "b"
  }
}
```

This toggle only triggers when the input value exactly equals 0.5.

#### Supported Action Types

Toggle assignments support the following action types:

- **KeysAction**: Key presses with optional timing
- **DirectControlAction**: Direct control with optional `hold`, `notify`, `enable_api_fallback`, `use_normalized`, `max_change_rate`
- **APIControlAction**: API control with optional `hold`, `max_change_rate`
- **VirtualAction**: Virtual control actions

### 📈 Linear

Used for analog levers or sliders with multiple threshold points.

```json
{
  "type": "linear",
  "neutral": 0.5,
  "thresholds": [
    { "value": 0.2, "action_activate": { ... } },
    { "value": 0.5, "action_activate": { ... } },
    { "value": 0.8, "action_activate": { ... } }
  ]
}
```

- Triggers **different actions** based on **axis position thresholds**.
- Ideal for manual implementation of **brake levers**, **throttles**, etc.
- Supports **auto-generation** of thresholds using `value_end` and `value_step`.
- Supports **neutral value mapping** to center the input range.

#### Properties

| Property | Description | Required |
|----------|-------------|----------|
| `thresholds` | Array of threshold definitions with actions | ✅ Yes |
| `neutral` | Optional neutral/idle value for value normalization (e.g., 0.5 for centering 0-1 range to -1 to 1) | No |
| `conditions` | Optional conditions that must be met for the assignment to execute | No |
| `rail_class_information` | Optional list of rail classes this assignment applies to | No |

#### Threshold Properties

| Property | Description | Required |
|----------|-------------|----------|
| `value` | The threshold value to exceed. Can use named references from calibration | ✅ Yes |
| `value_end` | End value for auto-generating thresholds (exclusive) | No |
| `value_step` | Step increment for auto-generating thresholds between `value` and `value_end` | No |
| `action_activate` | The action to execute when the threshold is exceeded | ✅ Yes |
| `action_deactivate` | The action to execute when the value falls below the threshold | No |

#### Auto-Generation of Thresholds

When `value_end` and `value_step` are provided, the system automatically generates multiple thresholds between `value` and `value_end`:

```json
{
  "type": "linear",
  "thresholds": [
    {
      "value": 0.3,
      "value_end": 0.6,
      "value_step": 0.05,
      "action_activate": { "keys": "w" }
    }
  ]
}
```

This generates thresholds at: 0.3, 0.35, 0.40, 0.45, 0.50, 0.55 (all triggering the same action).

#### Neutral Value Mapping

The `neutral` property allows you to map the input value range around a neutral point. This is useful for levers that have a centered neutral position:

```json
{
  "type": "linear",
  "neutral": 0.5,
  "thresholds": [
    { "value": -0.5, "action_activate": { "keys": "s" } },
    { "value": 0.5, "action_activate": { "keys": "w" } }
  ]
}
```

With `neutral: 0.5`, a raw input of 0.0 becomes -1.0 (neutralized), and 1.0 becomes 1.0.

#### Threshold Exceeding Logic

The Linear assignment uses asymmetric threshold logic (signifying applying power as opposed to mathmatical operations):
- **Negative values**: A value "exceeds" the threshold when it is **less than** the threshold (more negative)
- **Positive values**: A value "exceeds" the threshold when it is **greater than or equal to** the threshold

#### State Machine Behavior

Linear assignments track which thresholds are currently exceeding and which were previously passed:

1. **Activation**: When new thresholds start exceeding (that weren't previously passed), their `action_activate` is triggered.
2. **Deactivation**: When thresholds stop exceeding (were passed but no longer are), their `action_deactivate` is triggered if defined.
3. **Key release**: If `action_deactivate` is not defined but `action_activate` uses keys, the keys are released.
4. **Clear state**: If neither deactivation nor key release is possible, the previous call is cleared to allow re-triggering.

#### Examples

##### Simple Key Press at Thresholds

```json
{
  "type": "linear",
  "thresholds": [
    { "value": 0.2, "action_activate": { "keys": "a" } },
    { "value": 0.5, "action_activate": { "keys": "d" } },
    { "value": 0.8, "action_activate": { "keys": "f" } }
  ]
}
```

As the lever moves from 0 to 1:
- At 0.2: triggers `a`
- At 0.5: triggers `d`
- At 0.8: triggers `f`

##### Brake Lever with Neutral Position

```json
{
  "type": "linear",
  "neutral": 0.5,
  "thresholds": [
    { "value": -0.5, "action_activate": { "keys": "s" } },
    { "value": 0.5, "action_activate": { "keys": "w" } }
  ]
}
```

With a neutral value of 0.5, the lever's centered position (0.5 raw) is treated as 0 (neutralized), allowing negative and positive threshold values.

##### Auto-Generated Thresholds for Notched Control

```json
{
  "type": "linear",
  "thresholds": [
    {
      "value": 0.0,
      "value_end": 0.5,
      "value_step": 0.1,
      "action_activate": { "keys": "1" }
    },
    {
      "value": 0.6,
      "value_end": 1.0,
      "value_step": 0.1,
      "action_activate": { "keys": "2" }
    }
  ]
}
```

This creates 6 thresholds for each range (0.0, 0.1, 0.2, 0.3, 0.4, 0.5 and 0.6, 0.7, 0.8, 0.9, 1.0), all triggering the same action within each range.

##### Conditional Linear Assignment

Linear assignments can be conditioned on other control values:

```json
{
  "type": "linear",
  "conditions": [
    {
      "control": "mylever",
      "operator": "gte",
      "value": 0.5
    }
  ],
  "thresholds": [
    { "value": 0.2, "action_activate": { "keys": "a" } },
    { "value": 0.7, "action_activate": { "keys": "d" } }
  ]
}
```

In the above example, the assignment will only execute if `mylever` exceeds 0.5.

#### Supported Action Types

Linear assignments support the following action types:

- **KeysAction**: Key presses for discrete actions at thresholds
- **DirectControlAction**: Direct control with optional `hold`, `notify`, `enable_api_fallback`, `use_normalized`, `max_change_rate`
- **APIControlAction**: API control with optional `hold`, `max_change_rate`
- **VirtualAction**: Virtual control actions

### 🎚️ DirectControl

Maps an analog controller input to a continuous value in-game. This is the primary method for controlling cab levers and other continuous controls in Train Sim World, Train Simulator Classic and Wonders of Sodor.

```json
{
  "type": "direct_control",
  "controls": "Throttle1",
  "input_value": {
    "min": 0.0,
    "max": 1.0,
    "invert": true
  },
  "notify": true
}
```

- **Directly updates** a game control based on physical control input.
- Used for **continuous analog mappings** (throttles, brakes, levers).
- Supports `step` or `steps` to quantize values.
- Can be used with the `{SIDE}` placeholder to automatically select the correct side of the cab. This is specifically for controls named `Throttle_F` or `Throttle_B` where the `F` and `B` mark the side of the cab.
**Note: some locomotives don't use the F and R placeholders. The Czech route locomotives for example use 1 and 2. To support this you can use the expanded placeholder which defines which characters to use for front and back: {SIDE:F:B} [example with 1/2: {SIDE:1:2}]**

#### Properties

| Name                | Description                                                                                                                                  | Required |
| ------------------- | -------------------------------------------------------------------------------------------------------------------------------------------- | -------- |
| `controls`          | The UE4SS control identifier to control (e.g., `Throttle`, `AutomaticBrake`, `IndependentBrake`)                                            | ✅ Yes   |
| `input_value`       | Defines the input value constraints and mapping                                                    | ✅ Yes   |
| `control_range`     | Remaps a partial input range to a full 0-1 or 0,-1 output range (see [control_range example](#control-range-remapping))                      | No       |
| `hold`              | Whether to continuously hold this value. Useful for levers which automatically reset (such as the Deadman or some brake levers)          | No       |
| `use_normalized`    | Whether to use normalized values (-1 to 1) instead of raw values (0 to 1) [rarely used]                                                                    | No       |
| `enable_api_fallback` | Whether to enable fallback to the TSW API if direct control is unavailable                                                                  | No       |
| `notify`            | Whether to enable the in-game notifier when changing values to display the current value (defaults to `true`)                                  | No       |
| `conditions` | Optional conditions that must be met for the assignment to execute | No |
| `rail_class_information` | Optional list of rail classes this assignment applies to | No |

#### input_value Properties

| Name                | Description                                                                                                                                  | Required |
| ------------------- | -------------------------------------------------------------------------------------------------------------------------------------------- | -------- |
| `min`               | The minimum reachable value in the game cab                                                | ✅ Yes   |
| `max`               | The maximum reachable value in the game cab                                              | ✅ Yes   |
| `step`              | The step increment to auto-generate discrete values (alternative to `steps`)                                                                 | No       |
| `steps`             | Array of discrete values. Can include `null` to create free range zones between detents                                                      | No       |
| `invert`            | Whether to reverse the input value direction (mapping 0-1 to 1-0)                                                                                                | No       |
| `max_change_rate`   | The maximum rate at which the control value can change per frame  (rarely necessary)                                                                            | No       |
| `step_thresholds`   | Custom threshold definitions for each step (see [step_thresholds](#step-thresholds-custom-thresholds-for-stepped-controls)) | No       |

##### Step Thresholds - Custom Thresholds for Stepped Controls

The `step_thresholds` option allows you to define custom threshold values for each step in a direct control mapping. This is particularly useful when you want to remap a continuous analog input to match a notched control (e.g., a throttle with detents) or when you need to create custom value ranges.

```json
{
  "type": "direct_control",
  "controls": "Throttle1",
  "input_value": {
    "min": 0.1,
    "max": 1.0,
    "steps": [0.1, null, 1.0],
    "step_thresholds": [
      { "threshold": 0.2, "threshold_tolerance": 0.05 },
      { "threshold": 0.5, "threshold_end": 0.6, "threshold_tolerance": 0.03 },
      { "threshold": 0.8 }
    ]
  }
}
```

**Threshold Properties:**

| Property                    | Description                                                                                                                                  | Required |
| --------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------- | -------- |
| `threshold`                 | The actual threshold value where the step begins (can use named references from calibration)                                                 | ✅ Yes   |
| `threshold_end`             | The end value for range-based thresholds (optional)                                                                                          | No       |
| `threshold_tolerance`       | The tolerance around the threshold (optional)                                                                                                | No       |

**How It Works:**

1. **Single Threshold**: When only `threshold` is specified, it defines a single point value. The control will snap to this value when the input matches the threshold (+- the tolerance).

2. **Range Threshold**: When both `threshold` and `threshold_end` are specified, it defines a range. The control will accept any input value within this range and map it proportionally. This is mostly useful for free range steps.

3. **Tolerance**: The `threshold_tolerance` defines how much deviation from the threshold is acceptable. For example, if `threshold` is 0.5 and `threshold_tolerance` is 0.05, the control will accept input values between 0.45 and 0.55.

4. **Default Tolerance**: If no tolerance is specified, a default tolerance is calculated based on the number of steps (approximately half the step size).

5. **Free Range Zones**: When a step is marked as a free range zone (using `null` in the `steps` array), it gets special handling with no tolerance by default. The threshold defines the boundaries of the free range.

**Note:** The number of `step_thresholds` should match the number of `steps` as each step threshold definition corresponds to each step.

**Use Cases:**

- **Notched Controls**: Match a physical control with detents (like a notched throttle) by defining thresholds at each detent position.
- **Custom Value Ranges**: Create custom value ranges where certain input values map to specific output values.

##### Control Range - Remapping Partial Ranges

The `control_range` property allows you to remap a partial input range to a full 0-1 or 0,-1 output range. This can be useful for mapping a single physical lever or control to multiple in-game controls while retaining a full 0-1 direct control range.

```json
{
  "type": "direct_control",
  "controls": "Throttle1",
  "input_value": {
    "min": 0.0,
    "max": 1.0
  },
  "control_range": {
    "start": 0.2,
    "end": 0.8
  }
}
```

In the above example, an input value of 0.2 maps to 0.0, 0.8 maps to 1.0, and values outside this range are clamped. This effectively remaps the 0.2-0.8 input range to the full 0-1 output range.

**Control Range Properties:**

| Property | Description                                                                                                                                  | Required |
| -------- | -------------------------------------------------------------------------------------------------------------------------------------------- | -------- |
| `start`  | The start value of the input range to remap (depending on value direction, this is the min or max)                                         | ✅ Yes   |
| `end`    | The end value of the input range to remap                                                                                                   | ✅ Yes   |

##### Steps with Free Range Zones

The `steps` array can include `null` values to create free range zones between detents. This is useful for controls that are partly notched and partly free.

```json
{
  "type": "direct_control",
  "controls": "AutomaticBrake1",
  "input_value": {
    "min": 0,
    "max": 0.8,
    "steps": [0, 0.125, null, 0.6, 0.7, 0.8]
  }
}
```

In the above example:
- Steps at 0, 0.125 are discrete detents
- The `null` creates a free range zone between 0.125 and 0.6
- Steps at 0.6, 0.7, 0.8 are discrete detents after the free range zone

##### Conditional Direct Control Assignments

Direct Control assignments can be conditioned on other control values, allowing the same physical control to map to different game controls based on conditions.

```json
{
  "type": "direct_control",
  "conditions": [
    {
      "control": "mylever",
      "operator": "lt",
      "value": 0.5
    }
  ],
  "controls": "IndependentBrake",
  "input_value": {
    "min": 0.25,
    "max": 1,
    "invert": true
  }
}
```

In the above example, the IndependentBrake is only controlled when the Reverser (`mylever`) is less than 0.5. When the Reverser exceeds 0.5, a separate Direct Control assignment would map to DynamicBrake.

#### Examples

##### Basic Throttle Mapping

```json
{
  "type": "direct_control",
  "controls": "Throttle_{SIDE}",
  "input_value": {
    "min": 0,
    "max": 1
  }
}
```

##### Brake with Discrete Steps

```json
{
  "type": "direct_control",
  "controls": "AutomaticBrake_{SIDE}",
  "input_value": {
    "min": 0,
    "max": 0.8,
    "steps": [0, 0.125, 0.25, 0.375, 0.5, 0.625, 0.75, 0.8]
  }
}
```

### 🧭 SyncControl

An alternative to `DirectControl` for locomotives that don't work with direct control.

```json
{
  "type": "sync_control",
  "identifier": "Reverser1",
  "input_value": {
    "min": -1.0,
    "max": 1.0,
    "steps": [-1.0, 0.0, 1.0]
  },
  "action_increase": { "keys": "PageUp" },
  "action_decrease": { "keys": "PageDown" }
}
```

- **Reads current in-game state** and uses **keypresses** to reach desired state.
- Ideal for **syncing with controls that don't respond well to direct manipulation**.

### 🎚️ ApiControl

Maps an analog controller input to a continuous value in-game using the HTTP API.

```json
{
  "type": "api_control",
  "controls": "Throttle1",
  "input_value": {
    "min": 0.0,
    "max": 1.0,
    "invert": true
  }
}
```

- **Directly updates** a game control based on axis input using the HTTP API. May result in slight overhead compared to the full direct control mode, but does not require additional mod to be installed.
- Used for **continuous analog mappings**.
- Supports `step` or `steps` to quantize values.

---

## ⚙️ Action Types

Each assignment triggers an action when activated (and optionally when deactivated). Actions can be:

### 🖱️ Key Presses

```json
{
  "keys": "W",
  "press_time": 0.1,
  "wait_time": 0.05
}
```

- Simulates a key press.
- Optional timing controls for holding and releasing.

### 🎛️ Direct Control Action

```json
{
  "controls": "Throttle1",
  "value": 0.5,
  "hold": false,
  "relative": false
}
```

- Sends a value directly to a UE4SS control.
- Can be held or pulsed.
- Can be defined as a relative value (instead of sending the absolute value).

#### Direct Control Action Properties

| Property                | Description                                                                                                                                  | Required |
| ----------------------- | -------------------------------------------------------------------------------------------------------------------------------------------- | -------- |
| `controls`              | The UE4SS control identifier to control                                                           | ✅ Yes   |
| `value`                 | The value to send to the control                                                                                                             | ✅ Yes   |
| `max_change_rate`       | The maximum rate at which the control value can change per frame                                                                             | No       |
| `relative`              | Whether to use the value as a relative adjustment instead of absolute                                                                        | No       |
| `hold`                  | Whether to continuously hold the value by sending it repeatedly                                                                              | No       |
| `use_normalized`        | Whether to use normalized values instead of raw values (rarely used)                                                                                      | No       |
| `notify`                | Whether to enable the in-game notifier when changing values                                                                                  | No       |
| `enable_api_fallback`   | Whether to enable fallback to the TSW API if direct control is unavailable                                                                   | No       |

### 🎛️ Api Control Action

```json
{
  "controls": "Throttle1",
  "api_value": 0.5
}
```

- Sends a value directly to a control using the HTTP API.

### 🎛️ Virtual Action

```json
{
  "type": "virtual",
  "control": "virtual:MyVirtualControl",
  "value": 0.1
}
```

- Sets the value of a virtual control which in turn can activate other assignments.

---

## 🔧 Input Value Mapping

Used by `DirectControl`, `SyncControl`, and `ApiControl` to map axis input to control values.

```json
{
  "min": -1.0,
  "max": 1.0,
  "step": 0.1,
  "steps": [0.0, 0.2, null, 0.5, null, 1.0],
  "invert": true,
  "max_change_rate": 0.05
}
```

- `min` / `max`: Range of values.
- `step`: Optional increment size for auto-generating discrete values.
- `steps`: Optional list of discrete valid values. Can be used with `null` values to create zones of free motion between detents.
- `invert`: Whether to reverse the axis direction.
- `max_change_rate`: The maximum rate at which the control value can change per frame (optional).
- `step_thresholds`: Optional array of threshold definitions for remapping values to match notched or stepped controls.

### 📏 Step Thresholds

The `step_thresholds` option allows you to define custom threshold values for each step in a direct control mapping. This is particularly useful when you want to remap a continuous analog input to match a notched control (e.g., a throttle with detents) or when you need to create custom value ranges.

```json
{
  "type": "direct_control",
  "controls": "Throttle1",
  "input_value": {
    "min": -1.0,
    "max": 1.0,
    "steps": [0.1, null, 1.0],
    "step_thresholds": [
      { "threshold": 0.2, "threshold_tolerance": 0.05 },
      { "threshold": 0.5, "threshold_end": 0.6, "threshold_tolerance": 0.03 },
      { "threshold": 0.8 }
    ]
  }
}
```

#### Threshold Properties

| Property                    | Description                                                                                                                                  | Required |
| --------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------- | -------- |
| `threshold`                 | The actual threshold value where the step begins (can use named references from calibration)                                                 | ✅ Yes   |
| `threshold_end`             | The end value for range-based thresholds (optional)                                                                                          | No       |
| `threshold_tolerance`       | The tolerance around the threshold (optional)                                                                                                | No       |

#### How It Works

1. **Single Threshold**: When only `threshold` is specified, it defines a single point value. The control will snap to this value when the input matches the threshold (+- the tolerance).

2. **Range Threshold**: When both `threshold` and `threshold_end` are specified, it defines a range. The control will accept any input value within this range and map it proportionally. This is mostly useful for free range steps.

3. **Tolerance**: The `threshold_tolerance` defines how much deviation from the threshold is acceptable. For example, if `threshold` is 0.5 and `threshold_tolerance` is 0.05, the control will accept input values between 0.45 and 0.55.

4. **Default Tolerance**: If no tolerance is specified, a default tolerance is calculated based on the number of steps (approximately half the step size).

5. **Free Range Zones**: When a step is marked as a free range zone (using `null` in the `steps` array), it gets special handling with no tolerance by default. The threshold defines the boundaries of the free range.

**Note:** It is important that the number of `step_thresholds` matches the number of `steps` as each step threshold definition corresponds to each step.

#### Use Cases

- **Notched Controls**: Match a physical control with detents (like a notched throttle) by defining thresholds at each detent position.
- **Custom Value Ranges**: Create custom value ranges where certain input values map to specific output values.
- **Hysteresis**: Define different thresholds for activation and deactivation to prevent rapid toggling.

---

## 🔁 Conditional Assignments

It is also possible to only execute assignments depending on one or more conditions. This can be used to create multi-key assignments (e.g., the action of a button changes depending on the position of a lever).

This can be added to any assignment using the `conditions` key:

```json
{
  "type": "momentary",
  "conditions": [
    {
      "control": "mylever",
      "operator": "gte",
      "value": 0.5
    }
  ]
}
```

In the above example, the assignment will only execute if `mylever` exceeds 0.5. At this time the supported operators are `eq`,`gte`, `lte`, `gt`, and `lt`.

---

## 🏗️ Profile Structure

A controller profile is defined in a JSON file with the following structure:

```json
{
  "name": "MyProfile",
  "extends": "BaseProfile",
  "auto_select": true,
  "controls": [
    {
      "name": "Button1",
      "type": "momentary",
      "threshold": 0.5,
      "action_activate": { "keys": "H" },
      "action_deactivate": { "keys": "Shift+H" }
    }
  ],
  "controller": {
    "usb_id": "0x1234",
    "mapping": "Standard",
    "calibration": { ... }
  },
  "rail_class_information": [
    "Class 40",
    "Class 42",
    "Class 43"
  ]
}
```

### Root Properties

| Property | Description | Required |
|----------|-------------|----------|
| `name` | Profile name | ✅ Yes |
| `extends` | Profile to extend from | No |
| `auto_select` | Auto-detection support | No |
| `controls` | Array of control definitions | No |
| `controller` | Controller-specific info | No |
| `rail_class_information` | Supported rail classes | No |

### Controller Section

The `controller` section contains controller-specific information. It is optional and is mostly available for profile sharing.

```json
"controller": {
  "usb_id": "...",
  "mapping": { ... },
  "calibration": { ... }
}
```

- `usb_id`: The USB device ID of the controller
- `mapping`: The mapping profile used for the controller
- `calibration`: Calibration data for the controller

### Rail Class Information

The `rail_class_information` section is an array of supported rail class names:

```json
"rail_class_information": [{ "class_name": "..." }]
```

- Contains a list of rail class names (e.g., "Class 40", "Class 42", etc.)
- Used to specify which train classes the profile is compatible with

---

## ✅ Best Practices

- Use `DirectControl` for stable, high-resolution mappings, especially lever controls.
- Use `ApiControl` if you are unable to or do not want to use `DirectControl` (`ApiControl` is less flexible and is less performant, but still provides a near direct control option).
- Use `SyncControl` if you want a direct control-like experience but want to use keybindings.
- Use `Linear` for fine-grained, manually configured lever behavior.
- Use `Momentary` for temporary actions like horn or bell.
- Use `Toggle` for switches with two states.
- Use `VirtualAction` for controlling virtual controls.
- Use `DirectControl` as an action within `Momentary` or `Toggle` assignments for discrete control value changes (e.g., setting brake levels with buttons).

---

Happy simming! 🚂
