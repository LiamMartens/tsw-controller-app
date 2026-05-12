import { UseFormReturn } from "react-hook-form";
import { t } from "../../utils";
import { AssignmentTypeSelector } from "./AssignmentTypeSelector";
import z from "zod";
import { profileSchema } from "../schema";

type Props = {
  form: UseFormReturn<z.infer<typeof profileSchema>>;
  controlIndex: number;
  assignmentIndex: number;
  onRemoveAssignment: () => void;
};

export const AssignmentCard = ({
  form,
  controlIndex,
  assignmentIndex,
  onRemoveAssignment,
}: Props) => {
  return (
    <div className="collapse collapse-arrow bg-base-200 rounded-lg">
      <input type="checkbox" className="peer" />
      <div className="collapse-title flex justify-between items-center">
        <span className="font-medium">
          {t("Assignment")} {assignmentIndex + 1}
        </span>
        <button
          type="button"
          className="btn btn-error btn-xs btn-ghost"
          onClick={(e) => {
            e.stopPropagation();
            onRemoveAssignment();
          }}
        >
          {t("Remove")}
        </button>
      </div>

      <div className="collapse-content">
        <AssignmentTypeSelector
          form={form}
          controlIndex={controlIndex}
          assignmentIndex={assignmentIndex}
        />
      </div>
    </div>
  );
};
