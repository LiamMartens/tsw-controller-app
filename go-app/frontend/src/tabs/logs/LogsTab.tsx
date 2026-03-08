import { useEffect, useRef } from "react";
import { EventsOn } from "../../../wailsjs/runtime/runtime";
import { events } from "../../events";
import { LogLevel, logs } from "../../logs";
import { SaveLogs } from "../../../wailsjs/go/main/App";
import { alert } from "../../utils/alert";
import { useForm } from "react-hook-form";
import clsx from "clsx";

type FormValues = {
  /* debug | info | error */
  loglevel: 0 | 1 | 2;
};

export const LogsTab = () => {
  const logsRef = useRef<HTMLDivElement | null>(null);
  const form = useForm<FormValues>({
    defaultValues: {
      loglevel: 1,
    },
  });
  const loglevel = form.watch("loglevel");

  const handleSave = () => {
    SaveLogs(logs.map(([, msg]) => msg)).catch((err) =>
      alert(String(err), "error"),
    );
  };

  useEffect(() => {
    /* add initial logs once */
    console.log(logs);
    if (logsRef.current) {
      logsRef.current.replaceChildren();
    }
    if (logsRef.current && logs.length) {
      const LOGS_LIMIT = 1000;
      const logsSlice = logs.slice(-LOGS_LIMIT);
      logsRef.current.appendChild(
        document.createTextNode(
          "\n...only showing the last 1000 logs, for all logs please save them as a file...\n\n",
        ),
      );
      for (const [loglevel, msg] of logsSlice) {
        const span = document.createElement("span");
        span.dataset.loglevel = loglevel;
        span.appendChild(document.createTextNode(msg + "\n"));
        logsRef.current.appendChild(span);
      }
    }
  }, []);

  useEffect(() => {
    const handleLogMessageReceived = (level: LogLevel, msg: string) => {
      /* add new logs as they come in */
      requestAnimationFrame(() => {
        if (logsRef.current) {
          const isNearBottom =
            document.documentElement.scrollTop + window.innerHeight >=
            document.documentElement.scrollHeight - window.innerHeight * 0.1;
          const span = document.createElement("span");
          span.dataset.loglevel = level;
          span.appendChild(document.createTextNode(msg + "\n"));
          logsRef.current.appendChild(span);
          if (isNearBottom) {
            /* scroll bottom if near bottom */
            document.documentElement.scrollTop =
              document.documentElement.scrollHeight;
          }
        }
      });
    };

    const unsubscribe_debug = EventsOn(events.log.debug, (msg: string) =>
      handleLogMessageReceived("debug", msg),
    );
    const unsubscribe_info = EventsOn(events.log.info, (msg: string) =>
      handleLogMessageReceived("info", msg),
    );
    const unsubscribe_error = EventsOn(events.log.error, (msg: string) =>
      handleLogMessageReceived("error", msg),
    );

    return () => {
      unsubscribe_debug();
      unsubscribe_info();
      unsubscribe_error();
    };
  }, []);

  return (
    <div>
      <div
        ref={logsRef}
        key="logs"
        className={clsx(
          "whitespace-pre-wrap text-xs font-mono w-full overflow-hidden peer *:data-[loglevel=error]:text-error",
          {
            "*:data-[loglevel=debug]:hidden": loglevel > 0,
            "*:data-[loglevel=info]:hidden": loglevel > 1,
            "*:data-[loglevel=error]:hidden": loglevel > 2,
          },
        )}
      />
      <div className="sticky bottom-0 left-0 right-0 py-3 bg-(--root-bg,var(--color-base-100)) border-t border-t-base-100">
        <div className="flex items-center gap-2">
          <select
            className="select select-xs w-20"
            {...form.register("loglevel", { valueAsNumber: true })}
          >
            <option value="0">Debug</option>
            <option value="1">Info</option>
            <option value="2">Error</option>
          </select>
          <button className="btn btn-primary btn-xs" onClick={handleSave}>
            Save logs
          </button>
        </div>
      </div>
    </div>
  );
};
