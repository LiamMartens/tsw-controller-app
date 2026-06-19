import {
  MutableRefObject,
  useCallback,
  useEffect,
  useMemo,
  useState,
} from "react";
import { useForm } from "react-hook-form";
import { useCabControlState } from "../../swr";
import { CabDebuggerTabControl } from "./CabDebuggerTabControl";
import { main } from "../../../wailsjs/go/models";
import { CabDebuggerMapControlModal } from "./MapControlModal/CabDebuggerMapControlModal";
import { CabDebuggerMapKeybindingModal } from "./MapKeybindingModal/CabDebuggerMapKeybindingModal";
import { TiCog } from "react-icons/ti";

const pinnedDebuggerControlsRef: MutableRefObject<Set<string>> = {
  current: new Set(),
};

export const CabDebuggerTab = () => {
  const { register, watch } = useForm<{ query: string }>({
    defaultValues: { query: "" },
  });
  const [pinnedControls, setPinnedControls] = useState(
    pinnedDebuggerControlsRef.current,
  );
  const { data: cabControlState, mutate: refetchCabControlState } =
    useCabControlState();
  const [mapControlModalOpenState, setMapControlModalOpenState] =
    useState<main.Interop_Cab_ControlState_Control | null>(null);
  const [mapKeybindingModalOpen, setMapKeybindingModalOpen] = useState(false);

  const query = watch("query");
  const filteredAndSortedControls = useMemo(() => {
    if (!cabControlState?.Controls.length) return [];

    type ControlType = (typeof cabControlState)["Controls"][number];
    const sanitizedQuery = query.trim().toLowerCase();
    const filterFunc = (c: ControlType) =>
      [c.Identifier, c.PropertyName].some((t) =>
        t.toLowerCase().includes(sanitizedQuery),
      );

    const filtered = sanitizedQuery.length
      ? cabControlState.Controls.filter(filterFunc)
      : cabControlState.Controls;

    const sorted = filtered.sort((a, b) => {
      const isAPinned = pinnedControls.has(a.PropertyName);
      const isBPinned = pinnedControls.has(b.PropertyName);
      if (isAPinned && !isBPinned) return -1;
      if (isBPinned && !isAPinned) return 1;
      return `${a.PropertyName}_${a.Identifier}`.localeCompare(
        `${b.PropertyName}_${b.Identifier}`,
      );
    });

    return sorted;
  }, [cabControlState?.Controls, query, pinnedControls]);

  const handleOpenMapControlModal = useCallback(
    (controlState: main.Interop_Cab_ControlState_Control) => {
      setMapControlModalOpenState(controlState);
    },
    [],
  );

  const handleTogglePinControl = useCallback(
    (controlState: main.Interop_Cab_ControlState_Control) => {
      setPinnedControls((pinned) => {
        pinned.add(controlState.PropertyName);
        return new Set(pinned);
      });
    },
    [],
  );

  const handleCloseMapControlModal = useCallback(() => {
    setMapControlModalOpenState(null);
  }, []);

  const handleCloseMapKeybindingModal = useCallback(() => {
    setMapKeybindingModalOpen(false);
  }, []);

  useEffect(() => {
    pinnedDebuggerControlsRef.current = pinnedControls;
  }, [pinnedControls]);

  useEffect(() => {
    let interval: ReturnType<typeof setInterval> | null = null;
    interval = setInterval(() => {
      refetchCabControlState();
    }, 100);
    return () => {
      if (interval) clearInterval(interval);
    };
  }, [refetchCabControlState]);

  return (
    <div className="p-4 grid grid-cols-1 grid-flow-row auto-rows-max gap-4">
      {!!cabControlState?.Name && (
        <div className="alert alert-soft alert-info">
          <div>Currently driving {cabControlState.Name}</div>
        </div>
      )}
      <div className="grid items-center grid-cols-[minmax(0,1fr)_max-content] gap-2">
        <input
          className="input w-full"
          placeholder="Search for control(s)"
          {...register("query")}
        />
        <div>
          <div className="dropdown dropdown-bottom dropdown-end">
            <div tabIndex={0} role="button" className="btn py-4 px-2">
              <TiCog size={20} />
            </div>
            <ul
              tabIndex={-1}
              className="dropdown-content menu bg-base-300 rounded-box z-1 w-52 p-2 shadow-sm"
            >
              <li>
                <button onClick={() => setMapKeybindingModalOpen(true)}>
                  Map keybinding
                </button>
              </li>
            </ul>
          </div>
        </div>
      </div>

      {!cabControlState?.Controls?.length && (
        <div className="py-12 text-center">
          <p className="text-base-content/50 text-sm">
            Waiting for cab state...
          </p>
        </div>
      )}

      <ul className="list bg-base-100 rounded-box shadow-md">
        {filteredAndSortedControls?.map((controlState) => (
          <CabDebuggerTabControl
            key={controlState.PropertyName}
            isPinned={pinnedControls.has(controlState.PropertyName)}
            controlState={controlState}
            onMapControl={handleOpenMapControlModal}
            onTogglePinControl={handleTogglePinControl}
          />
        ))}
      </ul>

      <CabDebuggerMapControlModal
        controlState={mapControlModalOpenState?.PropertyName ?? null}
        onClose={handleCloseMapControlModal}
      />

      <CabDebuggerMapKeybindingModal
        openState={mapKeybindingModalOpen}
        onClose={handleCloseMapKeybindingModal}
      />
    </div>
  );
};
