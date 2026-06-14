import useSWR from "swr";
import { Environment } from "../../wailsjs/runtime/runtime";

export const useEnvironment = () => {
  return useSWR(["system", "environmnt"], async () => Environment(), {
    suspense: true,
    revalidateOnMount: true,
  });
};
