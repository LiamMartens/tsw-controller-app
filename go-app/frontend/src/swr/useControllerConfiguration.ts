import useSWR from "swr";
import { GetControllerConfiguration } from "../../wailsjs/go/main/App";
import { main } from "../../wailsjs/go/models";

export const useControllerConfiguration = (controller:  main.Interop_GenericController) => {
  return useSWR(
    ["system", "controller", controller.UniqueID, "configuration"],
    async () => GetControllerConfiguration(controller.UniqueID),
    { suspense: true, revalidateOnMount: true },
  );
};
