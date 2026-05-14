import { PropsWithChildren, ReactNode } from "react";

type Props = PropsWithChildren<{
  legend?: ReactNode;
  label?: ReactNode;
}>;

export const FieldGroup = ({ legend, label, children }: Props) => {
  return (
    <fieldset className="fieldset bg-base-100 border-base-300 flex flex-col gap-4 rounded-box border p-4 w-full">
      {!!legend && <legend className="fieldset-legend">{legend}</legend>}
      <div>{children}</div>
      {!!label && <p className="label whitespace-normal">{label}</p>}
    </fieldset>
  );
};
