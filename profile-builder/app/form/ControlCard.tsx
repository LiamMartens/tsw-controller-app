import { UseFormReturn } from "react-hook-form";
import { AssignmentsList } from "./assignments/AssignmentsList";
import z from "zod";
import { profileSchema } from "./schema";
import { t } from "../utils";
import { BaseField } from "./inputs";
import clsx from "clsx";

type Props = {
  form: UseFormReturn<z.infer<typeof profileSchema>>;
  index: number;
  onRemove: () => void;
};

export const ControlCard = ({ form, index, onRemove }: Props) => {
  const control = form.watch(`controls.${index}.name`);
  const title = control || `${t("Control")} ${index + 1}`;

  const handleRemove = () => {
    onRemove();
  };

  return (
    <div className="collapse collapse-arrow bg-base-200 rounded-lg">
      <input type="checkbox" className="peer" />
      <div className="collapse-title flex justify-between items-center">
        <span className="font-medium">{title}</span>
        <button
          type="button"
          className="btn btn-error btn-xs btn-ghost"
          onClick={handleRemove}
        >
          {t("Remove")}
        </button>
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
      </div>
    </div>
  );
};
