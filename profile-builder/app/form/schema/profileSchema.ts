import z from "zod";
import { t } from "../../utils";

export const railClassInformationSchema = z.object({
  class_name: z.string().min(1, t("Class name is required")),
});

export const controllerInformationSchema = z.object({
  usb_id: z.string().optional(),
  mapping: z.object({}).passthrough().optional(),
  calibration: z.object({}).passthrough().optional(),
});

export const profileControlSchema = z.object({
  name: z.string().min(1, t("Control name is required")),
  assignments: z.array(
    z
      .object({
        type: z.enum([
          "momentary",
          "toggle",
          "linear",
          "direct_control",
          "api_control",
          "sync_control",
        ]),
      })
      .passthrough(),
  ),
});

export const profileSchema = z.object({
  name: z.string().min(1, t("Profile name is required")),
  extends: z.string().optional(),
  auto_select: z.boolean().optional(),
  controls: z.array(profileControlSchema),
  controller: controllerInformationSchema.optional(),
  rail_class_information: z.array(railClassInformationSchema).optional(),
});
