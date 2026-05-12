import { Controller, useFieldArray, UseFormReturn } from "react-hook-form";
import z from "zod";
import { profileSchema } from "./schema";
import { t } from "../utils";
import { clsx } from "clsx";
import { BaseField, JsonTextareaInput } from "./inputs";

type Props = {
  form: UseFormReturn<z.infer<typeof profileSchema>>;
};

const ProfileHeaderRailClassInformationSection = ({ form }: Props) => {
  const {
    fields: railFields,
    append: appendRail,
    remove: removeRail,
  } = useFieldArray({
    control: form.control,
    name: "rail_class_information",
  });

  return (
    <div className="collapse collapse-arrow bg-base-100 border-base-300 border">
      <input type="checkbox" className="peer" />
      <div className="collapse-title font-semibold">
        {t("Supported Rail Classes")}
      </div>
      <div className="collapse-content">
        <div className="flex flex-col gap-2">
          {railFields.map((field, index) => (
            <div key={field.id} className="flex flex-col gap-2">
              <div className="flex gap-3 items-start">
                <BaseField
                  className="grow"
                  legend={t("USB ID")}
                  error={form.formState.errors.controller?.usb_id?.message}
                >
                  <input
                    type="text"
                    placeholder="Rail class name"
                    className="input input-bordered w-full"
                    {...form.register(
                      `rail_class_information.${index}.class_name`,
                    )}
                  />
                </BaseField>
                <button
                  type="button"
                  className="btn btn-ghost btn-error btn-sm mt-9"
                  onClick={() => removeRail(index)}
                >
                  {t("Remove")}
                </button>
              </div>
              <p className="label whitespace-normal text-error">
                {
                  form.formState.errors.rail_class_information?.[index]
                    ?.class_name?.message
                }
              </p>
            </div>
          ))}
          <button
            type="button"
            className="btn btn-sm"
            onClick={() => appendRail({ class_name: "" })}
          >
            {t("Add Rail Class")}
          </button>
        </div>
      </div>
    </div>
  );
};

const ProfileHeaderControllerInformationSection = ({ form }: Props) => {
  return (
    <div className="collapse collapse-arrow bg-base-100 border-base-300 border">
      <input type="checkbox" className="peer" />
      <div className="collapse-title font-semibold">
        {t("Controller Information")}
      </div>
      <div className="collapse-content">
        <div>
          <BaseField
            legend={t("USB ID")}
            label={t(
              "(Optional) defines the supported controller USB ID for this profile",
            )}
            error={form.formState.errors.controller?.usb_id?.message}
          >
            <input
              className={clsx(
                "input input-bordered w-full",
                form.formState.errors.controller?.usb_id && "input-error",
              )}
              placeholder={t("VID:PID")}
              {...form.register("name", { required: true })}
            />
          </BaseField>

          <Controller
            control={form.control}
            name="controller.mapping"
            render={({ field, fieldState }) => (
              <BaseField
                legend={t("SDL Mapping")}
                label={t(
                  "(Optional) embeds the SDL mapping into the profile for easier sharing",
                )}
                error={fieldState.error?.message}
              >
                <JsonTextareaInput
                  value={field.value}
                  onChange={field.onChange}
                />
              </BaseField>
            )}
          />

          <Controller
            control={form.control}
            name="controller.calibration"
            render={({ field, fieldState }) => (
              <BaseField
                legend={t("Calibration")}
                label={t(
                  "(Optional) embeds the controller calibration data into the profile for easy sharing",
                )}
                error={fieldState.error?.message}
              >
                <JsonTextareaInput
                  value={field.value}
                  onChange={field.onChange}
                />
              </BaseField>
            )}
          />
        </div>
      </div>
    </div>
  );
};

export const ProfileHeader = ({ form }: Props) => {
  return (
    <div className="space-y-4">
      <div className="flex flex-col gap-4">
        <fieldset className="fieldset">
          <legend className="fieldset-legend">{t("Profile Name")}</legend>
          <input
            className={clsx(
              "input input-bordered  w-full",
              form.formState.errors.name && "input-error",
            )}
            {...form.register("name", { required: true })}
          />
          {form.formState.errors.name && (
            <p className="label text-error">
              {form.formState.errors.name.message}
            </p>
          )}
        </fieldset>

        <BaseField
          legend={t("Extends")}
          label={t(
            '(Optional) Specifies the "name" of the profile to extend from. When extending a profile all existing controls will be inherited. Controls defined in this profile will be used as an override of the existing profile',
          )}
          error={form.formState.errors.extends?.message}
        >
          <input
            className={clsx(
              "input input-bordered  w-full",
              form.formState.errors.extends && "input-error",
            )}
            {...form.register("extends")}
          />
        </BaseField>

        <fieldset className="flex flex-col gap-2">
          <label className="label">
            <input
              type="checkbox"
              className="checkbox"
              {...form.register("auto_select")}
            />
            {t("Enable auto-select")}
          </label>
          <p className="label whitespace-normal">
            {t(
              "Enables automatically selecting the profile based on the supported controller and/or rail class information",
            )}
          </p>
        </fieldset>
      </div>

      <ProfileHeaderRailClassInformationSection form={form} />

      <ProfileHeaderControllerInformationSection form={form} />
    </div>
  );
};
