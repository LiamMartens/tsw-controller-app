import { useCallback, useEffect, useMemo, useState } from "react";
import { useForm } from "react-hook-form";
import { useCabControlState } from "../../swr";
import { CabDebuggerTabControl } from "./CabDebuggerTabControl";
import { main } from "../../../wailsjs/go/models";
import { CabDebuggerMapControlModal } from "./MapControlModal/CabDebuggerMapControlModal";

export const CabDebuggerTab = () => {
  const { register, watch } = useForm<{ query: string }>({
    defaultValues: { query: "" },
  });
  const { data: cabControlState, mutate: refetchCabControlState } =
    useCabControlState();
  const [mapControlModalOpenState, setMapControlModalOpenState] =
    useState<main.Interop_Cab_ControlState_Control | null>(null);

  const query = watch("query");
  const sortedControls = useMemo(
    () =>
      cabControlState?.Controls.filter((c) =>
        [c.Identifier, c.PropertyName].some((t) =>
          t.toLowerCase().includes(query.toLowerCase()),
        ),
      ).sort((a, b) =>
        `${a.PropertyName}_${a.Identifier}`.localeCompare(
          `${b.PropertyName}_${b.Identifier}`,
        ),
      ),
    [cabControlState?.Controls, query],
  );

  const handleOpenMapControlModal = useCallback(
    (controlState: main.Interop_Cab_ControlState_Control) => {
      setMapControlModalOpenState(controlState);
    },
    [],
  );

  const handleCloseMapControlModal = useCallback(() => {
    setMapControlModalOpenState(null);
  }, []);

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
      {!cabControlState?.Controls?.length && (
        <div className="py-12 text-center">
          <p className="text-base-content/50 text-sm">
            Waiting for cab state...
          </p>
        </div>
      )}
      {!!cabControlState?.Controls?.length && (
        <div>
          <input
            className="input w-full"
            placeholder="Search for control(s)"
            {...register("query")}
          />
        </div>
      )}
      <ul className="list bg-base-100 rounded-box shadow-md">
        {sortedControls?.map((controlState) => (
          <CabDebuggerTabControl
            key={controlState.PropertyName}
            controlState={controlState}
            onMapControl={handleOpenMapControlModal}
          />
        ))}
      </ul>

      <CabDebuggerMapControlModal
        controlState={mapControlModalOpenState?.PropertyName ?? null}
        onClose={handleCloseMapControlModal}
      />
    </div>
  );
};
