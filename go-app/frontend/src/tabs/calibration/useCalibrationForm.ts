import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { EventsOn } from "../../../wailsjs/runtime/runtime";
import { events } from "../../events";
import { main } from "../../../wailsjs/go/models";

export type Kind = "axis" | "button" | "hat";
export type CalibrationStateControl = {
  kind: Kind;
  index: number;
  value: number;
  name: string;
  min: number;
  max: number;
  idle: number;
  deadzone: number;
  antiJitter: number;
  invert: boolean;
  easingCurve: number[];
  override: boolean;
};
export type CalibrationState = {
  name: string;
  controls: CalibrationStateControl[];
  use_unique_id: boolean;
};

export type UseCalibrationFormType = ReturnType<typeof useCalibrationForm>;

const EMPTY_CONTROL_STATE: Omit<CalibrationStateControl, "kind" | "index"> = {
  name: "",
  value: 0,
  deadzone: 0,
  antiJitter: 0,
  invert: false,
  /* default to MAX_SAFE_INTEGER so any value is < min */
  min: Number.MAX_SAFE_INTEGER,
  idle: Number.MAX_SAFE_INTEGER,
  /* default to MIN_SAFE_INTEGER so any value is > max */
  max: Number.MIN_SAFE_INTEGER,
  easingCurve: [0.0, 0.0, 1.0, 1.0],
  override: false,
};

const roundToFactor = (value: number) => {
  const ROUNDING_FACTOR = 100;
  if (Math.abs(value) < 1)
    return Math.round(value * ROUNDING_FACTOR) / ROUNDING_FACTOR;
  return Math.round(value);
};

const applyDefaultDeadzoneToRawValue = (value: number) => {
  /* apply a default edge deadzone of 1% */
  return roundToFactor(value * 0.99);
};

export const useCalibrationForm = (defaultValues: CalibrationState) => {
  const form = useForm<CalibrationState>({
    defaultValues,
    mode: "onChange",
  });

  useEffect(() => {
    return EventsOn(events.rawevent, (data: main.Interop_RawEvent) => {
      const controls = form.getValues("controls");
      const existingIndex = form
        .getValues("controls")
        .findIndex((c) => c.kind === data.Kind && c.index === data.Index);

      const controlState: CalibrationStateControl =
        existingIndex === -1
          ? {
              ...EMPTY_CONTROL_STATE,
              kind: data.Kind as Kind,
              index: data.Index,
            }
          : { ...controls[existingIndex] };

      if (!controlState.override) {
        const nextState = {
          value: data.Value,
          deadzone: controlState.deadzone,
          antiJitter: controlState.antiJitter,
          min: Math.min(
            controlState.min,
            applyDefaultDeadzoneToRawValue(data.Value),
          ),
          max: Math.max(
            controlState.max,
            applyDefaultDeadzoneToRawValue(data.Value),
          ),
          idle: Math.min(
            controlState.min,
            applyDefaultDeadzoneToRawValue(data.Value),
          ),
        } as Partial<CalibrationStateControl>;
        nextState.deadzone = roundToFactor(
          Math.abs(controlState.max - controlState.min) * 0.01,
        );
        Object.assign(controlState, nextState);
      } else {
        controlState.value = data.Value;
      }

      if (existingIndex === -1) {
        form.setValue(
          "controls",
          [...form.getValues("controls"), controlState],
          {
            shouldDirty: true,
            shouldTouch: true,
          },
        );
      } else {
        form.setValue(`controls.${existingIndex}`, controlState, {
          shouldDirty: true,
          shouldTouch: true,
        });
      }
    });
  }, []);

  return form;
};
