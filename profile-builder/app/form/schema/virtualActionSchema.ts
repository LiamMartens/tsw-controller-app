import z from "zod";

export const virtualActionSchema = z.object({
  type: z.literal("virtual"),
  control: z.string(),
  value: z.number(),
});
