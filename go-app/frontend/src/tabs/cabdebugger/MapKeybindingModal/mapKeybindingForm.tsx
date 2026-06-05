import z from "zod";

export type MapKeybindingFormValues = z.infer<typeof mapKeybindingFormSchema>;
export const mapKeybindingFormSchema = z.object({
  keys: z.string().min(1, "Keys definition is required"),
  binding: z.string().min(1, "Binding is required"),
});
