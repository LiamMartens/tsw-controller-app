import z from "zod";

export type MapControlFormValues = z.infer<typeof mapControlFormSchema>;
export const mapControlFormSchema = z.object({
  controlType: z.enum(["button", "switch/simple", "switch/multi", "lever"]),
  binding: z.string().min(1, "Binding is required"),
  bindingOptions: z.object({
    values: z.array(
      z.object({
        value: z.number(),
        mapping: z.number(),
        freerange: z.boolean(),
      }),
    ),
  }),
});
