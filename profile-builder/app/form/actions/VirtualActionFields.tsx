import { useFormContext, Controller } from "react-hook-form";
import type { profile_builder_schema } from "../types";

interface ActionFieldsProps {
  controlName: string;
  name: string;
}

const getFieldPath = (controlName: string, name: string, field: string) =>
  `${controlName}.${name}.${field}` as any;

export const VirtualActionFields = ({ controlName, name }: ActionFieldsProps) => {
  const { control } = useFormContext<profile_builder_schema>();

  return (
    <div className="space-y-3">
      <Controller
        name={getFieldPath(controlName, name, "type")}
        control={control}
        render={({ field }) => (
          <input type="hidden" value="virtual" {...field} />
        )}
      />
      <Controller
        name={getFieldPath(controlName, name, "control")}
        control={control}
        render={({ field, fieldState }) => (
          <div className="form-control w-full">
            <label className="label">
              <span className="label-text">Control</span>
            </label>
            <input
              {...field}
              type="text"
              placeholder="e.g. virtual:Button1"
              className="input input-bordered w-full"
            />
            {fieldState.error && (
              <span className="text-error text-sm mt-1">{String(fieldState.error.message)}</span>
            )}
          </div>
        )}
      />
      <Controller
        name={getFieldPath(controlName, name, "value")}
        control={control}
        render={({ field, fieldState }) => (
          <div className="form-control w-1/2">
            <label className="label">
              <span className="label-text">Value</span>
            </label>
            <input
              {...field}
              onChange={(e) => field.onChange(parseFloat(e.target.value))}
              type="number"
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
