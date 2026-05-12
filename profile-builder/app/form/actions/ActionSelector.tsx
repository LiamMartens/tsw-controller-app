import { useFormContext, Controller } from "react-hook-form";
import type { profile_builder_schema } from "../types";
import { ACTION_TYPES, getActionType } from "../utils";
import { KeysActionFields } from "./KeysActionFields";
import { DirectControlActionFields } from "./DirectControlActionFields";
import { ApiControlActionFields } from "./ApiControlActionFields";
import { VirtualActionFields } from "./VirtualActionFields";

interface ActionSelectorProps {
  controlName: string;
  name: string;
}

const getFieldPath = (controlName: string, name: string, field: string) =>
  `${controlName}.${name}.${field}` as any;

const actionFieldComponents: Record<
  string,
  React.ComponentType<{ controlName: string; name: string }>
> = {
  keys: KeysActionFields,
  direct_control: DirectControlActionFields,
  api_control: ApiControlActionFields,
  virtual: VirtualActionFields,
};

export const ActionSelector = ({ controlName, name }: ActionSelectorProps) => {
  const { control, watch, setValue } = useFormContext<profile_builder_schema>();
  const actionPath = `${controlName}.${name}` as any;
  const currentValue = watch(actionPath);
  const currentType = getActionType(currentValue);
  const actionType = currentType || "keys";

  const ActionComponent =
    actionFieldComponents[actionType] || actionFieldComponents.keys;

  return (
    <div className="space-y-3">
      <Controller
        name={getFieldPath(controlName, name, "keys")}
        control={control}
        render={({ field }) => <input type="hidden" {...field} />}
      />
      <Controller
        name={getFieldPath(controlName, name, "controls")}
        control={control}
        render={({ field }) => <input type="hidden" {...field} />}
      />
      <Controller
        name={getFieldPath(controlName, name, "api_value")}
        control={control}
        render={({ field }) => <input type="hidden" {...field} />}
      />
      <Controller
        name={getFieldPath(controlName, name, "type")}
        control={control}
        render={({ field }) => <input type="hidden" {...field} />}
      />
      <div className="form-control w-1/3">
        <label className="label">
          <span className="label-text">Action Type</span>
        </label>
        <select
          className="select select-bordered w-full"
          value={actionType}
          onChange={(e) => {
            const newType = e.target.value;
            setValue(
              actionPath,
              newType === "keys"
                ? { keys: "" }
                : newType === "direct_control"
                  ? { controls: "", value: 0 }
                  : newType === "api_control"
                    ? { controls: "", api_value: 0 }
                    : { type: "virtual", control: "", value: 0 },
            );
          }}
        >
          {ACTION_TYPES.map((t) => (
            <option key={t.value} value={t.value}>
              {t.label}
            </option>
          ))}
        </select>
      </div>
      <ActionComponent controlName={controlName} name={name} />
    </div>
  );
};
