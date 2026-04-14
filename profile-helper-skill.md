---
name: tsw-controller-utility-profile-agent-skill
description: >
  This skill enables an LLM to understand, explain, diagnose problems in, and fix Train Sim World controller profiles based on the documented configuration format.
---

## Reference Document
Use the documentation from: https://github.com/LiahMartens/tsw-controller-app/blob/main/PROFILE_EXPLAINER.md

## Core Capabilities

### 1. Explain Profiles
When asked to explain a profile or configuration, analyze and describe:
- The overall purpose and structure of the profile
- Each control assignment and its behavior
- The controller hardware and mapping details
- Any conditional logic or special behaviors

**Example prompts:**
- "Explain this profile configuration"
- "What does this momentary assignment do?"
- "How does the throttle mapping work?"

### 2. Diagnose Problems
When presented with a profile that has issues, identify:
- Missing required properties
- Invalid or incorrect property values
- Logical errors in threshold configurations
- Incompatible assignment types for the intended use case
- Missing conditions that should be present
- Incorrect action configurations

**Example prompts:**
- "Why isn't my throttle working?"
- "What's wrong with this brake lever configuration?"
- "This button doesn't trigger - help diagnose"

### 3. Fix Profiles
When asked to fix or improve a profile:
- Correct syntax errors and missing required fields
- Suggest appropriate assignment types for the use case
- Optimize threshold values and conditions
- Ensure proper action configurations
- Maintain backward compatibility when extending profiles

**Example prompts:**
- "Fix this profile configuration"
- "How do I add a toggle for the marker light?"
- "Convert this momentary to a direct control"

## Assignment Types Reference

### Momentary
- **Use case**: Single-press controls that trigger on threshold crossing
- **Key properties**: `threshold`, `match`, `action_activate`, `action_deactivate`
- **Behavior**: Triggers when input crosses threshold, deactivates when falling below
- **Best for**: Horn, bell, buttons, temporary actions

### Toggle
- **Use case**: On/off state controls that alternate between actions
- **Key properties**: `threshold`, `match`, `action_activate`, `action_deactivate`
- **Behavior**: First activation runs `action_activate`, second runs `action_deactivate`
- **Best for**: Headlights, engine start, switches with two states

### Linear
- **Use case**: Multi-threshold controls for fine-grained behavior
- **Key properties**: `thresholds`, `neutral`, `conditions`
- **Behavior**: Triggers different actions based on axis position thresholds
- **Best for**: Brake levers, throttles with manual threshold configuration
- **Special features**: Auto-generation with `value_end` and `value_step`, neutral value mapping

### DirectControl
- **Use case**: Direct value mapping to the game
- **Key properties**: `controls`, `input_value`, `control_range`, `hold`
- **Behavior**: Directly updates a game control based on physical control input
- **Best for**: Cab levers, continuous analog mappings
- **Special features**: `{SIDE}` placeholder for cab selection, step thresholds for notched controls

### SyncControl (DEPRECATED)
- **Use case**: State machine approach using keypresses
- **Key properties**: `identifier`, `input_value`, `action_increase`, `action_decrease`
- **Behavior**: Reads current in-game state and uses keypresses to reach target
- **Note**: Generally not recommended in favor of DirectControl or ApiControl

### ApiControl
- **Use case**: HTTP API-based value mapping
- **Key properties**: `controls`, `input_value`, `control_range`, `hold`
- **Behavior**: Runs at 15fps processing loop, sends values via HTTP API
- **Best for**: When DirectControl is unavailable
- **Special features**: `max_change_rate` for realistic throttle simulation

## Action Types Reference

### KeysAction
- Simulates key presses with optional timing
- Properties: `keys`, `press_time`, `wait_time`

### DirectControlAction
- Sends values directly to UE4SS controls
- Properties: `controls`, `value`, `hold`, `relative`, `max_change_rate`, `notify`, `enable_api_fallback`

