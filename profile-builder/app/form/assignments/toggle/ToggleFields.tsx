import { ActionSelector } from "../../actions/ActionSelector";
import { CommonFields } from "../CommonFields";

interface ToggleFieldsProps {
  controlName: string;
}

export const ToggleFields = ({ controlName }: ToggleFieldsProps) => {
  return (
    <div className="space-y-4">
      <ActionSelector controlName={controlName} name="action_activate" />
      <ActionSelector controlName={controlName} name="action_deactivate" />
      <CommonFields controlName={controlName} />
    </div>
  );
};
