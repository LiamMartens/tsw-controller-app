import { UseFormReturn } from "react-hook-form";
import { t } from "../../utils";
import { AssignmentTypeSelector } from "./AssignmentTypeSelector";
import z from "zod";
import { profileSchema } from "../schema";
import { useConfirmModal } from "../modals/useConfirmModal";

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
  const { confirm, render: ConfirmDeleteComponent } = useConfirmModal({
    title: t("Are you sure?"),
    body: t("Are you sure you want to remove this assignment?"),
    onConfirm: () => onRemoveAssignment(),
  });

  return (
    <div className="collapse collapse-arrow bg-base-200 rounded-lg">
      <input type="checkbox" className="peer" />
      <div className="collapse-title flex justify-between items-center">
        <span className="font-medium">
          {t("Assignment")} {assignmentIndex + 1}
        </span>
      </div>

      <div className="collapse-content space-y-4">
        <AssignmentTypeSelector
          form={form}
          controlIndex={controlIndex}
          assignmentIndex={assignmentIndex}
        />

        <div className="flex justify-start border-t border-base-300 pt-4">
          <button
            type="button"
            className="btn btn-error btn-xs btn-ghost"
            onClick={() => confirm()}
          >
            {t("Remove Assignment")}
          </button>
        </div>
      </div>

      <ConfirmDeleteComponent />
    </div>
  );
};
