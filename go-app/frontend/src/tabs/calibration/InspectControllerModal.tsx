import { main } from "../../../wailsjs/go/models";
import { CalibrationModalForm } from "./CalibrationModalForm";
import { Modal, ModalContentProps } from "../../components";
import { useEffect, useMemo } from "react";
import {
  SubscribeChangeEvent,
  UnsubscribeChangeEvent,
} from "../../../wailsjs/go/main/App";
import { EventsOn } from "../../../wailsjs/runtime/runtime";
import { events } from "../../events";
import { useForm } from "react-hook-form";
import { useControllerConfiguration } from "../../swr";

type ControllerValuesForm = {
  values: Record<string, number>;
};

type Props = {
  controller: main.Interop_GenericController | null;
  onClose: () => void;
};

const InspectModalContent = ({
  openState: controller,
  onClose,
}: ModalContentProps<main.Interop_GenericController>) => {
  const { data: controllerConfiguration } =
    useControllerConfiguration(controller);
  const controls = useMemo(() =>
    controllerConfiguration?.SDLMapping.data.toSorted((a, b) =>
      a.name.localeCompare(b.name),
    ) ?? [],
    [controllerConfiguration?.SDLMapping.data]
  );

  const form = useForm<ControllerValuesForm>({
    defaultValues: {
      values: {},
    },
  });
  const values = form.watch("values");

  useEffect(() => {
    SubscribeChangeEvent();
    const unsubscribe = EventsOn(
      events.changeevent,
      (event: main.Interop_ChangeEvent) => {
        if (event.UniqueID !== controller.UniqueID) return;
        const values = form.getValues("values");
        form.setValue(
          "values",
          { ...values, [event.ControlName]: event.Value },
          { shouldTouch: true, shouldDirty: true },
        );
      },
    );

    return () => {
      unsubscribe();
      UnsubscribeChangeEvent();
    };
  }, [form, controller]);

  return (
    <div>
      <h3 className="font-bold text-base">Inspecting {controller?.Name} Values</h3>
      <div className="my-4 flex flex-col gap-2">
        {controls.map((control) => (
          <div key={control.name} className="flex justify-between items-center">
            <div>{control.name}</div>
            <div>
              <kbd className="kbd kbd-sm">
                {Object.hasOwn(values, control.name)
                  ? values[control.name].toFixed(2)
                  : "..."}
              </kbd>
            </div>
          </div>
        ))}
      </div>
      <div className="modal-action sticky bottom-0 bg-base-200 p-2 rounded-md">
        <button className="btn btn-sm" onClick={onClose}>
          Close
        </button>
      </div>
    </div>
  );
};

export const InspectControllerModal = ({ controller, onClose }: Props) => {
  return (
    <Modal
      openState={controller ?? false}
      onClose={onClose}
      Component={InspectModalContent}
    />
  );
};
