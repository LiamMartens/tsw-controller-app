import { useFormContext, Controller, useForm } from "react-hook-form";
import z from "zod";
import { virtualActionSchema } from "../schema";
import { useRef } from "react";
import { BaseField } from "../inputs";
import { t } from "../../utils";
import { zodResolver } from "@hookform/resolvers/zod";

type Value = z.infer<typeof virtualActionSchema>;

type Props = {
  value: Value;
  onChange: (value: Value) => void;
};

export const VirtualActionFields = ({ value, onChange }: Props) => {
  const onChangeRef = useRef(onChange);
  onChangeRef.current = onChange;

  const form = useForm<Value>({
    resolver: zodResolver(virtualActionSchema),
    mode: "onBlur",
    defaultValues: value,
  });

  return (
    <div className="space-y-3">
      <BaseField
        legend={t("Virtual Control Name")}
        label={t("The name of the virtual control to set")}
        error={form.formState.errors.control?.message}
      >
        <input
          className="input input-bordered w-full"
          {...form.register("control")}
        />
      </BaseField>

      <BaseField
        legend={t("Value")}
        label={t("The value to set the virtual control to")}
        error={form.formState.errors.value?.message}
      >
        <input
          type="number"
          className="input input-bordered w-full"
          {...form.register("value", { valueAsNumber: true })}
        />
      </BaseField>
    </div>
  );
};
