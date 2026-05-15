import { createPortal } from "react-dom";
import z from "zod";
import { CONDITION_OPERATORS } from "../utils";
import {
  useId,
  useMemo,
  useRef,
  useEffect,
  SubmitEvent,
  useCallback,
} from "react";
import { useForm, useFieldArray } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { BaseField } from "../inputs";
import { t } from "../../utils";
import clsx from "clsx";

const conditionSchema = z.object({
  control: z.string().min(1, t("Control name is required")),
  operator: z.enum(
    CONDITION_OPERATORS.map((c) => c.value) as [Operator, ...Operator[]],
  ),
  value: z.coerce.number(),
});

type Operator = (typeof CONDITION_OPERATORS)[number]["value"];

type ConditionSchema = z.infer<typeof conditionSchema>;

type Props = {
  value: unknown;
  onChange: (value: ConditionSchema[]) => void;
};

type ModalProps = Props & {
  dialogId: string;
  value: ConditionSchema[];
  onChange: (value: ConditionSchema[]) => void;
};

const EditConditionsModalComponent = ({
  dialogId,
  value,
  onChange,
}: ModalProps) => {
  const form = useForm<{ conditions: ConditionSchema[] }>({
    mode: "onChange",
    resolver: zodResolver(
      z.object({
        conditions: z.array(conditionSchema),
      }),
    ),
    defaultValues: { conditions: value },
  });

  const { fields, append, remove } = useFieldArray({
    control: form.control,
    name: "conditions",
  });

  const handleSubmit = (e: SubmitEvent<HTMLFormElement>) => {
    e.stopPropagation();

    const submitter = e.nativeEvent.submitter as
      | HTMLInputElement
      | HTMLButtonElement
      | null;

    if (submitter?.value === "cancel") {
      form.reset();
      return;
    }

    form.handleSubmit(
      (data) => onChange(data.conditions ?? []),
      () => e.preventDefault(),
    )();
  };

  useEffect(() => form.reset({ conditions: value }), [form, value]);

  return (
    <dialog id={dialogId} className="modal">
      <div className="modal-box w-11/12 max-w-3xl">
        <h3 className="font-bold text-lg">{t("Edit conditions")}</h3>
        <div className="space-y-2 py-4">
          <div className="space-y-1">
            {fields.map((field, index) => (
              <div
                key={field.id}
                className="flex gap-2 items-start border-b border-base-300 last:border-0"
              >
                <div className="flex-1 grid grid-cols-3 gap-2">
                  <BaseField
                    legend={t("Control")}
                    error={
                      form.formState.errors.conditions?.[index]?.control
                        ?.message
                    }
                  >
                    <input
                      className={clsx(
                        "input input-bordered w-full",
                        form.formState.errors.conditions?.[index]?.control &&
                          "input-error",
                      )}
                      placeholder={t("Control name")}
                      {...form.register(`conditions.${index}.control` as const)}
                    />
                  </BaseField>

                  <BaseField
                    legend={t("Operator")}
                    error={
                      form.formState.errors.conditions?.[index]?.operator
                        ?.message
                    }
                  >
                    <select
                      className={clsx(
                        "select select-bordered w-full",
                        form.formState.errors.conditions?.[index]?.operator &&
                          "select-error",
                      )}
                      {...form.register(
                        `conditions.${index}.operator` as const,
                      )}
                    >
                      {CONDITION_OPERATORS.map((op) => (
                        <option key={op.value} value={op.value}>
                          {op.label}
                        </option>
                      ))}
                    </select>
                  </BaseField>

                  <BaseField
                    legend={t("Value")}
                    error={
                      form.formState.errors.conditions?.[index]?.value?.message
                    }
                  >
                    <input
                      type="number"
                      className={clsx(
                        "input input-bordered w-full",
                        form.formState.errors.conditions?.[index]?.value &&
                          "input-error",
                      )}
                      placeholder={t("Value")}
                      {...form.register(`conditions.${index}.value` as const, {
                        valueAsNumber: true,
                      })}
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
                  control: "",
                  operator: "gte" as const,
                  value: 0,
                })
              }
            >
              {t("Add condition")}
            </button>

            <form
              method="dialog"
              className="flex gap-2"
              onSubmit={handleSubmit}
            >
              <button
                name="action"
                value="cancel"
                type="submit"
                className="btn btn-sm"
              >
                {t("Cancel")}
              </button>
              <button
                type="submit"
                name="action"
                value="save"
                className="btn btn-sm btn-primary"
              >
                {t("Save")}
              </button>
            </form>
          </div>
        </div>
      </div>
    </dialog>
  );
};

export const useEditConditionsModal = ({ value, onChange }: Props) => {
  const id = useId();
  const dialogId = useMemo(
    () => `edit-conditions-dialog-` + id.replace(/[^\w]+/g, "-"),
    [id],
  );

  const safeValue = useMemo((): ConditionSchema[] => {
    const safe = z.array(conditionSchema).safeParse(value);
    return safe.success ? safe.data : [];
  }, [value]);

  const open = useCallback(() => {
    (document.getElementById(dialogId) as HTMLDialogElement).showModal();
  }, [dialogId]);

  return {
    open,
    render: () =>
      createPortal(
        <EditConditionsModalComponent
          dialogId={dialogId}
          value={safeValue}
          onChange={onChange}
        />,
        document.body,
      ),
  };
};
