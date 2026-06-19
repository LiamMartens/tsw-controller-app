import { useCallback } from "react";
import { main } from "../../../wailsjs/go/models";
import { TiCog, TiPin } from "react-icons/ti";
import clsx from "clsx";

type Props = {
  isPinned: boolean;
  controlState: main.Interop_Cab_ControlState_Control;
  onMapControl: (controlState: main.Interop_Cab_ControlState_Control) => void;
  onTogglePinControl: (control: main.Interop_Cab_ControlState_Control) => void;
};

export const CabDebuggerTabControl = ({
  isPinned,
  controlState,
  onMapControl,
  onTogglePinControl,
}: Props) => {
  const handleMapControl = useCallback(() => {
    onMapControl(controlState);
  }, [controlState, onMapControl]);

  const handleTogglePin = useCallback(() => {
    onTogglePinControl(controlState);
  }, [controlState, onMapControl]);

  return (
    <li className="list-row">
      <div className="list-col-grow grid gap-2 grid-cols-[minmax(0,1fr)_max-content]">
        <div className="grid gap-2 grid-cols-2 grid-flow-row auto-rows-max">
          <div>
            <p className="text-slate-400">Sync Control Name</p>
            <p>{decodeURIComponent(controlState.Identifier)}</p>
          </div>
          <div>
            <p className="text-slate-400">Direct Control Name</p>
            <p>{decodeURIComponent(controlState.PropertyName)}</p>
          </div>
          <div>
            <p className="text-slate-400">Current Value</p>
            <p>{controlState.CurrentValue.toFixed(4)}</p>
          </div>
          <div>
            <p className="text-slate-400">Current Normalized Value</p>
            <p>{controlState.CurrentNormalizedValue.toFixed(4)}</p>
          </div>
        </div>
        <div>
          <button
            role="button"
            className={clsx("btn btn-xs py-4 px-2", isPinned && "btn-primary")}
            onClick={handleTogglePin}
          >
            <TiPin size={20} />
          </button>
          <div className="dropdown dropdown-bottom dropdown-end">
            <div tabIndex={0} role="button" className="btn btn-xs py-4 px-2">
              <TiCog size={20} />
            </div>
            <ul
              tabIndex={-1}
              className="dropdown-content menu bg-base-300 rounded-box z-1 w-52 p-2 shadow-sm"
            >
              <li>
                <button onClick={handleMapControl}>Map Control</button>
              </li>
            </ul>
          </div>
        </div>
      </div>
    </li>
  );
};
