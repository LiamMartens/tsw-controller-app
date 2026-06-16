import { EventsOn } from "../wailsjs/runtime/runtime";
import { events } from "./events";

const MAX_LOGS = 15_000;
const LOGS_ROTATE_LENGTH = 10_000;

export type LogLevel = "debug" | "info" | "error";

export const logs = [] as unknown as [LogLevel, string][] & {
  track: (level: LogLevel, msg: string) => void;
};

logs.track = (level, msg) => {
  logs.push([level, msg]);
  if (logs.length > MAX_LOGS) logs.splice(0, logs.length - LOGS_ROTATE_LENGTH);
};

EventsOn(events.log.debug, (msg: string) => logs.track("debug", msg));
EventsOn(events.log.info, (msg: string) => logs.track("info", msg));
EventsOn(events.log.error, (msg: string) => logs.track("error", msg));
