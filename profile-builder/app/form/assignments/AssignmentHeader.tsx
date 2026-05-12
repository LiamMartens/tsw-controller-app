import { useFormContext, Controller } from "react-hook-form";
import type { profile_builder_schema } from "./types";

interface AssignmentHeaderProps {
  controlName: string;
}

const getFieldPath = (controlName: string, field: string) =>
  `${controlName}.${field}` as any;

export const AssignmentHeader = ({ controlName }: AssignmentHeaderProps) => {
  const { control } = useFormContext<profile_builder_schema>();

  return (
    <div className="flex gap-4 flex-wrap">
      <Controller
        name={getFieldPath(controlName, "threshold")}
        control={control}
        render={({ field, fieldState }) => (
          <div className="form-control w-1/4">
            <label className="label">
              <span className="label-text">Threshold</span>
            </label>
            <input
              {...field}
              onChange={(e) => field.onChange(parseFloat(e.target.value))}
              type="number"
              step="0.01"
              className="input input-bordered w-full"
            />
            {fieldState.error && (
              <span className="text-error text-sm mt-1">{String(fieldState.error.message)}</span>
            )}
          </div>
        )}
      />
      <Controller
        name={getFieldPath(controlName, "match")}
        control={control}
        render={({ field }) => (
          <div className="form-control w-1/4">
            <label className="label">
              <span className="label-text">Match</span>
            </label>
            <select
              {...field}
              className="select select-bordered w-full"
            >
              <option value="exceeds">Exceeds</option>
              <option value="equals">Equals</option>
            </select>
          </div>
        )}
      />
    </div>
  );
};
