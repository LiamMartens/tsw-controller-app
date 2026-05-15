import { UseFormReturn } from "react-hook-form";
import { ASSIGNMENT_TYPES } from "../utils";
import { MomentaryFields } from "./momentary/MomentaryFields";
import z from "zod";
import { profileSchema } from "../schema";
import { BaseField } from "../inputs";
import { t } from "../../utils";

type Props = {
  form: UseFormReturn<z.infer<typeof profileSchema>>;
  controlIndex: number;
  assignmentIndex: number;
};

export const AssignmentTypeSelector = ({ form, controlIndex, assignmentIndex }: Props) => {
  const { watch } = form;
  const currentValue = watch(`controls.${controlIndex}.assignments.${assignmentIndex}`);

  return (
    <div className="space-y-3">
      <BaseField legend={t("Assignment Type")}>
        <select
          className="select select-bordered w-full"
          {...form.register(`controls.${controlIndex}.assignments.${assignmentIndex}.type`)}
        >
          {ASSIGNMENT_TYPES.map((t) => (
            <option key={t.value} value={t.value}>
              {t.label}
            </option>
          ))}
        </select>
      </BaseField>

      {currentValue.type === "momentary" && (
        <MomentaryFields
          form={form}
          controlIndex={controlIndex}
          assignmentIndex={assignmentIndex}
        />
      )}
    </div>
  );
};
