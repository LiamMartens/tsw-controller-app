import { useFormContext, Controller, useFieldArray } from "react-hook-form";
import type { profile_builder_schema } from "./types";
import { CONDITION_OPERATORS } from "../utils";

interface CommonFieldsProps {
  controlName: string;
}

export const CommonFields = ({ controlName }: CommonFieldsProps) => {
  const { control } = useFormContext<profile_builder_schema>();
  const { fields: conditionFields, append: appendCondition, remove: removeCondition } = useFieldArray({
    control,
    name: `${controlName}.conditions` as any,
  });
  const { fields: railFields, append: appendRail, remove: removeRail } = useFieldArray({
    control,
    name: `${controlName}.rail_class_information` as any,
  });

  return (
    <div className="space-y-4">
      <div className="divider text-sm">Conditions</div>
      {conditionFields.map((field, index) => (
        <div key={field.id} className="flex gap-2 items-end">
          <div className="form-control w-1/4">
            <label className="label">
              <span className="label-text">Control</span>
            </label>
            <input
              type="text"
              placeholder="Control name"
              className="input input-bordered w-full"
            />
          </div>
          <div className="form-control w-1/4">
            <label className="label">
              <span className="label-text">Operator</span>
            </label>
            <select className="select select-bordered w-full">
              {CONDITION_OPERATORS.map((op) => (
                <option key={op.value} value={op.value}>{op.label}</option>
              ))}
            </select>
          </div>
          <div className="form-control w-1/4">
            <label className="label">
              <span className="label-text">Value</span>
            </label>
            <input
              type="number"
              step="0.01"
              placeholder="Value"
              className="input input-bordered w-full"
            />
          </div>
          <button
            type="button"
            className="btn btn-error btn-sm mb-0.5"
            onClick={() => removeCondition(index)}
          >
            Remove
          </button>
        </div>
      ))}
      <button
        type="button"
        className="btn btn-outline btn-sm"
        onClick={() => appendCondition({ control: "", operator: "gte", value: 0 })}
      >
        Add Condition
      </button>
      <div className="divider text-sm">Rail Class Information</div>
      {railFields.map((field, index) => (
        <div key={field.id} className="flex gap-2 items-end">
          <div className="form-control w-2/3">
            <label className="label">
              <span className="label-text">Class Name</span>
            </label>
            <input
              type="text"
              placeholder="Rail class name"
              className="input input-bordered w-full"
            />
          </div>
          <button
            type="button"
            className="btn btn-error btn-sm mb-0.5"
            onClick={() => removeRail(index)}
          >
            Remove
          </button>
        </div>
      ))}
      <button
        type="button"
        className="btn btn-outline btn-sm"
        onClick={() => appendRail({ class_name: "" })}
      >
        Add Rail Class
      </button>
    </div>
  );
};
