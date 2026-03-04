import useSWR from "swr";
import { GetCabControlState } from "../../wailsjs/go/main/App";
import { useEffect } from "react";

type Input = {
  refreshInterval: number
}

export const useCabControlState = ({ refreshInterval }: Input) => {
  const state = useSWR(
    ["system", "cabControlState"],
    async () => GetCabControlState(),
    { suspense: true, revalidateOnMount: true },
  );

  useEffect(() => {
    let interval: ReturnType<typeof setInterval> | null = null;
    interval = setInterval(() => {
      state.mutate();
    }, refreshInterval);
    return () => {
      if (interval) clearInterval(interval);
    };
  }, [state.mutate, refreshInterval]);

  return state
};
