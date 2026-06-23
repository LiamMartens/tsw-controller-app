import clsx from "clsx";
import { AnimatePresence, motion } from "framer-motion";
import React, { useCallback, useEffect, useRef, useState } from "react";
import { createRoot } from "react-dom/client";

type AlertVariant = "info" | "success" | "error";
type AlertDetails = {
  key: string;
  message: string;
  variant: AlertVariant;
};

const alertVariantToClassName: Record<AlertVariant, string> = {
  info: "alert-info",
  error: "alert-error",
  success: "alert-success",
};

const Alerts = () => {
  const [alerts, setAlerts] = useState<AlertDetails[]>([]);
  const dismissTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const current = alerts?.[0];

  const dismiss = useCallback(() => {
    if (dismissTimeoutRef.current) {
      clearTimeout(dismissTimeoutRef.current);
      dismissTimeoutRef.current = null;
    }
    setAlerts((cur) => cur.slice(1));
  }, [dismissTimeoutRef, setAlerts]);

  useEffect(() => {
    const handleAlert = (event: Event) => {
      if (event instanceof CustomEvent) {
        const detail = event.detail as AlertDetails;
        dismissTimeoutRef.current = setTimeout(dismiss, 4000);
        setAlerts((current) => [...current, detail]);
      }
    };
    document.body.addEventListener("x-alert", handleAlert);
    return () => document.body.removeEventListener("x-alert", handleAlert);
  }, [dismiss, dismissTimeoutRef]);

  return (
    <div className="fixed bottom-4 w-full grid justify-center pointer-events-none z-50">
      <AnimatePresence mode="sync">
        {current && (
          <motion.div
            key={current.key}
            initial={{ opacity: 0, y: "100%" }}
            animate={{ opacity: 1, y: "0%" }}
            exit={{ opacity: 0, y: "-100%" }}
            onClick={dismiss}
            className={clsx(
              "alert cursor-pointer pointer-events-auto col-start-1 row-start-1",
              alertVariantToClassName[current.variant],
            )}
          >
            {current.message}
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
};
const alertsContainer = document.createElement("div");
document.body.appendChild(alertsContainer);
const alertsRoot = createRoot(alertsContainer);
alertsRoot.render(
  <React.StrictMode>
    <Alerts />
  </React.StrictMode>,
);

export const alert = (message: string, variant: AlertVariant) => {
  document.body.dispatchEvent(
    new CustomEvent("x-alert", {
      detail: { key: `${variant}_${message}_${Date.now()}`, message, variant },
    }),
  );
};
