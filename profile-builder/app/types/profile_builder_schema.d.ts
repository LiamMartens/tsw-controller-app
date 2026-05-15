export interface profile_builder_schema {
  /** @description The name of this profile */
  name: string;
  /** @description The name of the profile to extend from. The profile inheritance happens on a control basis; if you have the control defined here it will not be resolved from the extended profile. Note: when saving the profile for sharing the fully resolved profile is saved without the extends fields to make sure profiles are independently shareable. */
  extends?: string;
  /** @description Whether this profile supports auto-detection. Requires the supported controller USB ID and rail class information to be set */
  auto_select?: boolean;
  controls: {
    /** @description The given name of this control (as calibrated) */
    name: string;
    assignments: (
      | ({
          /** @enum {unknown} */
          type: "momentary";
          /** @description The threshold which the gamepad control needs to exceed before triggering the action. For most momentary implementations this can be any value since most buttons report a value of 0 or 1 */
          threshold: number;
          /**
           * @description Defines how to interpret the threshold. Defaults to exceeds where the action is executed when the value exceeds the threshold. Can also be set to equals for an exact comparison
           * @default exceeds
           * @enum {unknown}
           */
          match: "exceeds" | "equals";
          /** @description The actual action to activate when the threshold is exceeded */
          action_activate:
            | {
                /**
                 * @description The keys to trigger (a list of key identifiers separated by +'s)
                 * @example q+pagedown
                 */
                keys: string;
                /** @description The number of seconds to hold the button down; can be omitted to just hold it until released */
                press_time?: number;
                /** @description The minimum time in seconds to wait between keystrokes; can be omitted */
                wait_time?: number;
              }
            | {
                /**
                 * @description This is the direct control identifier which can be found using the Cab Debugger
                 * @example Throttle
                 * @example AutomaticBrake_{SIDE}
                 */
                controls: string;
                /** @description The value to send to the cab. Acceptable values depend on the cab and can be determined by using the Cab Debugger */
                value: number;
                /** @description The maximum rate at which this control can change */
                max_change_rate?: number;
                /** @description Defines whether to use the value as a relative adjustment instead of an absolute one. */
                relative?: boolean;
                /** @description Defines whether to hold the value by continuously sending the input value to the cab. This is only required for momentary levers which do not hold positions on their own in the game. (ie some independent brakes) */
                hold?: boolean;
                /** @description Whether to use the normalized value instead of the non-normalized value */
                use_normalized?: boolean;
                /** @description Enables showing the in-game notification when changing values */
                notify?: boolean;
                /** @description Determines whether to enable fallback to the TSW API if available */
                enable_api_fallback?: boolean;
              }
            | {
                /**
                 * @description This is the direct api control identifier which can be found using the Cab Debugger (same as the direct control one). Does not support the {SIDE} placeholder.
                 * @example Throttle
                 * @example AutomaticBrake_F
                 */
                controls: string;
                /** @description The value to send to the cab. Acceptable values depend on the cab and can be determined by using the Cab Debugger */
                api_value: number;
                hold?: boolean;
                max_change_rate?: number;
              }
            | {
                /** @enum {unknown} */
                type: "virtual";
                /**
                 * @description The name of the virtual control to update. Should start with 'virtual:' for clear segmentation
                 * @example virtual:Button1
                 */
                control: string;
                value: number;
              };
          /** @description The action to activate when the threshold is not exceeded anymore. This defaults to just releasing the previously activated key(s). */
          action_deactivate?:
            | {
                /**
                 * @description The keys to trigger (a list of key identifiers separated by +'s)
                 * @example q+pagedown
                 */
                keys: string;
                /** @description The number of seconds to hold the button down; can be omitted to just hold it until released */
                press_time?: number;
                /** @description The minimum time in seconds to wait between keystrokes; can be omitted */
                wait_time?: number;
              }
            | {
                /**
                 * @description This is the direct control identifier which can be found using the Cab Debugger
                 * @example Throttle
                 * @example AutomaticBrake_{SIDE}
                 */
                controls: string;
                /** @description The value to send to the cab. Acceptable values depend on the cab and can be determined by using the Cab Debugger */
                value: number;
                /** @description The maximum rate at which this control can change */
                max_change_rate?: number;
                /** @description Defines whether to use the value as a relative adjustment instead of an absolute one. */
                relative?: boolean;
                /** @description Defines whether to hold the value by continuously sending the input value to the cab. This is only required for momentary levers which do not hold positions on their own in the game. (ie some independent brakes) */
                hold?: boolean;
                /** @description Whether to use the normalized value instead of the non-normalized value */
                use_normalized?: boolean;
                /** @description Enables showing the in-game notification when changing values */
                notify?: boolean;
                /** @description Determines whether to enable fallback to the TSW API if available */
                enable_api_fallback?: boolean;
              }
            | {
                /**
                 * @description This is the direct api control identifier which can be found using the Cab Debugger (same as the direct control one). Does not support the {SIDE} placeholder.
                 * @example Throttle
                 * @example AutomaticBrake_F
                 */
                controls: string;
                /** @description The value to send to the cab. Acceptable values depend on the cab and can be determined by using the Cab Debugger */
                api_value: number;
                hold?: boolean;
                max_change_rate?: number;
              }
            | {
                /** @enum {unknown} */
                type: "virtual";
                /**
                 * @description The name of the virtual control to update. Should start with 'virtual:' for clear segmentation
                 * @example virtual:Button1
                 */
                control: string;
                value: number;
              };
        } & {
          /** @description The conditions to apply to this assignment */
          conditions?: {
            /** @description This is the control which needs to meet the condition */
            control: string;
            /**
             * @description The operation to apply to the control value (greater than, less than, ..)
             * @enum {unknown}
             */
            operator: "gte" | "lte" | "gt" | "lt" | "eq";
            /** @description The comparison value */
            value: number;
          }[];
          /** @description Defines the supported rail class names (can be found at the top of the cab debugger) */
          rail_class_information?: {
            class_name?: string;
          }[];
        })
      | ({
          /** @enum {unknown} */
          type: "toggle";
          /** @description The threshold which the gamepad control needs to exceed before triggering the action. For most toggle implementations this can be any value since most buttons report a value of 0 or 1 */
          threshold: number;
          /**
           * @description Defines how to interpret the threshold. Defaults to exceeds where the action is executed when the value exceeds the threshold. Can also be set to equals for an exact comparison
           * @default exceeds
           * @enum {unknown}
           */
          match: "exceeds" | "equals";
          /** @description The actual action to activate when toggling the first time */
          action_activate:
            | {
                /**
                 * @description The keys to trigger (a list of key identifiers separated by +'s)
                 * @example q+pagedown
                 */
                keys: string;
                /** @description The number of seconds to hold the button down; can be omitted to just hold it until released */
                press_time?: number;
                /** @description The minimum time in seconds to wait between keystrokes; can be omitted */
                wait_time?: number;
              }
            | {
                /**
                 * @description This is the direct control identifier which can be found using the Cab Debugger
                 * @example Throttle
                 * @example AutomaticBrake_{SIDE}
                 */
                controls: string;
                /** @description The value to send to the cab. Acceptable values depend on the cab and can be determined by using the Cab Debugger */
                value: number;
                /** @description The maximum rate at which this control can change */
                max_change_rate?: number;
                /** @description Defines whether to use the value as a relative adjustment instead of an absolute one. */
                relative?: boolean;
                /** @description Defines whether to hold the value by continuously sending the input value to the cab. This is only required for momentary levers which do not hold positions on their own in the game. (ie some independent brakes) */
                hold?: boolean;
                /** @description Whether to use the normalized value instead of the non-normalized value */
                use_normalized?: boolean;
                /** @description Enables showing the in-game notification when changing values */
                notify?: boolean;
                /** @description Determines whether to enable fallback to the TSW API if available */
                enable_api_fallback?: boolean;
              }
            | {
                /**
                 * @description This is the direct api control identifier which can be found using the Cab Debugger (same as the direct control one). Does not support the {SIDE} placeholder.
                 * @example Throttle
                 * @example AutomaticBrake_F
                 */
                controls: string;
                /** @description The value to send to the cab. Acceptable values depend on the cab and can be determined by using the Cab Debugger */
                api_value: number;
                hold?: boolean;
                max_change_rate?: number;
              }
            | {
                /** @enum {unknown} */
                type: "virtual";
                /**
                 * @description The name of the virtual control to update. Should start with 'virtual:' for clear segmentation
                 * @example virtual:Button1
                 */
                control: string;
                value: number;
              };
          /** @description The action to execute when toggling the second time */
          action_deactivate:
            | {
                /**
                 * @description The keys to trigger (a list of key identifiers separated by +'s)
                 * @example q+pagedown
                 */
                keys: string;
                /** @description The number of seconds to hold the button down; can be omitted to just hold it until released */
                press_time?: number;
                /** @description The minimum time in seconds to wait between keystrokes; can be omitted */
                wait_time?: number;
              }
            | {
                /**
                 * @description This is the direct control identifier which can be found using the Cab Debugger
                 * @example Throttle
                 * @example AutomaticBrake_{SIDE}
                 */
                controls: string;
                /** @description The value to send to the cab. Acceptable values depend on the cab and can be determined by using the Cab Debugger */
                value: number;
                /** @description The maximum rate at which this control can change */
                max_change_rate?: number;
                /** @description Defines whether to use the value as a relative adjustment instead of an absolute one. */
                relative?: boolean;
                /** @description Defines whether to hold the value by continuously sending the input value to the cab. This is only required for momentary levers which do not hold positions on their own in the game. (ie some independent brakes) */
                hold?: boolean;
                /** @description Whether to use the normalized value instead of the non-normalized value */
                use_normalized?: boolean;
                /** @description Enables showing the in-game notification when changing values */
                notify?: boolean;
                /** @description Determines whether to enable fallback to the TSW API if available */
                enable_api_fallback?: boolean;
              }
            | {
                /**
                 * @description This is the direct api control identifier which can be found using the Cab Debugger (same as the direct control one). Does not support the {SIDE} placeholder.
                 * @example Throttle
                 * @example AutomaticBrake_F
                 */
                controls: string;
                /** @description The value to send to the cab. Acceptable values depend on the cab and can be determined by using the Cab Debugger */
                api_value: number;
                hold?: boolean;
                max_change_rate?: number;
              }
            | {
                /** @enum {unknown} */
                type: "virtual";
                /**
                 * @description The name of the virtual control to update. Should start with 'virtual:' for clear segmentation
                 * @example virtual:Button1
                 */
                control: string;
                value: number;
              };
        } & {
          /** @description The conditions to apply to this assignment */
          conditions?: {
            /** @description This is the control which needs to meet the condition */
            control: string;
            /**
             * @description The operation to apply to the control value (greater than, less than, ..)
             * @enum {unknown}
             */
            operator: "gte" | "lte" | "gt" | "lt" | "eq";
            /** @description The comparison value */
            value: number;
          }[];
          /** @description Defines the supported rail class names (can be found at the top of the cab debugger) */
          rail_class_information?: {
            class_name?: string;
          }[];
        })
      | ({
          /** @enum {unknown} */
          type: "linear";
          /** @description The linear value which is considered neutral or idle - this can be used to map the lever value from 0-1 to -1 to 1 */
          neutral?: number;
          thresholds: {
            /** @description The threshold to exceed. When a neutral value is set; the value will exceed when below the -x.x value */
            value?: number;
            /** @description When used in combination with value_step, can generate a set of thresholds between value and value_end by value_step to repeat the same action(s) */
            value_end?: number;
            /** @description Only used in combination with value_end */
            value_step?: number;
            /** @description The actual action to activate when the threshold is exceeded */
            action_activate?:
              | {
                  /**
                   * @description The keys to trigger (a list of key identifiers separated by +'s)
                   * @example q+pagedown
                   */
                  keys: string;
                  /** @description The number of seconds to hold the button down; can be omitted to just hold it until released */
                  press_time?: number;
                  /** @description The minimum time in seconds to wait between keystrokes; can be omitted */
                  wait_time?: number;
                }
              | {
                  /**
                   * @description This is the direct control identifier which can be found using the Cab Debugger
                   * @example Throttle
                   * @example AutomaticBrake_{SIDE}
                   */
                  controls: string;
                  /** @description The value to send to the cab. Acceptable values depend on the cab and can be determined by using the Cab Debugger */
                  value: number;
                  /** @description The maximum rate at which this control can change */
                  max_change_rate?: number;
                  /** @description Defines whether to use the value as a relative adjustment instead of an absolute one. */
                  relative?: boolean;
                  /** @description Defines whether to hold the value by continuously sending the input value to the cab. This is only required for momentary levers which do not hold positions on their own in the game. (ie some independent brakes) */
                  hold?: boolean;
                  /** @description Whether to use the normalized value instead of the non-normalized value */
                  use_normalized?: boolean;
                  /** @description Enables showing the in-game notification when changing values */
                  notify?: boolean;
                  /** @description Determines whether to enable fallback to the TSW API if available */
                  enable_api_fallback?: boolean;
                }
              | {
                  /**
                   * @description This is the direct api control identifier which can be found using the Cab Debugger (same as the direct control one). Does not support the {SIDE} placeholder.
                   * @example Throttle
                   * @example AutomaticBrake_F
                   */
                  controls: string;
                  /** @description The value to send to the cab. Acceptable values depend on the cab and can be determined by using the Cab Debugger */
                  api_value: number;
                  hold?: boolean;
                  max_change_rate?: number;
                }
              | {
                  /** @enum {unknown} */
                  type: "virtual";
                  /**
                   * @description The name of the virtual control to update. Should start with 'virtual:' for clear segmentation
                   * @example virtual:Button1
                   */
                  control: string;
                  value: number;
                };
            /** @description The action to execute when the lever goes below the threshold (optional) */
            action_deactivate?:
              | {
                  /**
                   * @description The keys to trigger (a list of key identifiers separated by +'s)
                   * @example q+pagedown
                   */
                  keys: string;
                  /** @description The number of seconds to hold the button down; can be omitted to just hold it until released */
                  press_time?: number;
                  /** @description The minimum time in seconds to wait between keystrokes; can be omitted */
                  wait_time?: number;
                }
              | {
                  /**
                   * @description This is the direct control identifier which can be found using the Cab Debugger
                   * @example Throttle
                   * @example AutomaticBrake_{SIDE}
                   */
                  controls: string;
                  /** @description The value to send to the cab. Acceptable values depend on the cab and can be determined by using the Cab Debugger */
                  value: number;
                  /** @description The maximum rate at which this control can change */
                  max_change_rate?: number;
                  /** @description Defines whether to use the value as a relative adjustment instead of an absolute one. */
                  relative?: boolean;
                  /** @description Defines whether to hold the value by continuously sending the input value to the cab. This is only required for momentary levers which do not hold positions on their own in the game. (ie some independent brakes) */
                  hold?: boolean;
                  /** @description Whether to use the normalized value instead of the non-normalized value */
                  use_normalized?: boolean;
                  /** @description Enables showing the in-game notification when changing values */
                  notify?: boolean;
                  /** @description Determines whether to enable fallback to the TSW API if available */
                  enable_api_fallback?: boolean;
                }
              | {
                  /**
                   * @description This is the direct api control identifier which can be found using the Cab Debugger (same as the direct control one). Does not support the {SIDE} placeholder.
                   * @example Throttle
                   * @example AutomaticBrake_F
                   */
                  controls: string;
                  /** @description The value to send to the cab. Acceptable values depend on the cab and can be determined by using the Cab Debugger */
                  api_value: number;
                  hold?: boolean;
                  max_change_rate?: number;
                }
              | {
                  /** @enum {unknown} */
                  type: "virtual";
                  /**
                   * @description The name of the virtual control to update. Should start with 'virtual:' for clear segmentation
                   * @example virtual:Button1
                   */
                  control: string;
                  value: number;
                };
          }[];
        } & {
          /** @description The conditions to apply to this assignment */
          conditions?: {
            /** @description This is the control which needs to meet the condition */
            control: string;
            /**
             * @description The operation to apply to the control value (greater than, less than, ..)
             * @enum {unknown}
             */
            operator: "gte" | "lte" | "gt" | "lt" | "eq";
            /** @description The comparison value */
            value: number;
          }[];
          /** @description Defines the supported rail class names (can be found at the top of the cab debugger) */
          rail_class_information?: {
            class_name?: string;
          }[];
        })
      | ({
          /** @enum {unknown} */
          type: "direct_control";
          /** @description The direct control name to control; can be identified using the Cab Debugger */
          controls: string;
          /** @description Defines whether to hold the value by continuously sending the input value to the cab. This is only required for momentary levers which do not hold positions on their own in the game. (ie some independent brakes) */
          hold?: boolean;
          /** @description Whether to use the normalized value instead of the non-normalized value */
          use_normalized?: boolean;
          /** @description Determines whether to enable fallback to the TSW API if available */
          enable_api_fallback?: boolean;
          /**
           * @description Enables showing the in-game notification when changing values
           * @default true
           */
          notify: boolean;
          /** @description The control range that will be remapped to 0,1 or 0,-1. This is useful when direct mapping partial ranges. */
          control_range?: {
            start: number;
            end: number;
          };
          /**
           * Direct Like Input Value
           * @description Defines the input value constraints of the direct like control
           */
          input_value: {
            /** @description The minimum reachable value in the game cab; can be identified using the Cab Debugger */
            min: number;
            /** @description The maximum reachable value in the game cab; can be identified using the Cab Debugger */
            max: number;
            /** @description The maximum rate at which this control can change */
            max_change_rate?: number;
            /** @description The step value to increase/decrease the values in (optional) */
            step?: number;
            /**
             * @description Acts similarly to step but allows for finer control and can be combined with null values to define free range zones. This is useful if you have a control which is partly notched and partly free
             * @example [
             *       0.2,
             *       0.4,
             *       0.6,
             *       null,
             *       1
             *     ]
             */
            steps?: (null | number)[];
            /** @description Whether to invert the input value before calculating the game value */
            invert?: boolean;
            step_thresholds?: {
              threshold:
                | number
                | {
                    reference: string;
                    value: number;
                  };
              threshold_end?:
                | number
                | {
                    reference: string;
                    value: number;
                  };
              /** @description Defines the tolerance for the threshold. Defaults to (1/(max(1, num_steps-1))/2) or 0.1; whichever is the lowest. */
              threshold_tolerance?: number;
            }[];
          };
        } & {
          /** @description The conditions to apply to this assignment */
          conditions?: {
            /** @description This is the control which needs to meet the condition */
            control: string;
            /**
             * @description The operation to apply to the control value (greater than, less than, ..)
             * @enum {unknown}
             */
            operator: "gte" | "lte" | "gt" | "lt" | "eq";
            /** @description The comparison value */
            value: number;
          }[];
          /** @description Defines the supported rail class names (can be found at the top of the cab debugger) */
          rail_class_information?: {
            class_name?: string;
          }[];
        })
      | ({
          /** @enum {unknown} */
          type: "api_control";
          /** @description The direct/api control name to control; can be identified using the Cab Debugger */
          controls: string;
          /** @description The control range that will be remapped to 0,1 or 0,-1. This is useful when direct mapping partial ranges. */
          control_range?: {
            start: number;
            end: number;
          };
          hold?: boolean;
          /**
           * Direct Like Input Value
           * @description Defines the input value constraints of the direct like control
           */
          input_value: {
            /** @description The minimum reachable value in the game cab; can be identified using the Cab Debugger */
            min: number;
            /** @description The maximum reachable value in the game cab; can be identified using the Cab Debugger */
            max: number;
            /** @description The maximum rate at which this control can change */
            max_change_rate?: number;
            /** @description The step value to increase/decrease the values in (optional) */
            step?: number;
            /**
             * @description Acts similarly to step but allows for finer control and can be combined with null values to define free range zones. This is useful if you have a control which is partly notched and partly free
             * @example [
             *       0.2,
             *       0.4,
             *       0.6,
             *       null,
             *       1
             *     ]
             */
            steps?: (null | number)[];
            /** @description Whether to invert the input value before calculating the game value */
            invert?: boolean;
            step_thresholds?: {
              threshold:
                | number
                | {
                    reference: string;
                    value: number;
                  };
              threshold_end?:
                | number
                | {
                    reference: string;
                    value: number;
                  };
              /** @description Defines the tolerance for the threshold. Defaults to (1/(max(1, num_steps-1))/2) or 0.1; whichever is the lowest. */
              threshold_tolerance?: number;
            }[];
          };
        } & {
          /** @description The conditions to apply to this assignment */
          conditions?: {
            /** @description This is the control which needs to meet the condition */
            control: string;
            /**
             * @description The operation to apply to the control value (greater than, less than, ..)
             * @enum {unknown}
             */
            operator: "gte" | "lte" | "gt" | "lt" | "eq";
            /** @description The comparison value */
            value: number;
          }[];
          /** @description Defines the supported rail class names (can be found at the top of the cab debugger) */
          rail_class_information?: {
            class_name?: string;
          }[];
        })
      | ({
          /** @enum {unknown} */
          type: "sync_control";
          /** @description The sync control identifier to control; can be identified using the Cab Debugger */
          identifier: string;
          /** @description The control range that will be remapped to 0,1 or 0,-1. This is useful when direct mapping partial ranges. */
          control_range?: {
            start: number;
            end: number;
          };
          /**
           * Direct Like Input Value
           * @description Defines the input value constraints of the direct like control
           */
          input_value: {
            /** @description The minimum reachable value in the game cab; can be identified using the Cab Debugger */
            min: number;
            /** @description The maximum reachable value in the game cab; can be identified using the Cab Debugger */
            max: number;
            /** @description The maximum rate at which this control can change */
            max_change_rate?: number;
            /** @description The step value to increase/decrease the values in (optional) */
            step?: number;
            /**
             * @description Acts similarly to step but allows for finer control and can be combined with null values to define free range zones. This is useful if you have a control which is partly notched and partly free
             * @example [
             *       0.2,
             *       0.4,
             *       0.6,
             *       null,
             *       1
             *     ]
             */
            steps?: (null | number)[];
            /** @description Whether to invert the input value before calculating the game value */
            invert?: boolean;
            step_thresholds?: {
              threshold:
                | number
                | {
                    reference: string;
                    value: number;
                  };
              threshold_end?:
                | number
                | {
                    reference: string;
                    value: number;
                  };
              /** @description Defines the tolerance for the threshold. Defaults to (1/(max(1, num_steps-1))/2) or 0.1; whichever is the lowest. */
              threshold_tolerance?: number;
            }[];
          };
          /** Keys Action */
          action_increase: {
            /**
             * @description The keys to trigger (a list of key identifiers separated by +'s)
             * @example q+pagedown
             */
            keys: string;
            /** @description The number of seconds to hold the button down; can be omitted to just hold it until released */
            press_time?: number;
            /** @description The minimum time in seconds to wait between keystrokes; can be omitted */
            wait_time?: number;
          };
          /** Keys Action */
          action_decrease: {
            /**
             * @description The keys to trigger (a list of key identifiers separated by +'s)
             * @example q+pagedown
             */
            keys: string;
            /** @description The number of seconds to hold the button down; can be omitted to just hold it until released */
            press_time?: number;
            /** @description The minimum time in seconds to wait between keystrokes; can be omitted */
            wait_time?: number;
          };
        } & {
          /** @description The conditions to apply to this assignment */
          conditions?: {
            /** @description This is the control which needs to meet the condition */
            control: string;
            /**
             * @description The operation to apply to the control value (greater than, less than, ..)
             * @enum {unknown}
             */
            operator: "gte" | "lte" | "gt" | "lt" | "eq";
            /** @description The comparison value */
            value: number;
          }[];
          /** @description Defines the supported rail class names (can be found at the top of the cab debugger) */
          rail_class_information?: {
            class_name?: string;
          }[];
        })
    )[];
  }[];
  controller?: {
    /** @description The supported controller USB identifier (Vendor:Product) */
    usb_id?: string;
    /** @description This is a copy of the SDL mapping. This can be embedded for portability */
    mapping?: {
      [key: string]: unknown;
    };
    /** @description This is a copy of the controller calibration. This can be embedded for portability */
    calibration?: {
      [key: string]: unknown;
    };
  };
  /** @description Defines the supported rail class names (can be found at the top of the cab debugger) */
  rail_class_information?: {
    class_name?: string;
  }[];
}
