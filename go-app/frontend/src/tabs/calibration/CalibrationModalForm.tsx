import { useEffect, useState } from "react";
import { main } from "../../../wailsjs/go/models";
import {
  SaveCalibration,
  UnsubscribeRaw,
  SubscribeRaw,
  GetControllerConfiguration,
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
import { s } from "framer-motion/client";

type Props = {
  controller: main.Interop_GenericController;
  onClose: () => void;
};

export const CalibrationModalForm = ({ controller, onClose }: Props) => {
  const [isRunning, setIsRunning] = useState(false);
  const { data: controllerConfiguration, mutate: updateControllerConfiguration } =
    useControllerConfiguration(controller);
  const form = useCalibrationForm({
    name: controllerConfiguration?.Calibration.Name ?? "",
    controls: (controllerConfiguration?.Calibration.Controls ?? []).map(
      (control): CalibrationStateControl => ({
        kind: control.Kind as Kind,
        index: control.Index,
        name: control.Name,
        min: control.Min,
        max: control.Max,
        idle: control.Idle,
        deadzone: control.Deadzone,
        invert: control.Invert,
        value: control.Idle,
        easingCurve: control.EasingCurve,
        override: true,
      }),
    ).toSorted((a, b) =>
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
      await form.handleSubmit(async (values) => {
        const data = new main.Interop_ControllerCalibration();
        data.Name = values.name;
        data.DeviceID = controller.DeviceID;
        data.Controls = values.controls.map((control) => ({
          Kind: control.kind,
          Index: control.index,
          Name: control.name,
          Min: control.min,
          Max: control.max,
          Idle: control.idle,
          Deadzone: control.deadzone,
          Invert: control.invert,
          EasingCurve: control.easingCurve,
        }));
        await SaveCalibration(data);
        await LoadConfiguration();
        await updateControllerConfiguration();
      })();
    } catch (err) {
      alert(`Could not save calibration (${err})`, "error");
    } finally {
      onClose();
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
