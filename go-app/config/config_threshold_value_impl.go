package config

import (
	"encoding/json"
	"fmt"
	"tsw_controller_app/math_utils"
)

type thresholdValueAsObject struct {
	Value     *float64 `json:"value"`
	Reference *string  `json:"reference"`
}

func (v *Config_Threshold_Value) GetValue(thresholds map[string]float64, inverted bool) float64 {
	threshold_value := v.Value
	if v.Reference != nil {
		if value, has_ref := thresholds[*v.Reference]; has_ref {
			threshold_value = value
		}
	}
	if inverted {
		return math_utils.InvertInputValue(threshold_value)
	}
	return threshold_value
}

func (v *Config_Threshold_Value) UnmarshalJSON(data []byte) error {
	var obj thresholdValueAsObject
	if err := json.Unmarshal(data, &obj); err == nil {
		if obj.Value == nil {
			return fmt.Errorf("Invalid threshold value: missing \"value\" in JSON: %s", string(data))
		}

		*v = Config_Threshold_Value{Value: *obj.Value, Reference: obj.Reference}
		return nil
	}

	var num float64
	if err := json.Unmarshal(data, &num); err == nil {
		*v = Config_Threshold_Value{Value: num}
		return nil
	}

	return fmt.Errorf("unable to parse threshold as either an object or a number")
}

func (v Config_Threshold_Value) MarshalJSON() ([]byte, error) {
	if v.Reference != nil && *v.Reference != "" {
		/* a reference was set -> marshal as an object */
		as_dict := map[string]any{}
		as_dict["value"] = v.Value
		as_dict["reference"] = *v.Reference
		return json.Marshal(as_dict)
	}
	return json.Marshal(v.Value)
}
