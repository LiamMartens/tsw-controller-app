import { UseFormReturn } from "react-hook-form";
import { main } from "../../../../wailsjs/go/models";
import { MapControlFormValues } from "./mapControlForm";
import { useCallback, useEffect, useState } from "react";
import {
  SubscribeChangeEvent,
  UnsubscribeChangeEvent,
} from "../../../../wailsjs/go/main/App";
import { EventsOn } from "../../../../wailsjs/runtime/runtime";
import { events } from "../../../events";
import clsx from "clsx";
import { CabDebuggerMapControlModalContent_MultiSwitchOrLeverConfiguration_ValueBinding } from "./CabDebuggerMapControlModalContent_MultiSwitchOrLeverConfiguration_ValueBinding";

type MultiSwitchOrLeverConfigurationProps = {
  controlName: main.Interop_Cab_ControlState_Control["PropertyName"];
  form: UseFormReturn<MapControlFormValues>;
};

export const CabDebuggerMapControlModalContent_MultiSwitchOrLeverConfiguration =
  ({ form, controlName }: MultiSwitchOrLeverConfigurationProps) => {
    const [listening, setListening] = useState(false);

    const binding = form.watch("binding");
    const bindingOptions = form.watch("bindingOptions");

    const handleBindToggle = useCallback(() => {
      setListening((v) => !v);
    }, []);

    const handleAddValue = useCallback(() => {
      const bindingOptions = form.getValues("bindingOptions");
      form.setValue("bindingOptions", {
        values: [
          ...bindingOptions.values,
          bindingOptions.values[bindingOptions.values.length - 1] ?? {
            value: 0,
            mapping: 0,
            freerange: false,
          },
        ],
      });
    }, []);

    useEffect(() => {
      if (listening) {
        SubscribeChangeEvent();
        return () => {
          UnsubscribeChangeEvent();
        };
      }
    }, [listening]);

    useEffect(() => {
      if (listening) {
        return EventsOn(
          events.changeevent,
          (data: main.Interop_ChangeEvent) => {
            form.setValue("binding", data.ControlName, {
              shouldDirty: true,
              shouldTouch: true,
              shouldValidate: true,
            });
            setListening(false);
          },
        );
      }
    }, [form, listening]);

    return (
      <div className="flex flex-col gap-2 py-2 px-4 border bg-base-200 border-base-300 shadow-md rounded-md">
        <div className="grid items-center grid-cols-[minmax(0,1fr)_max-content] gap-2">
          <div>
            {!binding && (
              <p className="text-sm text-base-content/70">No binding yet...</p>
            )}
            {!!binding && (
              <p className="text-sm text-base-content">{binding}</p>
            )}
          </div>
          <div className="join">
            <button
              className={clsx("join-item btn btn-sm", {
                "btn-active btn-info": listening,
              })}
              onClick={handleBindToggle}
            >
              Bind
            </button>
            <button className="join-item btn btn-sm" onClick={handleAddValue}>
              Add value
            </button>
          </div>
        </div>
        {!!bindingOptions.values.length && (
          <>
            <div className="divider my-0"></div>
            <div className="alert alert-soft">
              To bind a value, click the "Bind Value" button and move both the
              in-game control and the physical control to their respective
              positions. Once ready, click "Bind Value" again to save the
              values.
            </div>
            {bindingOptions.values.map(({ value, mapping }, index) => (
              <CabDebuggerMapControlModalContent_MultiSwitchOrLeverConfiguration_ValueBinding
                key={`value_${value}_${mapping}_${index}`}
                form={form}
                index={index}
                controlName={controlName}
              />
            ))}
          </>
        )}
      </div>
    );
  };
