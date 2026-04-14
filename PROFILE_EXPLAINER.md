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
  "threshold": 0.5,
  "action_activate": { ... },
  "action_deactivate": { ... }
}
```

- **First activation** runs `action_activate`.
- **Next activation** runs `action_deactivate`.
- Useful for switches like headlights, engine start, etc.

### 📈 Linear

Used for analog levers or sliders with multiple threshold points.

```json
{
  "type": "linear",
  "thresholds": [
    { "threshold": 0.2, "action_activate": { ... }, "action_deactivate": { ... } },
    { "threshold": 0.7, "action_activate": { ... }, "action_deactivate": { ... } }
  ]
}
```

- Triggers **different actions** based on **axis position thresholds**.
- Ideal for **brake levers**, **throttles**, etc.

### 🎚️ DirectControl

Maps an analog controller input to a continuous value in-game.

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
- Used for **continuous analog mappings**.
- Supports `step` or `steps` to quantize values.
- Can be used with the `{SIDE}` placeholder to automatically select the correct side of the cab. This is specifically for controls named `Throttle_F` or `Throttle_B` where the `F` and `B` mark the side of the cab.
**Note: some locomotives don't use the F and R placeholders. The Czech route locomotives for example use 1 and 2. To support this you can use the expanded placeholder which defines which characters to use for front and back: {SIDE:F:B} [example with 1/2: {SIDE:1:2}]**

#### Options

| Name     | Description                                                                                                                                  |
| -------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| `hold`   | Whether to continuously hold this value. Useful for levers which automatically reset. (such as the Tube Deadman or some brake levers)        |
| `notify` | Whether to enable the in-game notifier when changing values to display the current value (defaults to `true` but can be explicitly disabled) |

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
- Can be defined as a relative value (instead of sending the absolute value)

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
  "invert": true
}
```

- `min` / `max`: Range of values.
- `step`: Optional increment size.
- `steps`: Optional list of discrete valid values. Can be used with `null` values to create zones of free motion between detents.
- `invert`: Whether to reverse the axis.
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

| Property | Description | Required |
|----------|-------------|----------|
| `threshold` | The actual threshold value where the step begins | ✅ Yes |
| `threshold_end` | The end value for range-based thresholds (optional) | No |
| `threshold_tolerance` | The tolerance around the threshold (optional) | No |

#### How It Works

1. **Single Threshold**: When only `threshold` is specified, it defines a single point value. The control will snap to this value when the input matches the threshold (+- the tolerance).

2. **Range Threshold**: When both `threshold` and `threshold_end` are specified, it defines a range. The control will accept any input value within this range and map it proportionally. This is mostly useful for free range steps.

3. **Tolerance**: The `threshold_tolerance` defines how much deviation from the threshold is acceptable. For example, if `threshold` is 0.5 and `threshold_tolerance` is 0.05, the control will accept input values between 0.45 and 0.55.

4. **Default Tolerance**: If no tolerance is specified, a default tolerance is calculated based on the number of steps (approximately half the step size).

5. **Free Range Zones**: When a step is marked as a free range zone (using `null` in the `steps` array), it gets special handling with no tolerance by default. The threshold defines the boundaries of the free range.

**Note** it is important to note that the number of `step_thresholds` should match the number of `steps` as each step threshold definition corresponds to each step.

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

---

Happy simming! 🚂
