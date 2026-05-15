import { useMemo } from "react";
import {
  keysActionchema,
  apiControlActionSchema,
  virtualActionSchema,
  directControlActionSchema,
} from "../schema";
import z from "zod";
import { KeysActionField } from "./KeysActionFields";
import { t } from "../../utils";
import { ACTION_TYPES } from "../utils";
import { VirtualActionFields } from "./VirtualActionFields";
import { ApiControlActionFields } from "./ApiControlActionFields";
import { DirectControlActionFields } from "./DirectControlActionFields";

type AnyAction =
  | z.infer<typeof keysActionchema>
  | z.infer<typeof apiControlActionSchema>
  | z.infer<typeof virtualActionSchema>
  | z.infer<typeof directControlActionSchema>;

type Props = {
  value: unknown;
  onChange: (value: AnyAction) => void;
};

export const ActionSelector = ({ value, onChange }: Props) => {
  const action = useMemo(() => {
    const keysAction = keysActionchema.safeParse(value);
    if (keysAction.success)
      return {
        type: "keys" as const,
        value: keysAction.data,
      };

    const apiControlAction = apiControlActionSchema.safeParse(value);
    if (apiControlAction.success)
      return {
        type: "api_control" as const,
        value: apiControlAction.data,
      };

    const virtualAction = virtualActionSchema.safeParse(value);
    if (virtualAction.success)
      return {
        type: "virtual" as const,
        value: virtualAction.data,
      };

    const directControlAction = directControlActionSchema.safeParse(value);
    if (directControlAction.success)
      return {
        type: "direct_control" as const,
        value: directControlAction.data,
      };

    return {
      type: "keys" as const,
      value: keysActionchema.parse({ keys: "" }),
    };
  }, [value]);

  return (
    <div className="space-y-3">
      <div className="form-control">
        <label className="label">
          <span className="label-text">{t("Action Type")}</span>
        </label>
        <select
          className="select select-bordered w-full"
          value={action.type}
          onChange={(e) => {
            const value = e.currentTarget.value as (typeof action)["type"];
            if (value === action.type) return; /*ignore if same */
            const emptyValues: Record<(typeof action)["type"], AnyAction> = {
              keys: { keys: "" },
              virtual: { type: "virtual", control: "", value: 0 },
              api_control: { controls: "", api_value: 0 },
              direct_control: {
                controls: "",
                value: 0,
                notify: true,
                enable_api_fallback: true,
              },
            };
            onChange(emptyValues[value]);
          }}
        >
          {ACTION_TYPES.map((t) => (
            <option key={t.value} value={t.value}>
              {t.label}
            </option>
          ))}
        </select>
      </div>

      {action.type === "keys" && <KeysActionField value={action.value} onChange={onChange} />}

      {action.type === "virtual" && (
        <VirtualActionFields value={action.value} onChange={onChange} />
      )}

      {action.type === "api_control" && (
        <ApiControlActionFields value={action.value} onChange={onChange} />
      )}

      {action.type === "direct_control" && (
        <DirectControlActionFields value={action.value} onChange={onChange} />
      )}
    </div>
  );
};
