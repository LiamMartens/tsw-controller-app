import { useEffect, useMemo, useRef, useState } from "react";
import { main } from "../../../wailsjs/go/models";
import { CalibrationModal } from "./CalibrationModal";
import { EventsOn } from "../../../wailsjs/runtime/runtime";
import { events } from "../../events";
import { useControllers } from "../../swr";
import { RemapNotchesModal } from "./RemapNotchesModal";

type ControllerRowProps = {
  controller: main.Interop_GenericController;
  onConfigure: (controller: main.Interop_GenericController) => void;
  onRemapNotches: (controller: main.Interop_GenericController) => void;
};

const CalibrationTabControllerRow = ({
  controller,
  onConfigure,
  onRemapNotches,
}: ControllerRowProps) => {
  const handleConfigure = () => {
    onConfigure(controller);
  };

  const handleRemapNotches = () => {
    onRemapNotches(controller);
  };

  return (
    <li className="list-row">
      <div className="list-col-grow">
        <div>{controller.Name}</div>
      </div>
      <div className="flex items-center gap-2">
        {controller.IsConfigured && controller.HasThresholds && (
          <div
            className="tooltip tooltip-bottom"
            data-tip="Remap controller notches"
          >
            <button
              className="btn btn-success btn-soft btn-xs"
              onClick={handleRemapNotches}
            >
              Remap Notches
            </button>
          </div>
        )}
        {controller.IsConfigured && (
          <div className="tooltip tooltip-bottom" data-tip="Re-configure">
            <button
              className="btn btn-success btn-soft btn-xs"
              onClick={handleConfigure}
            >
              Configured
            </button>
          </div>
        )}
        {!controller.IsConfigured && (
          <div className="tooltip tooltip-bottom" data-tip="Configure now">
            <button
              className="btn btn-error btn-soft btn-xs"
              onClick={handleConfigure}
            >
              Unconfigured
            </button>
          </div>
        )}
      </div>
    </li>
  );
};

export const CalibrationTab = () => {
  const calibrationDialogRef = useRef<HTMLDialogElement | null>(null);
  const remapNotchesDialogRef = useRef<HTMLDialogElement | null>(null);
  const [currentlyCalibratingController, setCurrentlyCalibratingController] =
    useState<main.Interop_GenericController | null>(null);
  const [
    currrentlyRemappingNotchesController,
    setCurrrentlyRemappingNotchesController,
  ] = useState<main.Interop_GenericController | null>(null);

  const { data: controllers, mutate: refetchControllers } = useControllers();
  const configurableControllers = useMemo(
    () => controllers.filter((c) => !c.IsVirtual),
    [controllers],
  );

  const handleConfigure = (c: main.Interop_GenericController) => {
    setCurrentlyCalibratingController(c);
    calibrationDialogRef.current?.showModal();
  };

  const handleRemapNotches = (c: main.Interop_GenericController) => {
    setCurrrentlyRemappingNotchesController(c);
    remapNotchesDialogRef.current?.showModal();
  };

  useEffect(() => {
    return EventsOn(events.joydevices_updated, () => {
      refetchControllers();
    });
  }, []);

  return (
    <div>
      <ul className="list bg-base-100 rounded-box shadow-md">
        {configurableControllers?.map((c) => (
          <CalibrationTabControllerRow
            key={c.UniqueID}
            controller={c}
            onConfigure={handleConfigure}
            onRemapNotches={handleRemapNotches}
          />
        ))}
      </ul>
      <CalibrationModal
        controller={currentlyCalibratingController}
        onClose={() => setCurrentlyCalibratingController(null)}
      />
      <RemapNotchesModal
        controller={currrentlyRemappingNotchesController}
        onClose={() => setCurrrentlyRemappingNotchesController(null)}
      />
    </div>
  );
};
