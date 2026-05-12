import { useFormContext, Controller, useFieldArray } from "react-hook-form";
import type { profile_builder_schema } from "./types";
import { CommonFields } from "../CommonFields";

interface DirectControlFieldsProps {
  controlName: string;
}

export const DirectControlFields = ({ controlName }: DirectControlFieldsProps) => {
  const { control } = useFormContext<profile_builder_schema>();
  const { fields: stepThresholds, append: appendStepThreshold, remove: removeStepThreshold } = useFieldArray({
    control,
    name: `${controlName}.input_value.step_thresholds` as any,
  });

  return (
    <div className="space-y-4">
      <Controller
        name={`${controlName}.controls`}
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
      <div className="flex gap-4 flex-wrap">
        <Controller
          name={`${controlName}.hold`}
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
          name={`${controlName}.use_normalized`}
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
          name={`${controlName}.enable_api_fallback`}
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
        <Controller
          name={`${controlName}.notify`}
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
      </div>
      <div className="divider text-sm">Control Range</div>
      <Controller
        name={`${controlName}.control_range.start`}
        control={control}
        render={({ field }) => (
          <div className="form-control w-1/4">
            <label className="label">
              <span className="label-text">Range Start</span>
            </label>
            <input
              {...field}
              onChange={(e) => field.onChange(e.target.value === "" ? undefined : parseFloat(e.target.value))}
              type="number"
              step="0.01"
              placeholder="Optional"
              className="input input-bordered w-full"
            />
          </div>
        )}
      />
      <Controller
        name={`${controlName}.control_range.end`}
        control={control}
        render={({ field }) => (
          <div className="form-control w-1/4">
            <label className="label">
              <span className="label-text">Range End</span>
            </label>
            <input
              {...field}
              onChange={(e) => field.onChange(e.target.value === "" ? undefined : parseFloat(e.target.value))}
              type="number"
              step="0.01"
              placeholder="Optional"
              className="input input-bordered w-full"
            />
          </div>
        )}
      />
      <div className="divider text-sm">Input Value</div>
      <Controller
        name={`${controlName}.input_value.min`}
        control={control}
        render={({ field, fieldState }) => (
          <div className="form-control w-1/4">
            <label className="label">
              <span className="label-text">Min</span>
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
        name={`${controlName}.input_value.max`}
        control={control}
        render={({ field, fieldState }) => (
          <div className="form-control w-1/4">
            <label className="label">
              <span className="label-text">Max</span>
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
        name={`${controlName}.input_value.max_change_rate`}
        control={control}
        render={({ field }) => (
          <div className="form-control w-1/4">
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
          </div>
        )}
      />
      <Controller
        name={`${controlName}.input_value.step`}
        control={control}
        render={({ field }) => (
          <div className="form-control w-1/4">
            <label className="label">
              <span className="label-text">Step</span>
            </label>
            <input
              {...field}
              onChange={(e) => field.onChange(e.target.value === "" ? undefined : parseFloat(e.target.value))}
              type="number"
              min={0}
              placeholder="Optional"
              className="input input-bordered w-full"
            />
          </div>
        )}
      />
      <Controller
        name={`${controlName}.input_value.invert`}
        control={control}
        render={({ field }) => (
          <label className="flex items-center gap-2 cursor-pointer">
            <input
              type="checkbox"
              className="checkbox checkbox-sm"
              checked={field.value || false}
              onChange={(e) => field.onChange(e.target.checked)}
            />
            <span className="label-text">Invert</span>
          </label>
        )}
      />
      <div className="divider text-sm">Step Thresholds</div>
      {stepThresholds.map((field, index) => (
        <div key={field.id} className="flex gap-2 items-end">
          <div className="form-control w-1/4">
            <label className="label">
              <span className="label-text">Threshold</span>
            </label>
            <input
              type="number"
              step="0.01"
              placeholder="Threshold"
              className="input input-bordered w-full"
            />
          </div>
          <div className="form-control w-1/4">
            <label className="label">
              <span className="label-text">Threshold End</span>
            </label>
            <input
              type="number"
              step="0.01"
              placeholder="Optional"
              className="input input-bordered w-full"
            />
          </div>
          <div className="form-control w-1/4">
            <label className="label">
              <span className="label-text">Tolerance</span>
            </label>
            <input
              type="number"
              step="0.01"
              placeholder="Optional"
              className="input input-bordered w-full"
            />
          </div>
          <button
            type="button"
            className="btn btn-error btn-sm mb-0.5"
            onClick={() => removeStepThreshold(index)}
          >
            Remove
          </button>
        </div>
      ))}
      <button
        type="button"
        className="btn btn-outline btn-sm"
        onClick={() => appendStepThreshold({ threshold: 0 })}
      >
        Add Step Threshold
      </button>
      <CommonFields controlName={controlName} />
    </div>
  );
};
