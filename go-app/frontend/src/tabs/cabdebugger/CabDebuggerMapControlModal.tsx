import { Controller, useForm, UseFormReturn } from "react-hook-form";
import { main } from "../../../wailsjs/go/models";
import { Modal, ModalCloseReason, ModalContentProps } from "../../components";
import { ChangeEvent, useCallback, useEffect, useState } from "react";
import clsx from "clsx";
import {
  GetCabControlState,
  SubscribeChangeEvent,
  UnsubscribeChangeEvent,
} from "../../../wailsjs/go/main/App";
import { EventsOn } from "../../../wailsjs/runtime/runtime";
import { events } from "../../events";
import {
  DirectLikeInputValue,
  ProfileSchema,
} from "../../profile-schema/schema";
import {
  CabDebuggerMapControlSaveModal,
  ProfileSavedModalCloseReason,
} from "./CabDebuggerMapControlSaveModal";
import z from "zod";
import { zodResolver } from "@hookform/resolvers/zod";

type Props = {
  controlState: main.Interop_Cab_ControlState_Control["PropertyName"] | null;
  onClose: () => void;
};

type ContentProps = ModalContentProps<
  main.Interop_Cab_ControlState_Control["PropertyName"]
>;

type SimpleButtonOrSwitchConfigurationProps = {
  form: UseFormReturn<ContentFormValues>;
};

type MultiSwitchOrLeverConfigurationProps = {
  controlName: main.Interop_Cab_ControlState_Control["PropertyName"];
  form: UseFormReturn<ContentFormValues>;
};

type MultiSwitchOrLeverConfigurationValueBindingProps = {
  controlName: main.Interop_Cab_ControlState_Control["PropertyName"];
  form: UseFormReturn<ContentFormValues>;
  index: number;
};
type ContentFormValues = z.infer<typeof formSchema>;
const formSchema = z.object({
  controlType: z.enum(["button", "switch/simple", "switch/multi", "lever"]),
  binding: z.string().min(1, "Binding is required"),
  bindingOptions: z.object({
    values: z.array(
      z.object({
        value: z.number(),
        mapping: z.number(),
        freerange: z.boolean(),
      }),
    ),
  }),
});

const CabDebuggerMapControlModalContent_SimpleButtonOrSwitchConfiguration = ({
  form,
}: SimpleButtonOrSwitchConfigurationProps) => {
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
        className={clsx("btn btn-sm", { "btn-active btn-info": listening })}
        onClick={handleBindToggle}
      >
        Bind
      </button>
    </div>
  );
};

const CabDebuggerMapControlModalContent_MultiSwitchOrLeverConfiguration_ValueBinding =
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
              className="join-item btn btn-sm btn-ghost btn-error"
              onClick={handleDeleteValue}
            >
              Delete
            </button>
            <button
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

