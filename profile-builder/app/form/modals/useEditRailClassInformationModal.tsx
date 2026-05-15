import { createPortal } from "react-dom";
import z from "zod";
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

const railClassSchema = z.object({
  class_name: z.string().min(1, t("Class name is required")),
});

type RailClassSchema = z.infer<typeof railClassSchema>;

type Props = {
  value: unknown;
  onChange: (value: RailClassSchema[]) => void;
};

type ModalProps = Props & {
  dialogId: string;
  value: RailClassSchema[];
  onChange: (value: RailClassSchema[]) => void;
};

const EditRailClassInformationModalComponent = ({
  dialogId,
  value,
  onChange,
}: ModalProps) => {
  const form = useForm<{ rail_class_information: RailClassSchema[] }>({
    mode: "onChange",
    resolver: zodResolver(
      z.object({
        rail_class_information: z.array(railClassSchema),
      }),
    ),
    defaultValues: { rail_class_information: value },
  });

  const { fields, append, remove } = useFieldArray({
    control: form.control,
    name: "rail_class_information",
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
      (data) => onChange(data.rail_class_information ?? []),
      () => e.preventDefault(),
    )();
  };

  useEffect(() => form.reset({ rail_class_information: value }), [form, value]);

  return (
    <dialog id={dialogId} className="modal">
      <div className="modal-box w-11/12 max-w-3xl">
        <h3 className="font-bold text-lg">
          {t("Edit rail class information")}
        </h3>
        <div className="space-y-2 py-4">
          <div className="space-y-1">
            {fields.map((field, index) => (
              <div
                key={field.id}
                className="flex gap-4 items-start border-b border-base-300 last:border-0"
              >
                <div className="flex-1 grid grid-cols-1 gap-2">
                  <BaseField
                    legend={t("Class Name")}
                    error={
                      form.formState.errors.rail_class_information?.[index]
                        ?.class_name?.message
                    }
                  >
                    <input
                      className={clsx(
                        "input input-bordered w-full",
                        form.formState.errors.rail_class_information?.[index]
                          ?.class_name && "input-error",
                      )}
                      placeholder={t("Class name")}
                      {...form.register(
                        `rail_class_information.${index}.class_name` as const,
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
              onClick={() => append({ class_name: "" })}
            >
              {t("Add rail class")}
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

export const useEditRailClassInformationModal = ({
  value,
  onChange,
}: Props) => {
  const id = useId();
  const dialogId = useMemo(
    () => `edit-rail-class-dialog-` + id.replace(/[^\w]+/g, "-"),
    [id],
  );

  const safeValue = useMemo((): RailClassSchema[] => {
    const safe = z.array(railClassSchema).safeParse(value);
    return safe.success ? safe.data : [];
  }, [value]);

  const open = useCallback(() => {
    (document.getElementById(dialogId) as HTMLDialogElement).showModal();
  }, [dialogId]);

  return {
    open,
    render: () =>
      createPortal(
        <EditRailClassInformationModalComponent
          dialogId={dialogId}
          value={safeValue}
          onChange={onChange}
        />,
        document.body,
      ),
  };
};
