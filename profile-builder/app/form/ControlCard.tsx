import { UseFormReturn } from "react-hook-form";
import { AssignmentsList } from "./assignments/AssignmentsList";
import z from "zod";
import { profileSchema } from "./schema";
import { t } from "../utils";
import { BaseField } from "./inputs";
import clsx from "clsx";
import { useConfirmModal } from "./modals/useConfirmModal";

type Props = {
  form: UseFormReturn<z.infer<typeof profileSchema>>;
  index: number;
  onRemove: () => void;
};

export const ControlCard = ({ form, index, onRemove }: Props) => {
  const { confirm, render: ConfirmDeleteComponent } = useConfirmModal({
    title: t("Are you sure?"),
    body: t("Are you sure you want to remove this control?"),
    onConfirm: () => onRemove(),
  });

  const control = form.watch(`controls.${index}.name`);
  const title = control || `${t("Control")} ${index + 1}`;

  const handleRemove = () => {
    confirm();
  };

  return (
    <div className="collapse collapse-arrow bg-base-100 border border-base-300 rounded-lg">
      <input type="checkbox" className="peer" />
      <div className="collapse-title flex justify-between items-center">
        <span className="font-medium">{title}</span>
      </div>

      <div className="collapse-content space-y-4">
        <BaseField
          legend={t("Controller Control Name")}
          label={t("This is the physical control name as mapped")}
          error={form.formState.errors.controls?.[index]?.name?.message}
        >
          <input
            {...form.register(`controls.${index}.name`)}
            type="text"
            className={clsx(
              "input input-bordered w-full",
              form.formState.errors.controls?.[index]?.name && "input-error",
            )}
          />
        </BaseField>

        <AssignmentsList controlIndex={index} form={form} />

        <div className="flex justify-start border-t border-base-300 pt-4">
          <button
            type="button"
            className="btn btn-error btn-xs btn-ghost"
            onClick={handleRemove}
          >
            {t("Remove Control")}
          </button>
        </div>
      </div>

      <ConfirmDeleteComponent />
    </div>
  );
};
