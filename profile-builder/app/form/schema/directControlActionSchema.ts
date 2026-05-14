import z from "zod";

export const directControlActionSchema = z.object({
  controls: z.string(),
  value: z.number(),
  max_change_rate: z.number().optional(),
  relative: z.boolean().optional(),
  hold: z.boolean().optional(),
  use_normalized: z.boolean().optional(),
  notify: z.boolean().optional(),
  enable_api_fallback: z.boolean().optional(),
});
