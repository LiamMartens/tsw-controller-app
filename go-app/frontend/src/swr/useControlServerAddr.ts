import useSWR from "swr";
import { GetControlServerAddr } from "../../wailsjs/go/main/App";

export const useControlServerAddr = () => {
  return useSWR(["system", "controlServerAddr"], async () => GetControlServerAddr(), {
    suspense: true,
    revalidateOnMount: true,
  });
};
