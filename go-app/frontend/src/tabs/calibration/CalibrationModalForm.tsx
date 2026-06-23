import { useEffect, useState } from "react";
import { main } from "../../../wailsjs/go/models";
import {
  SaveCalibration,
  UnsubscribeRaw,
  SubscribeRaw,
  LoadConfiguration,
} from "../../../wailsjs/go/main/App";
import {
  CalibrationStateControl,
  Kind,
  useCalibrationForm,
} from "./useCalibrationForm";
import { CalibrationModalFormControl } from "./CalibrationModalFormControl";
import { alert } from "../../utils/alert";
import { useControllerConfiguration } from "../../swr";

type Props = {
  controller: main.Interop_GenericController;
  onClose: () => void;
};

export const CalibrationModalForm = ({ controller, onClose }: Props) => {
  const [isRunning, setIsRunning] = useState(false);
  const {
    data: controllerConfiguration,
    mutate: updateControllerConfiguration,
  } = useControllerConfiguration(controller);
  const form = useCalibrationForm({
    name: controllerConfiguration?.SDLMapping.name ?? "",
    /* enable the toggle if the SDL map has unique_id defined */
    use_unique_id: !!controllerConfiguration?.SDLMapping.unique_id,
    controls: (controllerConfiguration?.Calibration.Controls ?? [])
      .map(
        (control): CalibrationStateControl => ({
          kind: control.Kind as Kind,
          index: control.Index,
          name: control.Name,
          min: control.Min,
          max: control.Max,
          idle: control.Idle,
          antiJitter: control.AntiJitter,
          deadzone: control.Deadzone,
          invert: control.Invert,
          value: control.Idle,
          easingCurve: control.EasingCurve,
          override: true,
        }),
      )
      .toSorted((a, b) =>
        `${a.kind}_${a.index}`.localeCompare(`${b.kind}_${b.index}`),
      ),
  });
  const controls = form.watch("controls");

  const handleStart = async () => {
    try {
      await SubscribeRaw(controller.UniqueID);
      setIsRunning(true);
    } catch (err) {
      alert(`Could not start calibration (${err})`, "error");
    }
  };

  const handleCancel = async () => {
    try {
      await UnsubscribeRaw();
    } catch (err) {
      alert(`Could not cancel calibration (${err})`, "error");
    } finally {
      onClose();
    }
  };

  const handleStopAndSave = async () => {
    try {
      await UnsubscribeRaw();
      await form.handleSubmit(
        async (values) => {
          const data = new main.Interop_ControllerCalibration();
          data.Name = values.name;
          data.DeviceID = controller.DeviceID;
          if (values.use_unique_id) {
            /* send unique ID if enabled */
            data.UniqueID = controller.UniqueID;
          }
          data.Controls = values.controls
            .filter((c) => !!c.name)
            .map((control) => {
              const controlCalibration =
                controllerConfiguration?.Calibration.Controls.find(
                  (c) => c.Name === control.name,
                );
              return main.Interop_ControllerCalibration_Control.createFrom({
                Kind: control.kind,
                Index: control.index,
                Name: control.name,
                Min: control.min,
                Max: control.max,
                Idle: control.idle,
                Deadzone: control.deadzone,
                AntiJitter: control.antiJitter,
                Invert: control.invert,
                EasingCurve: control.easingCurve,
                Thresholds: controlCalibration?.Thresholds ?? [],
              });
            });
          await SaveCalibration(data);
          await LoadConfiguration();
          await updateControllerConfiguration();
        },
        (err) => {
          console.error(err);
          throw new Error(`Invalid submission`);
        },
      )();
      onClose();
    } catch (err) {
      alert(`Could not save calibration (${err})`, "error");
    }
  };

  useEffect(() => {
    return () => {
      /* force unsubscribe */
      UnsubscribeRaw();
    };
  }, []);

  return (
    <div>
      <h3 className="font-bold text-base">Configuring {controller?.Name}</h3>
      <div className="py-4 grid grid-cols-1 grid-flow-row auto-rows-max gap-2">
        <div>
          <label className="input input-xs">
            Controller Name
            <input
              type="text"
              className="grow"
              disabled={!isRunning}
              {...form.register(`name`, { required: true })}
            />
          </label>
        </div>

        <div className="alert alert-sm">
          <p className="text-sm">
            Note: by default when configuring controller axes a deadzone near
            the idle value of 1% will be applied. At the minimum and maximum
            ends of the axis a deadzone of 1% is also applied by default. This
            is because most axes can not hold their extreme values consistently.
            If 1% is not sufficient you can override the calibrated values and
            idle deadzone as necessary.
          </p>
        </div>

        <div>
          {controls.map((control, index) => (
            <div key={`${control.kind}_${control.index}`}>
              <CalibrationModalFormControl
                form={form}
                index={index}
                field={control}
                isRunning={isRunning}
              />
            </div>
          ))}
        </div>
        <div>
          <fieldset className="fieldset bg-base-100 border-base-300 rounded-box w-full border p-4">
            <legend className="fieldset-legend">Options</legend>
            <label className="label whitespace-normal">
              <input
                type="checkbox"
                className="checkbox"
                disabled={!isRunning}
                {...form.register("use_unique_id")}
              />
              Use unique ID for calibration (not recommended - only for complex
              use cases - identifier may not be stable).
            </label>
            {form.watch("use_unique_id") && (
              <div role="alert" className="alert alert-warning alert-soft mt-3">
                <span className="text-xs">
                  When unique ID calibration is enabled you have to ensure all
                  identical connected devices are also configured using unique
                  ID. Otherwise, your device may not be using the expected
                  configuration.
                </span>
              </div>
            )}
          </fieldset>
        </div>
      </div>
      <div className="modal-action sticky bottom-0 bg-base-100">
        <button className="btn btn-sm" onClick={handleCancel}>
          Cancel
        </button>
        {!isRunning && (
          <button className="btn btn-sm" onClick={handleStart}>
            Start
          </button>
        )}
        {isRunning && (
          <button
            className="btn btn-sm"
            disabled={!controller}
            onClick={handleStopAndSave}
          >
            Stop & Save
          </button>
        )}
      </div>
    </div>
  );
};