const CabDebuggerMapControlModalContent_MultiSwitchOrLeverConfiguration = ({
  form,
  controlName,
}: MultiSwitchOrLeverConfigurationProps) => {
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
      return EventsOn(events.changeevent, (data: main.Interop_ChangeEvent) => {
        form.setValue("binding", data.ControlName, {
          shouldDirty: true,
          shouldTouch: true,
          shouldValidate: true,
        });
        setListening(false);
      });
    }
  }, [form, listening]);

  return (
    <div className="flex flex-col gap-2 py-2 px-4 border bg-base-200 border-base-300 shadow-md rounded-md">
      <div className="grid items-center grid-cols-[minmax(0,1fr)_max-content] gap-2">
        <div>
          {!binding && (
            <p className="text-sm text-base-content/70">No binding yet...</p>
          )}
          {!!binding && <p className="text-sm text-base-content">{binding}</p>}
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
            positions. Once ready, click "Bind Value" again to save the values.
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

const CabDebuggerMapControlModalContent = ({
  openState: controlName,
  onClose,
}: ContentProps) => {
  const [saveProfileOpenState, setSaveProfileOpenState] = useState<
    ProfileSchema["controls"][number] | null
  >(null);
  const form = useForm<ContentFormValues>({
    mode: "all",
    resolver: zodResolver(formSchema),
    defaultValues: {
      controlType: "button",
      binding: "",
      bindingOptions: {
        values: [],
      },
    },
  });
  const controlType = form.watch("controlType");

  const handleSave = async () => {
    await form.handleSubmit((values) => {
      if (
        values.controlType === "button" ||
        values.controlType === "switch/simple"
      ) {
        const control: ProfileSchema["controls"][number] = {
          name: values.binding,
          assignments: [
            {
              type: "direct_control",
              hold: true,
              controls: controlName,
              conditions: [
                { control: values.binding, operator: "gte", value: 0.9 },
              ],
              input_value: {
                min: 1,
                max: 1,
              },
            },
            {
              type: "direct_control",
              controls: controlName,
              conditions: [
                { control: values.binding, operator: "lt", value: 0.9 },
              ],
              input_value: {
                min: 0,
                max: 0,
              },
            },
            {
              type: "api_control",
              hold: true,
              controls: controlName,
              conditions: [
                { control: values.binding, operator: "gte", value: 0.9 },
              ],
              input_value: {
                min: 1,
                max: 1,
              },
            },
            {
              type: "api_control",
              controls: controlName,
              conditions: [
                { control: values.binding, operator: "lt", value: 0.9 },
              ],
              input_value: {
                min: 0,
                max: 0,
              },
            },
          ],
        };
        setSaveProfileOpenState(control);
      } else if (
        values.controlType === "lever" ||
        values.controlType === "switch/multi"
      ) {
        const sortedValues = values.bindingOptions.values.toSorted((a, b) =>
          a.value < b.value ? -1 : 1,
        );
        const minValue = sortedValues[0]?.value ?? 0;
        const maxValue = sortedValues[sortedValues.length - 1]?.value ?? 1;
        const inputValue: DirectLikeInputValue = {
          min: minValue,
          max: maxValue,
        };
        if (sortedValues.length > 0) {
          inputValue.steps = sortedValues.reduce<(number | null)[]>(
            (steps, val, index) => {
              if (index > 0 && val.freerange) steps.push(null);
              steps.push(val.value);
              return steps;
            },
            [],
          );
          inputValue.step_thresholds = sortedValues.map((srv, index, self) => {
            if (index > 0 && srv.freerange) {
              return {
                threshold: self[index - 1].value,
                threshold_end: srv.value,
              };
            }
            return { threshold: srv.mapping };
          });
        }

        const control: ProfileSchema["controls"][number] = {
          name: values.binding,
          assignments: [
            {
              type: "direct_control",
              controls: controlName,
              input_value: inputValue,
            },
            {
              type: "api_control",
              controls: controlName,
              input_value: inputValue,
            },
          ],
        };
        setSaveProfileOpenState(control);
      }
    })();
  };

  const handleCloseSaveModal = (reason?: ModalCloseReason) => {
    setSaveProfileOpenState(null);
    if (reason && reason instanceof ProfileSavedModalCloseReason) {
      onClose();
    }
  };

  return (
    <>
      <div className="flex flex-col gap-2">
        <h3 className="font-bold text-base">Mapping {controlName}</h3>
        <div role="alert" className="alert alert-info alert-soft">
          <span>
            This functionality is currently intended for setting up initial
            mappings and does not allow editing existing or creating complex
            mappings.
          </span>
        </div>
        <div className="flex flex-col gap-2">
          <fieldset className="fieldset">
            <legend className="fieldset-legend">Control Type</legend>
            <select className="select w-full" {...form.register("controlType")}>
              <option value="button">Simple Button</option>
              <option value="switch/simple">Simple Switch (On/Off)</option>
              <option value="switch/multi">Multi-Value Switch</option>
              <option value="lever">Multi-Value Lever</option>
            </select>
          </fieldset>
          {(controlType === "button" || controlType === "switch/simple") && (
            <CabDebuggerMapControlModalContent_SimpleButtonOrSwitchConfiguration
              form={form}
            />
          )}
          {(controlType === "switch/multi" || controlType === "lever") && (
            <CabDebuggerMapControlModalContent_MultiSwitchOrLeverConfiguration
              form={form}
              controlName={controlName}
            />
          )}
        </div>
        <div className="flex flex-row gap-2 items-center justify-end">
          <form method="dialog">
            <button className="btn btn-sm">Cancel</button>
          </form>
          <button
            className="btn btn-sm"
            disabled={!form.formState.isValid}
            onClick={handleSave}
          >
            Save
          </button>
        </div>
      </div>
      <CabDebuggerMapControlSaveModal
        openState={saveProfileOpenState}
        onClose={handleCloseSaveModal}
      />
    </>
  );
};

export const CabDebuggerMapControlModal = ({
  controlState,
  onClose,
}: Props) => {
  return (
    <Modal
      openState={controlState ?? false}
      onClose={onClose}
      Component={CabDebuggerMapControlModalContent}
    />
  );
};
