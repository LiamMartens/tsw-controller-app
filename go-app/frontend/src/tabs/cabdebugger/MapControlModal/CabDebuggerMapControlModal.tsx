import { useForm } from "react-hook-form";
import { main } from "../../../../wailsjs/go/models";
import {
  Modal,
  ModalCloseReason,
  ModalContentProps,
} from "../../../components";
import { BaseSyntheticEvent, useState } from "react";
import {
  DirectLikeInputValue,
  ProfileSchema,
} from "../../../profile-schema/schema";
import {
  CabDebuggerProfileAssignmentSaveModal,
  ProfileSavedModalCloseReason,
} from "../ProfileAssignmentSaveModal/CabDebuggerProfileAssignmentSaveModal";
import { zodResolver } from "@hookform/resolvers/zod";
import { mapControlFormSchema, MapControlFormValues } from "./mapControlForm";
import { CabDebuggerMapControlModalContent_SimpleButtonOrSwitchConfiguration } from "./CabDebuggerMapControlModalContent_SimpleButtonOrSwitchConfiguration";
import { CabDebuggerMapControlModalContent_MultiSwitchOrLeverConfiguration } from "./CabDebuggerMapControlModalContent_MultiSwitchOrLeverConfiguration";

const FORM_ID = "CabDebuggerMapControlForm";

type Props = {
  controlState: main.Interop_Cab_ControlState_Control["PropertyName"] | null;
  onClose: () => void;
};

type ContentProps = ModalContentProps<
  main.Interop_Cab_ControlState_Control["PropertyName"]
>;

const CabDebuggerMapControlModalContent = ({
  openState: controlName,
  onClose,
}: ContentProps) => {
  const [saveProfileOpenState, setSaveProfileOpenState] = useState<
    ProfileSchema["controls"][number] | null
  >(null);
  const form = useForm<MapControlFormValues>({
    mode: "all",
    resolver: zodResolver(mapControlFormSchema),
    defaultValues: {
      controlType: "button",
      binding: "",
      bindingOptions: {
        values: [],
      },
    },
  });
  const controlType = form.watch("controlType");

  const handleFormValid = async (
    values: MapControlFormValues,
    event?: BaseSyntheticEvent,
  ) => {
    event?.stopPropagation();
    if (
      values.controlType === "button" ||
      values.controlType === "switch/simple"
    ) {
      const control: ProfileSchema["controls"][number] = {
        name: values.binding,
        assignments: [
          {
            type: "momentary",
            threshold: 0.9,
            action_activate: {
              controls: controlName,
              enable_api_fallback: true,
              value: 1.0,
              hold: true,
            },
            action_deactivate: {
              controls: controlName,
              enable_api_fallback: true,
              value: 0.0,
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
            enable_api_fallback: true,
          },
        ],
      };
      setSaveProfileOpenState(control);
    }
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
        <form
          id={FORM_ID}
          className="flex flex-col gap-2"
          onSubmit={form.handleSubmit(handleFormValid)}
        >
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
        </form>
        <div className="flex flex-row gap-2 items-center justify-end">
          <form method="dialog">
            <button className="btn btn-sm">Cancel</button>
          </form>
          <button
            type="submit"
            form={FORM_ID}
            className="btn btn-sm"
            disabled={!form.formState.isValid}
          >
            Save
          </button>
        </div>
      </div>
      <CabDebuggerProfileAssignmentSaveModal
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
