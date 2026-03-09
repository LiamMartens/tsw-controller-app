import { UseFormReturn } from "react-hook-form";
import { MapControlFormValues } from "./mapControlForm";
import { useCallback, useEffect, useState } from "react";
import {
  SubscribeChangeEvent,
  UnsubscribeChangeEvent,
} from "../../../../wailsjs/go/main/App";
import { EventsOn } from "../../../../wailsjs/runtime/runtime";
import { events } from "../../../events";
import { main } from "../../../../wailsjs/go/models";
import clsx from "clsx";

type SimpleButtonOrSwitchConfigurationProps = {
  form: UseFormReturn<MapControlFormValues>;
};

export const CabDebuggerMapControlModalContent_SimpleButtonOrSwitchConfiguration =
  ({ form }: SimpleButtonOrSwitchConfigurationProps) => {
    const [listening, setListening] = useState(false);
    const binding = form.watch("binding");

    const handleBindToggle = useCallback(() => {
      setListening((v) => !v);
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
      return EventsOn(events.changeevent, (data: main.Interop_ChangeEvent) => {
        form.setValue("binding", data.ControlName, {
          shouldDirty: true,
          shouldTouch: true,
          shouldValidate: true,
        });
        setListening(false);
      });
    }, [form]);

    return (
      <div className="grid items-center grid-cols-[minmax(0,1fr)_max-content] gap-2 py-2 px-4 border bg-base-200 border-base-300 shadow-md rounded-md">
        <div>
          {!binding && (
            <p className="text-sm text-base-content/70">No binding yet...</p>
          )}
          {!!binding && <p className="text-sm text-base-content">{binding}</p>}
        </div>
        <button
          type="button"
          className={clsx("btn btn-sm", { "btn-active btn-info": listening })}
          onClick={handleBindToggle}
        >
          Bind
        </button>
      </div>
    );
  };
