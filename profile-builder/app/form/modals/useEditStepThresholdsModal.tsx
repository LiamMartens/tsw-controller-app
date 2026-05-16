import { createPortal } from "react-dom";
import z from "zod";
import { useId, useMemo, useCallback } from "react";
import { useForm, useFieldArray } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { BaseField } from "../inputs";
import { t } from "../../utils";
import clsx from "clsx";

const stepThresholdSchema = z.object({
  threshold: z.object({
    reference: z.string(),
    value: z.coerce.number(),
  }),
  threshold_end: z
    .object({
      reference: z.string(),
      value: z.coerce.number(),
    })
    .optional(),
  threshold_tolerance: z.coerce.number().optional(),
});

type StepThresholdSchema = z.infer<typeof stepThresholdSchema>;

type Props = {
  value: unknown;
  onChange: (value: StepThresholdSchema[]) => void;
};

type ModalProps = Props & {
  dialogId: string;
  value: StepThresholdSchema[];
  onChange: (value: StepThresholdSchema[]) => void;
};

const EditStepThresholdsModalComponent = ({
  dialogId,
  value,
  onChange,
}: ModalProps) => {
  const form = useForm<{ step_thresholds: StepThresholdSchema[] }>({
    mode: "onChange",
    resolver: zodResolver(
      z.object({
        step_thresholds: z.array(stepThresholdSchema),
      }),
    ),
    defaultValues: { step_thresholds: value },
  });

  const { fields, append, remove } = useFieldArray({
    control: form.control,
    name: "step_thresholds",
  });

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.stopPropagation();

    form.handleSubmit(
      (data) => onChange(data.step_thresholds ?? []),
      () => e.preventDefault(),
    )();
  };

  return (
    <dialog id={dialogId} className="modal">
      <div className="modal-box w-11/12 max-w-3xl">
        <h3 className="font-bold text-lg">{t("Edit step thresholds")}</h3>
        <div className="space-y-2 py-4">
          <div className="space-y-1">
            {fields.map((field, index) => (
              <div
                key={field.id}
                className="flex gap-4 items-start border-b border-base-300 last:border-0"
              >
                <div className="flex-1 grid grid-cols-3 gap-2">
                  <BaseField
                    legend={t("Threshold Reference")}
                    error={
                      form.formState.errors.step_thresholds?.[index]?.threshold
                        ?.reference?.message
                    }
                    label={t(
                      "(Optional) can only be used to refer to named thresholds",
                    )}
                  >
                    <input
                      className={clsx(
                        "input input-bordered w-full",
                        form.formState.errors.step_thresholds?.[index]
                          ?.threshold?.reference && "input-error",
                      )}
                      placeholder={t("Reference")}
                      {...form.register(
                        `step_thresholds.${index}.threshold.reference` as const,
                      )}
                    />
                  </BaseField>

                  <BaseField
                    legend={t("Threshold Value")}
                    error={
                      form.formState.errors.step_thresholds?.[index]?.threshold
                        ?.value?.message
                    }
                  >
                    <input
                      type="number"
                      className={clsx(
                        "input input-bordered w-full",
                        form.formState.errors.step_thresholds?.[index]
                          ?.threshold?.value && "input-error",
                      )}
                      placeholder={t("Value")}
                      {...form.register(
                        `step_thresholds.${index}.threshold.value` as const,
                        { valueAsNumber: true },
                      )}
                    />
                  </BaseField>

                  <BaseField
                    legend={t("Threshold End Reference")}
                    error={
                      form.formState.errors.step_thresholds?.[index]
                        ?.threshold_end?.reference?.message
                    }
                    label={t(
                      "(Optional) can only be used to refer to named thresholds",
                    )}
                  >
                    <input
                      className={clsx(
                        "input input-bordered w-full",
                        form.formState.errors.step_thresholds?.[index]
                          ?.threshold_end?.reference && "input-error",
                      )}
                      placeholder={t("End reference")}
                      {...form.register(
                        `step_thresholds.${index}.threshold_end.reference` as const,
                      )}
                    />
                  </BaseField>

                  <BaseField
                    legend={t("Threshold End Value")}
                    error={
                      form.formState.errors.step_thresholds?.[index]
                        ?.threshold_end?.value?.message
                    }
                  >
                    <input
                      type="number"
                      className={clsx(
                        "input input-bordered w-full",
                        form.formState.errors.step_thresholds?.[index]
                          ?.threshold_end?.value && "input-error",
                      )}
                      placeholder={t("End value")}
                      {...form.register(
                        `step_thresholds.${index}.threshold_end.value` as const,
                        {
                          valueAsNumber: true,
                        },
                      )}
                    />
                  </BaseField>

                  <BaseField
                    legend={t("Threshold Tolerance")}
                    error={
                      form.formState.errors.step_thresholds?.[index]
                        ?.threshold_tolerance?.message
                    }
                  >
                    <input
                      type="number"
                      step="0.01"
                      className={clsx(
                        "input input-bordered w-full",
                        form.formState.errors.step_thresholds?.[index]
                          ?.threshold_tolerance && "input-error",
                      )}
                      placeholder={t("Tolerance")}
                      {...form.register(
                        `step_thresholds.${index}.threshold_tolerance` as const,
                        {
                          valueAsNumber: true,
                        },
                      )}
                    />
                  </BaseField>
                </div>
                <button
                  type="button"
                  className="btn btn-error btn-sm btn-ghost mt-9"
                  onClick={() => remove(index)}
                >
                  {t("Remove")}
                </button>
              </div>
            ))}
          </div>

          <div className="modal-action">
            <button
              type="button"
              className="btn btn-sm mr-auto"
              onClick={() =>
                append({
                  threshold: { reference: "", value: 0 },
                  threshold_end: { reference: "", value: 0 },
                  threshold_tolerance: undefined,
                })
              }
            >
              {t("Add step threshold")}
            </button>

            <form
              method="dialog"
              className="flex gap-2"
              onSubmit={handleSubmit}
            >
              <button
                type="button"
                className="btn btn-sm"
                onClick={() => {
                  const dialog = document.getElementById(
                    dialogId,
                  ) as HTMLDialogElement;
                  dialog?.close();
                }}
              >
                {t("Cancel")}
              </button>
              <button type="submit" className="btn btn-sm btn-primary">
                {t("Save")}
              </button>
            </form>
          </div>
        </div>
      </div>
    </dialog>
  );
};

export const useEditStepThresholdsModal = ({ value, onChange }: Props) => {
  const id = useId();
  const dialogId = useMemo(
    () => `edit-step-thresholds-dialog-` + id.replace(/[^\w]+/g, "-"),
    [id],
  );

  const safeValue = useMemo((): StepThresholdSchema[] => {
    const safe = z.array(stepThresholdSchema).safeParse(value);
    return safe.success ? safe.data : [];
  }, [value]);

  const open = useCallback(() => {
    (document.getElementById(dialogId) as HTMLDialogElement).showModal();
  }, [dialogId]);

  return {
    open,
    render: () =>
      createPortal(
        <EditStepThresholdsModalComponent
          dialogId={dialogId}
          value={safeValue}
          onChange={onChange}
        />,
        document.body,
      ),
  };
};
