import { UseFormReturn } from "react-hook-form";
import { main } from "../../../../wailsjs/go/models";
import { MapControlFormValues } from "./mapControlForm";
import { ChangeEvent, useCallback, useEffect, useState } from "react";
import {
  GetCabControlState,
  SubscribeChangeEvent,
  UnsubscribeChangeEvent,
} from "../../../../wailsjs/go/main/App";
import { EventsOn } from "../../../../wailsjs/runtime/runtime";
import { events } from "../../../events";
import clsx from "clsx";

type MultiSwitchOrLeverConfigurationValueBindingProps = {
  controlName: main.Interop_Cab_ControlState_Control["PropertyName"];
  form: UseFormReturn<MapControlFormValues>;
  index: number;
};

export const CabDebuggerMapControlModalContent_MultiSwitchOrLeverConfiguration_ValueBinding =
  ({
    form,
    index,
    controlName,
  }: MultiSwitchOrLeverConfigurationValueBindingProps) => {
    const [listening, setListening] = useState(false);
    const [currentCabControlValue, setCurrentCabControlValue] = useState(0);
    const [currentBindControlValue, setCurrentBindControlValue] = useState(0);

    const controlType = form.watch("controlType");
    const bindingOptions = form.watch("bindingOptions");
    const currentValue = bindingOptions.values[index];

    const handleFreeRangeCheckedChange = useCallback(
      (event: ChangeEvent<HTMLInputElement>) => {
        const bindingOptions = form.getValues("bindingOptions");
        const updatedValues = [...bindingOptions.values];
        updatedValues[index].freerange = event.currentTarget.checked;
        form.setValue("bindingOptions", { values: updatedValues });
      },
      [index],
    );

    const handleDeleteValue = useCallback(() => {
      const bindingOptions = form.getValues("bindingOptions");
      const updatedValues = [...bindingOptions.values];
      updatedValues.splice(index, 1);
      form.setValue("bindingOptions", { values: updatedValues });
    }, [index]);

    const handleToggleBindValue = useCallback(() => {
      setListening((v) => {
        if (v) {
          const bindingOptions = form.getValues("bindingOptions");
          const updatedValues = [...bindingOptions.values];
          updatedValues[index] = {
            value: currentCabControlValue,
            mapping: currentBindControlValue,
            freerange: false,
          };
          updatedValues.sort((a, b) => (a.value < b.value ? -1 : 1));
          form.setValue(
            "bindingOptions",
            { values: updatedValues },
            { shouldValidate: true, shouldDirty: true, shouldTouch: true },
          );
        }
        return !v;
      });
    }, [index, form, currentCabControlValue, currentBindControlValue]);

    useEffect(() => {
      if (listening) {
        SubscribeChangeEvent();
        return () => {
          UnsubscribeChangeEvent();
        };
      }
    }, [listening]);

    useEffect(() => {
      const unsbuscribe = EventsOn(
        events.changeevent,
        (data: main.Interop_ChangeEvent) => {
          const binding = form.getValues("binding");
          if (data.ControlName === binding) {
            setCurrentBindControlValue(data.Value);
          }
        },
      );
      const cabControlInterval = setInterval(() => {
        GetCabControlState()
          .then((cs) => {
            const control = cs.Controls.find(
              (c) => c.PropertyName === controlName,
            );
            setCurrentCabControlValue(control?.CurrentValue ?? 0);
          })
          .catch(() => {});
      }, 200);

      return () => {
        clearInterval(cabControlInterval);
        unsbuscribe();
      };
    }, [form, controlName]);

    return (
      <>
        {controlType === "lever" && index > 0 && (
          <div className="divider my-4">
            <label className="label text-sm">
              <input
                type="checkbox"
                checked={!!currentValue.freerange}
                className="checkbox checkbox-sm"
                onChange={handleFreeRangeCheckedChange}
              />
              Enable free range zone
            </label>
          </div>
        )}
        <div className="flex flex-row items-center gap-2 text-sm text-base-content">
          <div>
            Game Value:{" "}
            <span className="kbd">
              {listening
                ? currentCabControlValue.toFixed(2)
                : currentValue.value.toFixed(2)}
            </span>
            , Controller Value:{" "}
            <span className="kbd">
              {listening
                ? currentBindControlValue.toFixed(2)
                : currentValue.mapping.toFixed(2)}
            </span>
          </div>
          <div className="join ml-auto">
            <button
              type="button"
              className="join-item btn btn-sm btn-ghost btn-error"
              onClick={handleDeleteValue}
            >
              Delete
            </button>
            <button
              type="button"
              className={clsx("join-item btn btn-sm", {
                "btn-active btn-info": listening,
              })}
              onClick={handleToggleBindValue}
            >
              Bind Value
            </button>
          </div>
        </div>
      </>
    );
  };
