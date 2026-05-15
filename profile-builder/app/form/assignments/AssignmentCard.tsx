import { UseFormReturn } from "react-hook-form";
import { t } from "../../utils";
import { AssignmentTypeSelector } from "./AssignmentTypeSelector";
import z from "zod";
import { profileSchema } from "../schema";
import { useConfirmModal } from "../modals/useConfirmModal";
import { useEditConditionsModal } from "../modals/useEditConditionsModal";
import { useEditRailClassInformationModal } from "../modals/useEditRailClassInformationModal";

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
  const assignment = form.watch(
    `controls.${controlIndex}.assignments.${assignmentIndex}`,
  );

  const { confirm: confirmDelete, render: ConfirmDeleteComponent } =
    useConfirmModal({
      title: t("Are you sure?"),
      body: t("Are you sure you want to remove this assignment?"),
      onConfirm: () => onRemoveAssignment(),
    });

  const { open: openEditConditionsModal, render: EditConditionsModal } =
    useEditConditionsModal({
      value: assignment.conditions,
      onChange: (conditions) =>
        form.setValue(
          `controls.${controlIndex}.assignments.${assignmentIndex}.conditions`,
          conditions,
          { shouldTouch: true, shouldDirty: true },
        ),
    });

  const {
    open: openEditRailClassInformationModal,
    render: EditRailClassInformationModal,
  } = useEditRailClassInformationModal({
    value: assignment.conditions,
    onChange: (conditions) =>
      form.setValue(
        `controls.${controlIndex}.assignments.${assignmentIndex}.rail_class_information`,
        conditions,
        { shouldTouch: true, shouldDirty: true },
      ),
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

        <div className="flex justify-start gap-4 border-t border-base-300 pt-4">
          <button
            type="button"
            className="btn btn-error btn-xs btn-ghost"
            onClick={() => confirmDelete()}
          >
            {t("Remove Assignment")}
          </button>
          <button
            type="button"
            className="btn btn-primary btn-xs btn-ghost"
            onClick={() => openEditConditionsModal()}
          >
            {t("Edit Conditions")}
          </button>
          <button
            type="button"
            className="btn btn-primary btn-xs btn-ghost"
            onClick={() => openEditRailClassInformationModal()}
          >
            {t("Edit Rail Class Information")}
          </button>
        </div>
      </div>

      <ConfirmDeleteComponent />
      <EditConditionsModal />
      <EditRailClassInformationModal />
    </div>
  );
};
