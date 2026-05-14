import z from "zod";
import { keysActionchema } from "../schema";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { BaseField } from "../inputs";
import { t } from "../../utils";
import clsx from "clsx";
import { useEffect, useRef } from "react";

type Value = z.infer<typeof keysActionchema>;

type Props = {
  value: Value;
  onChange: (value: Value) => void;
};

export const KeysActionField = ({ value, onChange }: Props) => {
  const onChangeRef = useRef(onChange);
  onChangeRef.current = onChange;

  const form = useForm<Value>({
    resolver: zodResolver(keysActionchema),
    mode: "onBlur",
    defaultValues: {
      keys: "",
    },
  });

  useEffect(() => {
    form.watch(() => {
      form.handleSubmit(onChange)();
    });
  }, [onChangeRef, form]);

  return (
    <div className="space-y-3">
      <BaseField
        legend={t("Keys")}
        label={t("The keys to activate. Can be multiple, eg: ctrl+d")}
        error={form.formState.errors.keys?.message}
      >
        <input
          className="input input-bordered w-full"
          {...form.register("keys")}
        />
      </BaseField>

      <BaseField
        legend={t("Wait Time")}
        label={t(
          "(Optional) Specifies the number of seconds to wait between key actions",
        )}
        error={form.formState.errors.wait_time?.message}
      >
        <input
          type="number"
          className={clsx(
            "input input-bordered w-full",
            form.formState.errors.wait_time && "input-error",
          )}
          {...form.register("wait_time", {
            setValueAs: (value) => {
              if (String(value).trim() === "") return undefined;
              return Number(value);
            },
          })}
        />
      </BaseField>

      <BaseField
        legend={t("Press Time")}
        label={t(
          "(Optional) Specifies the number of seconds to hold the keys for. When omitted, keys will be held until manually released.",
        )}
      >
        <input
          type="number"
          className={clsx(
            "input input-bordered w-full",
            form.formState.errors.press_time && "input-error",
          )}
          {...form.register("press_time", {
            setValueAs: (value) => {
              if (String(value).trim() === "") return undefined;
              return Number(value);
            },
          })}
        />
      </BaseField>
    </div>
  );
};
