import z from "zod";
import { t } from "../../utils";

export const keysActionchema = z.object({
  keys: z.string(),
  press_time: z
    .number()
    .positive(t("The press time must be a positive number"))
    .optional(),
  wait_time: z
    .number()
    .min(0, t("The wait time can not be negative"))
    .optional(),
});
