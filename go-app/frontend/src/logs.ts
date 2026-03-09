import { EventsOn } from "../wailsjs/runtime/runtime";
import { events } from "./events";

export type LogLevel = 'debug' | 'info' | 'error'

export const logs: [LogLevel, string][] = [];

EventsOn(events.log.debug, (msg: string) => { logs.push(['debug', msg]) });
EventsOn(events.log.info, (msg: string) => { logs.push(['info', msg]) });
EventsOn(events.log.error, (msg: string) => { logs.push(['error', msg]) });
