import { useFormContext, Controller, useFieldArray } from "react-hook-form";
import type { profile_builder_schema } from "./types";
import { ActionSelector } from "../../actions/ActionSelector";
import { CommonFields } from "../CommonFields";

interface LinearFieldsProps {
  controlName: string;
}

export const LinearFields = ({ controlName }: LinearFieldsProps) => {
  const { control } = useFormContext<profile_builder_schema>();
  const { fields, append, remove } = useFieldArray({
    control,
    name: `${controlName}.thresholds` as any,
  });

  return (
    <div className="space-y-4">
      <Controller
        name={`${controlName}.neutral`}
        control={control}
        render={({ field, fieldState }) => (
          <div className="form-control w-1/4">
            <label className="label">
              <span className="label-text">Neutral</span>
            </label>
            <input
              {...field}
              onChange={(e) =>
                field.onChange(e.target.value === "" ? undefined : parseFloat(e.target.value))
              }
              type="number"
              step="0.01"
              placeholder="Optional (0-1 maps to -1 to 1)"
              className="input input-bordered w-full"
            />
            {fieldState.error && (
              <span className="text-error text-sm mt-1">{String(fieldState.error.message)}</span>
            )}
          </div>
        )}
      />
      <div className="divider text-sm">Thresholds</div>
      {fields.map((field, index) => (
        <div key={field.id} className="collapse collapse-arrow bg-base-200 rounded-lg">
          <input type="checkbox" />
          <div className="collapse-title font-medium">Threshold {index + 1}</div>
          <div className="collapse-content space-y-3">
            <Controller
              name={`${controlName}.thresholds.${index}.value`}
              control={control}
              render={({ field, fieldState }) => (
                <div className="form-control w-1/4">
                  <label className="label">
                    <span className="label-text">Value</span>
                  </label>
                  <input
                    {...field}
                    onChange={(e) => field.onChange(parseFloat(e.target.value))}
                    type="number"
                    step="0.01"
                    className="input input-bordered w-full"
                  />
                  {fieldState.error && (
                    <span className="text-error text-sm mt-1">
                      {String(fieldState.error.message)}
                    </span>
                  )}
                </div>
              )}
            />
            <Controller
              name={`${controlName}.thresholds.${index}.value_end`}
              control={control}
              render={({ field }) => (
                <div className="form-control w-1/4">
                  <label className="label">
                    <span className="label-text">Value End</span>
                  </label>
                  <input
                    {...field}
                    onChange={(e) =>
                      field.onChange(e.target.value === "" ? undefined : parseFloat(e.target.value))
                    }
                    type="number"
                    step="0.01"
                    placeholder="Optional (with value_step)"
                    className="input input-bordered w-full"
                  />
                </div>
              )}
            />
            <Controller
              name={`${controlName}.thresholds.${index}.value_step`}
              control={control}
              render={({ field }) => (
                <div className="form-control w-1/4">
                  <label className="label">
                    <span className="label-text">Value Step</span>
                  </label>
                  <input
                    {...field}
                    onChange={(e) =>
                      field.onChange(e.target.value === "" ? undefined : parseFloat(e.target.value))
                    }
                    type="number"
                    step="0.01"
                    placeholder="Optional (with value_end)"
                    className="input input-bordered w-full"
                  />
                </div>
              )}
            />
            <ActionSelector
              controlName={`${controlName}.thresholds.${index}`}
              name="action_activate"
            />
            <ActionSelector
              controlName={`${controlName}.thresholds.${index}`}
              name="action_deactivate"
            />
            <div className="flex justify-end">
              <button type="button" className="btn btn-error btn-sm" onClick={() => remove(index)}>
                Remove Threshold
              </button>
            </div>
          </div>
        </div>
      ))}
      <button
        type="button"
        className="btn btn-outline btn-sm"
        onClick={() => append({ value: 0, action_activate: { keys: "" } })}
      >
        Add Threshold
      </button>
      <CommonFields controlName={controlName} />
    </div>
  );
};
