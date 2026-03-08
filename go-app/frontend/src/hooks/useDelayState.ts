import { useMemo, useRef, useState } from "react"

export const useDelayState = <T>(defaultValue: T) => {
  const delayValueTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const [value, setValue] = useState(defaultValue);

  return useMemo(() => ({
    value,
    setValueDelayed: (value: T, delay: number) => {
      if (delayValueTimeoutRef.current) {
        clearInterval(delayValueTimeoutRef.current)
      }
      delayValueTimeoutRef.current = setTimeout(() => setValue(value), delay);
    },
    setValueInstant: (value: T) => {
      if (delayValueTimeoutRef.current) {
        clearInterval(delayValueTimeoutRef.current);
        delayValueTimeoutRef.current = null;
      }
      setValue(value);
    }
  }), [value])
}