### ApiControlAction
- Sends values via HTTP API
- Properties: `controls`, `api_value`

### VirtualAction
- Sets values of virtual controls
- Properties: `type`, `control`, `value`

## Input Value Mapping Reference

### Common Properties
- `min` / `max`: Range of values
- `step`: Increment size for auto-generating discrete values
- `steps`: List of discrete valid values (can include `null` for free range zones)
- `invert`: Reverse axis direction
- `max_change_rate`: Maximum rate of value change per frame
- `step_thresholds`: Custom thresholds for remapping values

### Step Thresholds
- Define custom threshold values for each step
- Properties: `threshold`, `threshold_end`, `threshold_tolerance`
- Use cases: Notched controls, custom value ranges, hysteresis

## Conditional Assignments

Assignments can be conditioned on other control values using the `conditions` array:

```json
{
  "conditions": [
    {
      "control": "mylever",
      "operator": "gte",
      "value": 0.5
    }
  ]
}
```

Supported operators: `eq`, `gte`, `lte`, `gt`, `lt`

## Profile Structure Reference

### Root Properties
- `name`: Profile name (required)
- `extends`: Profile to extend from (optional)
- `auto_select`: Auto-detection support (optional)
- `controls`: Array of control definitions (optional)
- `controller`: Controller-specific info (optional)
- `rail_class_information`: Supported rail classes (optional)

### Controller Section
- `usb_id`: USB device ID
- `mapping`: Mapping profile used
- `calibration`: Calibration data

### Rail Class Information
- Array of rail class names (e.g., "Class 40", "Class 42")
- Specifies which train classes the profile is compatible with

## Best Practices Reference

1. Use `DirectControl` for stable, high-resolution mappings (especially levers)
2. Use `ApiControl` if DirectControl is unavailable or not desired
3. Use `SyncControl` for direct control-like experience with keybindings (deprecated)
4. Use `Linear` for fine-grained, manually configured lever behavior
5. Use `Momentary` for temporary actions (horn, bell)
6. Use `Toggle` for switches with two states
7. Use `VirtualAction` for controlling virtual controls
8. Use `DirectControl` as an action within `Momentary` or `Toggle` for discrete control value changes

## Response Guidelines

When responding to user queries:

1. **For explanations**: Be clear and concise, highlighting the purpose and behavior of each configuration element
2. **For diagnoses**: Identify the root cause, explain why it's a problem, and provide the fix
3. **For fixes**: Provide corrected code with explanations of changes made
4. **Always reference the documentation**: When explaining concepts, refer back to the PROFILE_EXPLAINER.md
5. **Consider use cases**: Suggest the most appropriate assignment type for the intended behavior
6. **Validate configurations**: Ensure all required properties are present and values are valid

## Example Interactions

### Explanation Request
**User**: "What does this momentary assignment do?"
**Agent**: "This momentary assignment triggers when the input value exceeds 0.9. When activated, it sends the 'h' key press. When deactivated (input falls below 0.9), it sends 'shift+h'. This is ideal for press-and-hold style controls like horn or bell."

### Diagnosis Request
**User**: "My throttle isn't working"
**Agent**: "I can help diagnose this. Please share your throttle configuration. Common issues include: missing `controls` property, incorrect `min`/`max` values, or using the wrong assignment type. DirectControl is recommended for throttle mappings."

### Fix Request
**User**: "How do I add a toggle for the marker light?"
**Agent**: "Here's a toggle configuration for the marker light:\n\n```json\n{\n  \"type\": \"toggle\",\n  \"threshold\": 0.9,\n  \"action_activate\": {\n    \"controls\": \"MarkerLight_R\",\n    \"value\": 1\n  },\n  \"action_deactivate\": {\n    \"controls\": \"MarkerLight_R\",\n    \"value\": 0.5\n  }\n}\n```\n\nThis will set the marker light to 1 on first press and 0.5 on second press."
