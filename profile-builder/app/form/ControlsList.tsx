import { useFieldArray, Control, UseFormReturn } from "react-hook-form";
import type { profile_builder_schema } from "./types";
import { ControlCard } from "./ControlCard";
import z from "zod";
import { profileSchema } from "./schema";
import { t } from "../utils";

interface ControlsListProps {
  form: UseFormReturn<z.infer<typeof profileSchema>>;
}

export const ControlsList = ({ form }: ControlsListProps) => {
  const { fields, append, remove } = useFieldArray({
    control: form.control,
    name: "controls",
  });

  const handleAddControl = () => {
    append({
      name: "",
      assignments: [],
    });
  };

  return (
    <div className="space-y-2">
      {fields.map((field, index) => (
        <ControlCard key={field.id} form={form} index={index} onRemove={() => remove(index)} />
      ))}

      <button type="button" className="btn btn-sm w-full" onClick={handleAddControl}>
        {t("Add control")}
      </button>
    </div>
  );
};
