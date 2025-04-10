# 🎮 Train Sim World Controller Configuration Format

This document describes the structure and semantics of the configuration system used to map game controllers (e.g., joysticks, gamepads) to controls in Train Sim World using UE4SS. It is designed to be flexible, extensible, and friendly to both analog and digital input devices.

---

## 📦 Overview

Each control on a game controller can be assigned an **action**. Assignments describe *when* and *how* the actions are triggered based on the input. Actions describe *what* happens when triggered.

All assignments conform to a top-level enum `ControllerProfileControlAssignment`, which contains the following variants:

- `Momentary`
- `Toggle`
- `Linear`
- `DirectControl`
- `SyncControl`

Each assignment type has a specific use case and behavior, described below.

---

## 🧩 Assignment Types

### 🔘 Momentary
Used for buttons that act while held.

```json
{
  "type": "momentary",
  "threshold": 0.5,
  "action_activate": { ... },
  "action_deactivate": { ... }
}
```

- **Triggers** when input value crosses `threshold`.
- **Deactivates** when input falls below `threshold`. (optional - by default if the `action_activate` defines a keystroke to be held; it will be released automatically when releasing the gamepad control)
- Ideal for **press-and-hold** style controls.

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
  "hold": true
}
```

- **Directly updates** a UE4SS control based on axis input.
- Used for **continuous analog mappings**.
- Supports `step` or `steps` to quantize values.

### 🧭 SyncControl
A safer alternative to `DirectControl` for unstable locos.

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
- Ideal for **syncing with controls that don’t respond well to direct manipulation**.

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
  "hold": false
}
```
- Sends a value directly to a UE4SS control.
- Can be held or pulsed.

---

## 🔧 Input Value Mapping

Used by `DirectControl` and `SyncControl` to map axis input to control values.

```json
{
  "min": -1.0,
  "max": 1.0,
  "step": 0.1,
  "invert": true
}
```

- `min` / `max`: Range of values.
- `step`: Optional increment size.
- `steps`: List of discrete valid values.
- `invert`: Whether to reverse the axis.

---

## ✅ Best Practices

- Use `DirectControl` for stable, high-resolution mappings.
- Use `SyncControl` only when direct manipulation is buggy.
- Use `Linear` for fine-grained lever behavior.
- Use `Momentary` for temporary actions like horn or bell.
- Use `Toggle` for switches with two states.

---

## 📝 Example Full Assignment
```json
{
  "type": "momentary",
  "threshold": 0.5,
  "action_activate": {
    "keys": "H"
  },
  "action_deactivate": {
    "keys": "Shift+H"
  }
}
```

---

Happy simming! 🚂

