import { useForm } from "react-hook-form";
import { main } from "../../../../wailsjs/go/models";
import {
  Modal,
  ModalCloseReason,
  ModalContentProps,
} from "../../../components";
import { BaseSyntheticEvent, useCallback, useEffect, useState } from "react";
import {
  DirectLikeInputValue,
  ProfileSchema,
} from "../../../profile-schema/schema";
import {
  CabDebuggerProfileAssignmentSaveModal,
  ProfileSavedModalCloseReason,
} from "../ProfileAssignmentSaveModal/CabDebuggerProfileAssignmentSaveModal";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  mapKeybindingFormSchema,
  MapKeybindingFormValues,
} from "./mapKeybindingForm";
import clsx from "clsx";
import {
  SubscribeChangeEvent,
  UnsubscribeChangeEvent,
} from "../../../../wailsjs/go/main/App";
import { EventsOn } from "../../../../wailsjs/runtime/runtime";
import { events } from "../../../events";
// import { CabDebuggerMapControlModalContent_SimpleButtonOrSwitchConfiguration } from "./CabDebuggerMapControlModalContent_SimpleButtonOrSwitchConfiguration";
// import { CabDebuggerMapControlModalContent_MultiSwitchOrLeverConfiguration } from "./CabDebuggerMapControlModalContent_MultiSwitchOrLeverConfiguration";

const FORM_ID = "CabDebuggerMapKeybindingForm";

type Props = {
  openState: boolean;
  onClose: () => void;
};

type ContentProps = ModalContentProps<boolean>;

const CabDebuggerMapKeybindingModalContent = ({ onClose }: ContentProps) => {
  const [saveProfileOpenState, setSaveProfileOpenState] = useState<
    ProfileSchema["controls"][number] | null
  >(null);

  const form = useForm<MapKeybindingFormValues>({
    mode: "all",
    resolver: zodResolver(mapKeybindingFormSchema),
    defaultValues: {
      keys: "",
      binding: "",
    },
  });

  const [listening, setListening] = useState(false);
  const binding = form.watch("binding");

  const handleFormValid = async (
    values: MapKeybindingFormValues,
    event?: BaseSyntheticEvent,
  ) => {
    event?.stopPropagation();
    const control: ProfileSchema["controls"][number] = {
      name: values.binding,
      assignments: [
        {
          type: "momentary",
          threshold: 0.9,
          action_activate: {
            keys: values.keys,
          },
        },
      ],
    };
    setSaveProfileOpenState(control);
  };

  const handleBindToggle = useCallback(() => {
    setListening((v) => !v);
  }, []);

  const handleCloseSaveModal = (reason?: ModalCloseReason) => {
    setSaveProfileOpenState(null);
    if (reason && reason instanceof ProfileSavedModalCloseReason) {
      onClose();
    }
  };

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
    <>
      <div className="flex flex-col gap-2">
        <h3 className="font-bold text-base">Map keybinding</h3>
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
          <div className="flex flex-col gap-2 py-2 px-4 border bg-base-200 border-base-300 shadow-md rounded-md">
            <input
              type="text"
              className="input w-full"
              placeholder="eg: ctrl+a"
              {...form.register("keys", { required: true })}
            />
            <div className="p-2 flex justify-between gap-2">
              {!binding && (
                <p className="text-sm text-base-content/70">
                  No binding yet...
                </p>
              )}
              {!!binding && (
                <p className="text-sm text-base-content">{binding}</p>
              )}
              <button
                type="button"
                className={clsx("btn btn-sm", {
                  "btn-active btn-info": listening,
                })}
                onClick={handleBindToggle}
              >
                Bind
              </button>
            </div>
          </div>
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

export const CabDebuggerMapKeybindingModal = ({
  openState,
  onClose,
}: Props) => {
  return (
    <Modal
      openState={openState}
      onClose={onClose}
      Component={CabDebuggerMapKeybindingModalContent}
    />
  );
};
