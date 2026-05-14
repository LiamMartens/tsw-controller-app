import { profile_builder_schema } from "../types/profile_builder_schema";
import { t } from "../utils";

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
  { value: "momentary", label: t("Momentary") },
  { value: "toggle", label: t("Toggle") },
  { value: "linear", label: t("Linear") },
  { value: "direct_control", label: t("Direct Control") },
  { value: "api_control", label: t("API Control") },
  { value: "sync_control", label: t("Sync Control") },
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
