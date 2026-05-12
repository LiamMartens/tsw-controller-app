import { useFormContext, Controller } from "react-hook-form";
import type { profile_builder_schema } from "../types";

interface ActionFieldsProps {
  controlName: string;
  name: string;
}

const getFieldPath = (controlName: string, name: string, field: string) =>
  `${controlName}.${name}.${field}` as any;

export const KeysActionFields = ({ controlName, name }: ActionFieldsProps) => {
  const { control } = useFormContext<profile_builder_schema>();

  return (
    <div className="space-y-3">
      <Controller
        name={getFieldPath(controlName, name, "keys")}
        control={control}
        render={({ field, fieldState }) => (
          <div className="form-control w-full">
            <label className="label">
              <span className="label-text">Keys</span>
            </label>
            <input
              {...field}
              type="text"
              placeholder="e.g. q+pagedown"
              className="input input-bordered w-full"
            />
            {fieldState.error && (
              <span className="text-error text-sm mt-1">{String(fieldState.error.message)}</span>
            )}
          </div>
        )}
      />
      <Controller
        name={getFieldPath(controlName, name, "press_time")}
        control={control}
        render={({ field, fieldState }) => (
          <div className="form-control w-1/2">
            <label className="label">
              <span className="label-text">Press Time (seconds)</span>
            </label>
            <input
              {...field}
              onChange={(e) => field.onChange(e.target.value === "" ? undefined : parseFloat(e.target.value))}
              type="number"
              min={0}
              placeholder="Optional"
              className="input input-bordered w-full"
            />
            {fieldState.error && (
              <span className="text-error text-sm mt-1">{String(fieldState.error.message)}</span>
            )}
          </div>
        )}
      />
      <Controller
        name={getFieldPath(controlName, name, "wait_time")}
        control={control}
        render={({ field, fieldState }) => (
          <div className="form-control w-1/2">
            <label className="label">
              <span className="label-text">Wait Time (seconds)</span>
            </label>
            <input
              {...field}
              onChange={(e) => field.onChange(e.target.value === "" ? undefined : parseFloat(e.target.value))}
              type="number"
              min={0}
              placeholder="Optional"
              className="input input-bordered w-full"
            />
            {fieldState.error && (
              <span className="text-error text-sm mt-1">{String(fieldState.error.message)}</span>
            )}
          </div>
        )}
      />
    </div>
  );
};
