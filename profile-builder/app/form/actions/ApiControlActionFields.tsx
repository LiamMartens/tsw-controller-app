import { useForm } from "react-hook-form";
import { apiControlActionSchema } from "../schema";
import z from "zod";
import { BaseField, FieldGroup } from "../inputs";
import { t } from "../../utils";
import { useEffect, useRef } from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import clsx from "clsx";

type Value = z.infer<typeof apiControlActionSchema>;

type Props = {
  value: Value;
  onChange: (value: Value) => void;
};

export const ApiControlActionFields = ({ value, onChange }: Props) => {
  const onChangeRef = useRef(onChange);
  onChangeRef.current = onChange;

  const form = useForm<Value>({
    resolver: zodResolver(apiControlActionSchema),
    mode: "onBlur",
    defaultValues: value,
  });

  useEffect(() => {
    form.watch(() => {
      form.handleSubmit(onChange)();
    });
  }, [onChangeRef, form]);

  return (
    <div className="space-y-3">
      <BaseField
        legend={t("Controls")}
        label={t("The API control name to change")}
        error={form.formState.errors.controls?.message}
      >
        <input
          className={clsx(
            "input input-bordered w-full",
            form.formState.errors.controls && "input-error",
          )}
          {...form.register("controls")}
        />
      </BaseField>

      <BaseField
        legend={t("Value")}
        label={t("The value to send to the API")}
        error={form.formState.errors.api_value?.message}
      >
        <input
          type="number"
          className={clsx(
            "input input-bordered w-full",
            form.formState.errors.api_value && "input-error",
          )}
          {...form.register("api_value", { valueAsNumber: true })}
        />
      </BaseField>

      <BaseField
        legend={t("Maximum Change Rate")}
        label={t(
          "Defines the maximum change rate of the control. This is only required for some train controls who don't react well to instant changes.",
        )}
        error={form.formState.errors.max_change_rate?.message}
      >
        <input
          type="number"
          className={clsx(
            "input input-bordered w-full",
            form.formState.errors.max_change_rate && "input-error",
          )}
          {...form.register("max_change_rate", {
            setValueAs: (v) => {
              if (String(v).trim() === "") return undefined;
              return Number(v);
            },
          })}
        />
      </BaseField>

      <FieldGroup legend={t("Options")}>
        <div className="grid grid-cols-2 gap-4">
          <label className="label whitespace-normal">
            <input
              type="checkbox"
              className="checkbox"
              {...form.register("hold")}
            />
            <p>{t("Enable hold (Continuously send the value to the API)")}</p>
          </label>
        </div>
      </FieldGroup>
    </div>
  );
};
