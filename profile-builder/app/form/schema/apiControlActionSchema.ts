import z from "zod";

export const apiControlActionSchema = z.object({
  controls: z.string(),
  api_value: z.number(),
  hold: z.boolean().optional(),
  max_change_rate: z.number().optional(),
});
