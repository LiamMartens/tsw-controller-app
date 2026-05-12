import { ActionSelector } from "../../actions/ActionSelector";
import { CommonFields } from "../CommonFields";

interface MomentaryFieldsProps {
  controlName: string;
}

export const MomentaryFields = ({ controlName }: MomentaryFieldsProps) => {
  return (
    <div className="space-y-4">
      <ActionSelector controlName={controlName} name="action_activate" />
      <ActionSelector controlName={controlName} name="action_deactivate" />
      <CommonFields controlName={controlName} />
    </div>
  );
};
