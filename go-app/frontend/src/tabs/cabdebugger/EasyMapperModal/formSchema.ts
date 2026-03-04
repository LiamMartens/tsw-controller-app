import z from "zod";

export const formControlButtonSchema = z.object({
  control: z.string().min(1, "Control selection is required"),
  type: z.literal("button"),
  binding: z.object({
    name: z.string(),
  }),
});

export const formSwitchSimpleSchema = z.object({
  control: z.string().min(1, "Control selection is required"),
  type: z.literal("switch/simple"),
  binding: z.object({
    name: z.string(),
  }),
});

export const formSwitchMultiSchema = z.object({
  control: z.string().min(1, "Control selection is required"),
  type: z.literal("switch/multi"),
  binding: z.object({
    name: z.string(),
    states: z.array(
      z.object({
        value: z.number(),
        threshold: z.number(),
      }),
    ),
  }),
});

export const formLeverSchema = z.object({
  control: z.string().min(1, "Control selection is required"),
  type: z.literal("lever"),
  binding: z.object({
    name: z.string(),
    notches: z.array(
      z.object({
        value: z.number(),
        threshold: z.number(),
        freerange: z.enum(["start", "end"]),
      }),
    ),
  }),
});

export const formControlSchema = z.discriminatedUnion("type", [
  formControlButtonSchema,
  formSwitchSimpleSchema,
  formSwitchMultiSchema,
  formLeverSchema,
]);

export const formSchema = z.object({
  controller: z.string().min(1, "Controller selection is required"),
  controls: z.array(formControlSchema),
});
