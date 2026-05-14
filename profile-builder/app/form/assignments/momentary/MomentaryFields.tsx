import { Controller, useForm, UseFormReturn } from "react-hook-form";
import { ActionSelector } from "../../actions/ActionSelector";
import z from "zod";
import { profileSchema } from "../../schema";
import { BaseField, FieldGroup } from "../../inputs";
import { t } from "../../../utils";
import clsx from "clsx";
import { useEffect, useMemo, useRef } from "react";
import { zodResolver } from "@hookform/resolvers/zod";

type Props = {
  form: UseFormReturn<z.infer<typeof profileSchema>>;
  controlIndex: number;
  assignmentIndex: number;
};

type ContentProps = {
  value: unknown;
  onChange: (v: unknown) => void;
};

const momentarySchema = z
  .object({
    type: z.literal("momentary"),
    threshold: z.coerce.number(),
    match: z.enum(["exceeds", "equals"]).optional(),
    action_activate: z.unknown(),
    action_deactivate: z.unknown(),
  })
  .passthrough();

const numberSchema = z.number().or(z.any().transform(() => 0));

const MomentaryFieldsContent = ({ value, onChange }: ContentProps) => {
  const safeValue = useMemo((): z.infer<typeof momentarySchema> => {
    const validated = momentarySchema.safeParse(value);
    return validated.success
      ? validated.data
      : {
          type: "momentary",
          threshold: 0.9,
          match: "exceeds",
          action_activate: { keys: "" },
        };
  }, [value]);

  const form = useForm<z.infer<typeof momentarySchema>>({
    mode: "onChange",
    resolver: zodResolver(momentarySchema),
    defaultValues: safeValue,
  });

  const actionActivate = form.watch("action_activate");
  const actionDeactivate = form.watch("action_deactivate");

  const onChangeRef = useRef(onChange);
  onChangeRef.current = onChange;
  useEffect(() => {
    form.watch(() => form.handleSubmit(onChange)());
  }, [form, onChangeRef]);

  return (
    <>
      <div className="grid grid-cols-2 gap-4">
        <BaseField
          legend={t("Threshold")}
          label={t("The threshold to surpass before activating the action")}
          error={form.formState.errors.threshold?.message}
        >
          <input
            className={clsx(
              "input input-bordered w-full",
              form.formState.errors.threshold && "input-error",
            )}
            {...form.register("threshold", { valueAsNumber: true })}
          />
        </BaseField>
        <BaseField
          legend={t("Match")}
          label={t(
            "Defines how to match the threshold value (defaults to exceeds)",
          )}
          error={form.formState.errors.match?.message}
        >
          <select
            className={clsx(
              "select w-full",
              form.formState.errors.match && "select-error",
            )}
            {...form.register("match")}
          >
            <option value="exceeds">{t("Exceeds")}</option>
            <option value="equals">{t("Exact match")}</option>
          </select>
        </BaseField>
      </div>

      <FieldGroup
        legend={t("Action Activate")}
        label={t(
          "This is the action that will activate when the control matches/exceeds the threshold",
        )}
      >
        <ActionSelector
          value={actionActivate}
          onChange={(v) =>
            form.setValue("action_activate", v, {
              shouldDirty: true,
              shouldTouch: true,
            })
          }
        />
      </FieldGroup>

      <FieldGroup
        legend={t("Action De-activate")}
        label={t(
          "This is the action that will trigger when the control is returned to the normal position. Defaults to releasing the keys if not provided.",
        )}
      >
        <label className="label mb-4">
          <input
            type="checkbox"
            checked={!!actionDeactivate}
            className="checkbox"
            onChange={(e) =>
              form.setValue(
                "action_deactivate",
                e.currentTarget.checked ? { keys: "" } : undefined,
                { shouldDirty: true, shouldTouch: true },
              )
            }
          />
          {t("Enable de-activation action")}
        </label>

        {!!actionDeactivate && (
          <ActionSelector
            value={actionDeactivate}
            onChange={(v) =>
              form.setValue("action_deactivate", v, {
                shouldDirty: true,
                shouldTouch: true,
              })
            }
          />
        )}
      </FieldGroup>
    </>
  );
};

export const MomentaryFields = ({
  form,
  controlIndex,
  assignmentIndex,
}: Props) => {
  return (
    <Controller
      control={form.control}
      name={`controls.${controlIndex}.assignments.${assignmentIndex}`}
      render={({ field }) => (
        <MomentaryFieldsContent value={field.value} onChange={field.onChange} />
      )}
    />
  );
};
