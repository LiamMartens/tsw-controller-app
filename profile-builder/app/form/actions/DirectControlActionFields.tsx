import { useFormContext, Controller } from "react-hook-form";
import type { profile_builder_schema } from "../types";

interface ActionFieldsProps {
  controlName: string;
  name: string;
}

const getFieldPath = (controlName: string, name: string, field: string) =>
  `${controlName}.${name}.${field}` as any;

export const DirectControlActionFields = ({ controlName, name }: ActionFieldsProps) => {
  const { control } = useFormContext<profile_builder_schema>();

  return (
    <div className="space-y-3">
      <Controller
        name={getFieldPath(controlName, name, "controls")}
        control={control}
        render={({ field, fieldState }) => (
          <div className="form-control w-full">
            <label className="label">
              <span className="label-text">Controls</span>
            </label>
            <input
              {...field}
              type="text"
              placeholder="e.g. Throttle"
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
      <Controller
        name={getFieldPath(controlName, name, "max_change_rate")}
        control={control}
        render={({ field, fieldState }) => (
          <div className="form-control w-1/2">
            <label className="label">
              <span className="label-text">Max Change Rate</span>
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
      <div className="flex gap-4 flex-wrap">
        <Controller
          name={getFieldPath(controlName, name, "relative")}
          control={control}
          render={({ field }) => (
            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="checkbox"
                className="checkbox checkbox-sm"
                checked={field.value || false}
                onChange={(e) => field.onChange(e.target.checked)}
              />
              <span className="label-text">Relative</span>
            </label>
          )}
        />
        <Controller
          name={getFieldPath(controlName, name, "hold")}
          control={control}
          render={({ field }) => (
            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="checkbox"
                className="checkbox checkbox-sm"
                checked={field.value || false}
                onChange={(e) => field.onChange(e.target.checked)}
              />
              <span className="label-text">Hold</span>
            </label>
          )}
        />
        <Controller
          name={getFieldPath(controlName, name, "use_normalized")}
          control={control}
          render={({ field }) => (
            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="checkbox"
                className="checkbox checkbox-sm"
                checked={field.value || false}
                onChange={(e) => field.onChange(e.target.checked)}
              />
              <span className="label-text">Use Normalized</span>
            </label>
          )}
        />
        <Controller
          name={getFieldPath(controlName, name, "notify")}
          control={control}
          render={({ field }) => (
            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="checkbox"
                className="checkbox checkbox-sm"
                checked={field.value || false}
                onChange={(e) => field.onChange(e.target.checked)}
              />
              <span className="label-text">Notify</span>
            </label>
          )}
        />
        <Controller
          name={getFieldPath(controlName, name, "enable_api_fallback")}
          control={control}
          render={({ field }) => (
            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="checkbox"
                className="checkbox checkbox-sm"
                checked={field.value || false}
                onChange={(e) => field.onChange(e.target.checked)}
              />
              <span className="label-text">API Fallback</span>
            </label>
          )}
        />
      </div>
    </div>
  );
};
