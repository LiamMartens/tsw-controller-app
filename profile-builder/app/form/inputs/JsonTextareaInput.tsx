import { useState } from "react";

type JsonTextareaProps = {
  value?: Record<string, unknown> | null;
  onChange: (value: Record<string, unknown>) => void;
};

const stringify = (value?: Record<string, unknown> | null) =>
  value ? JSON.stringify(value, null, 2) : "";

export const JsonTextareaInput = ({ value, onChange }: JsonTextareaProps) => {
  const [internalStringValue, setInternalStringValue] = useState(
    stringify(value),
  );

  return (
    <textarea
      value={internalStringValue}
      className="w-full textarea textarea-bordered h-20 font-mono text-sm resize-y"
      onChange={(event) => {
        setInternalStringValue(event.currentTarget.value);
        try {
          onChange(JSON.parse(event.currentTarget.value));
        } catch {}
      }}
    />
  );
};
