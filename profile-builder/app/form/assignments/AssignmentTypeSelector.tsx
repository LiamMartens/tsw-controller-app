import { Controller, UseFormReturn } from "react-hook-form";
import {
  ASSIGNMENT_TYPES,
  getAssignmentType,
  createEmptyAssignment,
} from "../utils";
import { AssignmentHeader } from "./AssignmentHeader";
import { MomentaryFields } from "./momentary/MomentaryFields";
import { ToggleFields } from "./toggle/ToggleFields";
import { LinearFields } from "./linear/LinearFields";
import { DirectControlFields } from "./direct_control/DirectControlFields";
import { ApiControlFields } from "./api_control/ApiControlFields";
import { SyncControlFields } from "./sync_control/SyncControlFields";
import z from "zod";
import { profileSchema } from "../schema";

type Props = {
  form: UseFormReturn<z.infer<typeof profileSchema>>;
  controlIndex: number;
  assignmentIndex: number;
};

const assignmentFieldComponents: Record<
  string,
  React.ComponentType<{ controlName: string }>
> = {
  momentary: MomentaryFields,
  toggle: ToggleFields,
  linear: LinearFields,
  direct_control: DirectControlFields,
  api_control: ApiControlFields,
  sync_control: SyncControlFields,
};

export const AssignmentTypeSelector = ({
  form,
  controlIndex,
  assignmentIndex,
}: Props) => {
  const { watch, control, setValue } = form;
  const currentValue = watch(
    `controls.${controlIndex}.assignments.${assignmentIndex}`,
  );
  const currentType = getAssignmentType(currentValue);
  const assignmentType = currentType || "momentary";

  const AssignmentComponent =
    assignmentFieldComponents[assignmentType] ||
    assignmentFieldComponents.momentary;

  return (
    <div className="space-y-3">
      {/*<div className="form-control w-1/3">
        <label className="label">
          <span className="label-text">Assignment Type</span>
        </label>
        <select
          className="select select-bordered w-full"
          value={assignmentType}
          onChange={(e) => {
            const newType = e.target.value;
            setValue(assignmentPath, createEmptyAssignment(newType as any));
          }}
        >
          {ASSIGNMENT_TYPES.map((t) => (
            <option key={t.value} value={t.value}>
              {t.label}
            </option>
          ))}
        </select>
      </div>
      {(assignmentType === "momentary" ||
        assignmentType === "toggle" ||
        assignmentType === "linear") && (
        <AssignmentHeader controlName={controlName} />
      )}
      <AssignmentComponent controlName={controlName} />*/}
    </div>
  );
};
