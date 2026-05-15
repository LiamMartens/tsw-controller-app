import { PropsWithChildren } from "react";
import clsx from "clsx";

type Props = PropsWithChildren<{
  className?: string;
  legend?: string;
  label?: string;
  error?: string;
}>;

export const BaseField = ({ legend, label, error, children, className }: Props) => {
  return (
    <fieldset className={clsx("fieldset", className)}>
      {!!legend && <legend className="fieldset-legend">{legend}</legend>}
      <div className="flex flex-col gap-2">
        <div>{children}</div>
        {!!label && <p className="label whitespace-normal">{label}</p>}
        {!!error && <p className="label text-error whitespace-normal">{error}</p>}
      </div>
    </fieldset>
  );
};
