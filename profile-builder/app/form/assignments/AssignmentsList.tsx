import z from "zod";
import { useFieldArray, UseFormReturn } from "react-hook-form";
import { profileSchema } from "../schema";
import { t } from "../../utils";
import { AssignmentCard } from "./AssignmentCard";

type Props = {
  controlIndex: number;
  form: UseFormReturn<z.infer<typeof profileSchema>>;
};

export const AssignmentsList = ({ form, controlIndex }: Props) => {
  const { fields, append, remove } = useFieldArray({
    control: form.control,
    name: `controls.${controlIndex}.assignments`,
  });

  const handleAddAssignment = () => {
    append({
      type: "momentary",
      threshold: 0,
      match: "exceeds",
      action_activate: { keys: "" },
    });
  };

  return (
    <div className="space-y-2">
      {fields.map((field, index) => (
        <AssignmentCard
          key={field.id}
          form={form}
          controlIndex={controlIndex}
          assignmentIndex={index}
          onRemoveAssignment={() => remove(index)}
        />
      ))}

      <button
        type="button"
        className="btn btn-sm w-full"
        onClick={handleAddAssignment}
      >
        {t("Add Assignment")}
      </button>
    </div>
  );
};
