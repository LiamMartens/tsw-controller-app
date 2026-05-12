import type { profile_builder_schema } from "./types";

export type AssignmentType =
  | "momentary"
  | "toggle"
  | "linear"
  | "direct_control"
  | "api_control"
  | "sync_control";

export type ActionValueType =
  | "keys"
  | "direct_control"
  | "api_control"
  | "virtual";

export const ASSIGNMENT_TYPES: { value: AssignmentType; label: string }[] = [
  { value: "momentary", label: "Momentary" },
  { value: "toggle", label: "Toggle" },
  { value: "linear", label: "Linear" },
  { value: "direct_control", label: "Direct Control" },
  { value: "api_control", label: "API Control" },
  { value: "sync_control", label: "Sync Control" },
];

export const ACTION_TYPES: { value: ActionValueType; label: string }[] = [
  { value: "keys", label: "Keys Action" },
  { value: "direct_control", label: "Direct Control Action" },
  { value: "api_control", label: "API Control Action" },
  { value: "virtual", label: "Virtual Action" },
];

export const CONDITION_OPERATORS: {
  value: "gte" | "lte" | "gt" | "lt" | "eq";
  label: string;
}[] = [
  { value: "gte", label: "Greater than or equal (>=)" },
  { value: "lte", label: "Less than or equal (<=)" },
  { value: "gt", label: "Greater than (>)" },
  { value: "lt", label: "Less than (<)" },
  { value: "eq", label: "Equals (==)" },
];

export function isMomentary(
  a: unknown,
): a is Extract<
  profile_builder_schema["controls"][number]["assignments"][number],
  { type: "momentary" }
> {
  return (a as any)?.type === "momentary";
}

export function isToggle(
  a: unknown,
): a is Extract<
  profile_builder_schema["controls"][number]["assignments"][number],
  { type: "toggle" }
> {
  return (a as any)?.type === "toggle";
}

export function isLinear(
  a: unknown,
): a is Extract<
  profile_builder_schema["controls"][number]["assignments"][number],
  { type: "linear" }
> {
  return (a as any)?.type === "linear";
}

export function isDirectControl(
  a: unknown,
): a is Extract<
  profile_builder_schema["controls"][number]["assignments"][number],
  { type: "direct_control" }
> {
  return (a as any)?.type === "direct_control";
}

export function isApiControl(
  a: unknown,
): a is Extract<
  profile_builder_schema["controls"][number]["assignments"][number],
  { type: "api_control" }
> {
  return (a as any)?.type === "api_control";
}

export function isSyncControl(
  a: unknown,
): a is Extract<
  profile_builder_schema["controls"][number]["assignments"][number],
  { type: "sync_control" }
> {
  return (a as any)?.type === "sync_control";
}

export function isKeysAction(
  a: unknown,
): a is { keys: string; press_time?: number; wait_time?: number } {
  return (a as any)?.keys !== undefined && (a as any)?.type !== "virtual";
}

export function isDirectControlAction(a: unknown): a is {
  controls: string;
  value: number;
  max_change_rate?: number;
  relative?: boolean;
  hold?: boolean;
  use_normalized?: boolean;
  notify?: boolean;
  enable_api_fallback?: boolean;
} {
  return (
    (a as any)?.controls !== undefined &&
    (a as any)?.value !== undefined &&
    (a as any)?.type !== "virtual"
  );
}

export function isApiControlAction(a: unknown): a is {
  controls: string;
  api_value: number;
  hold?: boolean;
  max_change_rate?: number;
} {
  return (
    (a as any)?.controls !== undefined && (a as any)?.api_value !== undefined
  );
}

export function isVirtualAction(
  a: unknown,
): a is { type: "virtual"; control: string; value: number } {
  return (a as any)?.type === "virtual";
}

export function getAssignmentType(a: unknown): AssignmentType | undefined {
  return (a as any)?.type;
}

export function getActionType(a: unknown): ActionValueType | undefined {
  if ((a as any)?.type === "virtual") return "virtual";
  if ((a as any)?.keys !== undefined) return "keys";
  if ((a as any)?.api_value !== undefined) return "api_control";
  if ((a as any)?.controls !== undefined && (a as any)?.value !== undefined)
    return "direct_control";
  return undefined;
}

export function createEmptyAssignment(
  type: AssignmentType,
): profile_builder_schema["controls"][number]["assignments"][number] {
  switch (type) {
    case "momentary":
      return {
        type: "momentary",
        threshold: 0,
        match: "exceeds",
        action_activate: { keys: "" },
      };
    case "toggle":
      return {
        type: "toggle",
        threshold: 0,
        match: "exceeds",
        action_activate: { keys: "" },
        action_deactivate: { keys: "" },
      };
    case "linear":
      return {
        type: "linear",
        thresholds: [{ value: 0, action_activate: { keys: "" } }],
      };
    case "direct_control":
      return {
        type: "direct_control",
        controls: "",
        notify: true,
        input_value: { min: 0, max: 1 },
      };
    case "api_control":
      return {
        type: "api_control",
        controls: "",
        input_value: { min: 0, max: 1 },
      };
    case "sync_control":
      return {
        type: "sync_control",
        identifier: "",
        input_value: { min: 0, max: 1 },
        action_increase: { keys: "" },
        action_decrease: { keys: "" },
      };
  }
}

export function createEmptyProfile(): profile_builder_schema {
  return { name: "", controls: [] };
}
