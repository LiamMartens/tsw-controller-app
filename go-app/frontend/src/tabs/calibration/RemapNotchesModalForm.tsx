import { useEffect, useMemo, useState } from "react";
import { config, main } from "../../../wailsjs/go/models";
import {
  SaveCalibration,
  LoadConfiguration,
  UnsubscribeChangeEvent,
  SubscribeChangeEvent,
} from "../../../wailsjs/go/main/App";
import { alert } from "../../utils/alert";
import { useControllerConfiguration } from "../../swr";
import { useForm } from "react-hook-form";
import { EventsOn } from "../../../wailsjs/runtime/runtime";
import { events } from "../../events";

type Props = {
  controller: main.Interop_GenericController;
  onClose: () => void;
};

export const RemapNotchesModalForm = ({ controller, onClose }: Props) => {
  const [bindingNamedThreshold, setBindingNamedThreshold] = useState<{
    control: string;
    threshold: string;
  } | null>(null);
  const {
    data: controllerConfiguration,
    mutate: updateControllerConfiguration,
  } = useControllerConfiguration(controller);

  const defaultFormThresholds = useMemo(() => {
    const thresholds: Record<string, Record<string, number>> = {};
    for (const control of controllerConfiguration?.Calibration.Controls ?? []) {
      thresholds[control.Name] = {};
      for (const threshold of control.Thresholds) {
        thresholds[control.Name][threshold.name] = threshold.value;
      }
    }
    return thresholds;
  }, [controllerConfiguration]);

  const form = useForm<{
    thresholds: Record<string, Record<string, number>>;
  }>({
    defaultValues: {
      thresholds: defaultFormThresholds,
    },
  });

  const handleCancel = () => {
    onClose();
  };

  const handleStartBinding = (control: string, threshold: string) => {
    setBindingNamedThreshold({ control, threshold });
  };

  const handleStopBinding = () => {
    setBindingNamedThreshold(null);
  };

  const handleStopAndSave = async () => {
    try {
      await UnsubscribeChangeEvent();
      await form.handleSubmit(async (values) => {
        const data = new main.Interop_ControllerCalibration();
        data.Name = controller.Name;
        data.DeviceID = controller.DeviceID;
        data.Controls = (
          controllerConfiguration?.Calibration.Controls ?? []
        ).map((control) => {
          return main.Interop_ControllerCalibration_Control.createFrom({
            Kind: control.Kind,
            Index: control.Index,
            Name: control.Name,
            Min: control.Min,
            Max: control.Max,
            Idle: control.Idle,
            Deadzone: control.Deadzone,
            Invert: control.Invert,
            EasingCurve: control.EasingCurve,
            Thresholds: Object.entries(values.thresholds[control.Name]).map(
              ([name, value]) =>
                config.Config_Controller_Calibration_Threshold.createFrom({
                  name,
                  value,
                }),
            ),
          });
        });
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
    if (bindingNamedThreshold) {
      SubscribeChangeEvent();
      const unsubscribe = EventsOn(
        events.changeevent,
        (event: main.Interop_ChangeEvent) => {
          if (event.UniqueID !== controller.UniqueID) return;
          if (event.ControlName === bindingNamedThreshold.control) {
            const thresholds = form.getValues("thresholds");
            thresholds[bindingNamedThreshold.control] =
              thresholds[bindingNamedThreshold.control] ?? {};
            thresholds[bindingNamedThreshold.control][
              bindingNamedThreshold.threshold
            ] = Math.round(event.Value * 1000) / 1000;
            form.setValue("thresholds", thresholds, {
              shouldDirty: true,
              shouldTouch: true,
            });
          }
        },
      );
      return () => {
        unsubscribe();
        UnsubscribeChangeEvent();
      };
    }

    return () => {
      /* force unsubscribe */
      UnsubscribeChangeEvent();
    };
  }, [bindingNamedThreshold, controller]);

  const thresholds = form.watch("thresholds");

  return (
    <div>
      <h3 className="font-bold text-base">
        Remapping Notches {controller.Name}
      </h3>
      <div className="mt-4 flex flex-col gap-2">
        {controllerConfiguration?.Calibration.Controls.flatMap((control) =>
          control.Thresholds.map((threshold) => {
            const isBinding =
              bindingNamedThreshold &&
              bindingNamedThreshold.control == control.Name &&
              bindingNamedThreshold.threshold == threshold.name;
            const value =
              thresholds?.[control.Name]?.[threshold.name] ?? threshold.value;

            return (
              <div
                key={`${control.Name}_${threshold.name}`}
                className="rounded-sm bg-base-200 p-2 flex items-center"
              >
                <div className="join">
                  <div className="badge badge-soft badge-info rounded-r-none">
                    {control.Name}
                  </div>
                  <div className="badge badge-info rounded-l-none">
                    {threshold.name}
                  </div>
                </div>
                <div className="flex gap-2 items-center ml-auto">
                  <div className="kbd kbd-sm">{value}</div>
                  <button
                    className="btn btn-xs btn-primary"
                    disabled={!!bindingNamedThreshold && !isBinding}
                    onClick={() => {
                      if (!isBinding)
                        handleStartBinding(control.Name, threshold.name);
                      else handleStopBinding();
                    }}
                  >
                    {isBinding ? "Set" : "Bind"}
                  </button>
                </div>
              </div>
            );
          }),
        )}
      </div>
      <div className="modal-action sticky bottom-0 bg-base-100">
        <button className="btn btn-sm" onClick={handleCancel}>
          Cancel
        </button>

        <button
          className="btn btn-sm"
          disabled={!controller}
          onClick={handleStopAndSave}
        >
          Stop & Save
        </button>
      </div>
    </div>
  );
};
