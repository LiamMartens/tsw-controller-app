
# CHANGELOG

## v1.9.0
- Add search for explore tab

## v1.8.3
- Disabled UE4ss hot reload

## v1.8.2
- Fix tags not always displaying

## v1.8.1
- Add supported apps tags to profiles

## v1.8.0
### Major Features
- Added support for Train Simulator Classic

### Fixes & Improvements
- Refactored data fetching mechanism for better performance and maintainability
- Added support for rail class specific assignments in controller configurations
- Implemented additional configuration options: max_change_rate, control_range
- Made general improvements to API control and improved stability
- Made general improvements to direct control and improved stability
- Fixed CommAPIKey detection for TSW5
- Add support for embedded configuration and profiles
- Fixed bug with inverted calibration normalization
- Add way to remap values when using conditional direct control with partial range

## 1.7.1
- Fix breaking calibration bug

## v1.7.0
- Added virtual controllers; allowing android devices to be used as a customizable controller

## v1.6.1
- Clarify exe selection
- Add select CommAPIKey.txt option
- Don't use control mode if not connected

## v1.6.0
- Allow embedding calibration into shared profiles to enable fully pre-configured profiles

## v1.5.3
- Add notice for unconfigured controllers

## v1.5.2
- Change deadzone logic to update the value but set it to 0
- Update deadzone logic to start 0 at the deadzone limit instead
- Apply default deadzones and min/max limit adjustment when configuring new controllers
- Don't override previously configured controls automatically
- Update smoothing logic when processing raw events
- Interpret `notify` to true by default

## v1.5.1
- Add `notify` flag to enable value changes being displayed in-game when using direct control.

## v1.5.0
- Add axis easing to calibration dialog

## v1.4.2
- Fix virtual toggle

## 1.4.1
- Fixes auto-selection
- Fixes virtual conditions

## 1.4.0
- Implement virtual controls
- Cab Debugger can connect
- Only show last 1000 logs by default
- Fix raw events
- Fix unique joystick IDs
- Allow auto-selecting with only rail class information
- Fix SDL events
- Improvements to DLL mod for memory safety and performance

## 1.3.6
- Fix bug when saving calibration files
- Add way to select profile for all controllers

## 1.3.5
- Update webkit to 4.1 for debian/ubuntu
- Fix conditions missing from sharing and profile builder
- Fix "Submit now" link

## 1.3.4
- Fix bug with throttling

## 1.3.3
- Fix explore tab sources
- Add proxy mode

## 1.3.2
- Update profile IDs to not include updated time
- Streamline profile selection logic
- Fixes event throttling bug
- Fixes sync control auto-detection
- Fixes bug in TSW API debugger

## 1.3.1
- Added controller name and USB ID to UI
- Added auto-detection notice

## 1.3.0
- Show all profiles regardless of duplicate names
- Show better warnings for duplicate extends
- Add full support for autodetection based on rail class and controller

## 1.2.1
- Omit "Metadata" key

## 1.2.0
- Added "Always on top" option in settings
- Fixed api control action bug
- Added more information in profile selection dropdown
- Added support for author information in explore tab
- Add "extends" option for profiles to compose profiles

## 1.1.0
- Implements new "api_control" mode to utilize the API instead of the mod. This uses the same control name but sends the value using the HTTP API (may require the API key to be configured).  
**Note** has higher overhead than the direct control mode.
- Updates cab debuggger to utilize HTTP API if available
- Added search box to cab debugger
- Added preferred control mode in settings

Example assignment:
```json
{
  "type": "api_control",
  "controls": "Throttle1",
  "input_value": {
    "min": 0.0,
    "max": 1.0,
    "invert": true
  }
}
```

Example action:
```json
{
  "controls": "Throttle1",
  "api_value": 0.5
}
```

## 1.0.5
- Use custom dwmapi proxy
- Run SDL polling using `SDL.Do`
- Add option to delete profile
- Updated alerts and confirm dialogs

## 1.0.4
- Fixed filepath handling on Windows

## 1.0.3
- Fixed build version on Windows
- Show all logs in logs tab
- Add option to save logs to file
- Add option to reset cab debugger state

## 1.0.2
- Fixed calibration mode if directories are non-existant

## 1.0.1
- Fixed "Browse Configuration" not opening the correct folder on Windows

## 1.0.0
#### MAJOR RELEASE - BREAKING CHANGES
- Complete runtime and UI overhaul
- Added visual Cab Debugger
- Added visual Calibration mode
- Added controller specific profile selection
- Added conditional assignments

#### UPGRADE GUIDE FROM < 1.0.0
If you are already using the previous software you can do one of either:
- Manually download the program and replace the mod and program as normal.
- Remove the current UE4SS + mod installation. Download the new release and install the mod using the graphical mod installer.

In either case - the locations of the profiles and configuration has changed. You will need to move the profiles to the new directory by opening the App and using the "Browse configuration" option from the "More" menu next to your controller.

#### NORMAL INSTALLATION GUIDE CAN BE FOUND ON THE [HOME README PAGE](./README.md)

## 0.2.5
- Fix calibration for calibrating multiple controllers.

## 0.2.4
- Fix calibration mode not exiting and writing files.

## 0.2.3
- Update the mixing of `null` values to act as free range zones instead of automatic interpolation zones. This makes for smoother actions between detents. Eg, the following steps value: `[0.0, null, 0.5, 0.6, null, 1.0]` - will snap to `0.5` and `0.6` but allow free range of motion between `0.0` and `0.5` and `0.6` and `1.0`.

## 0.2.2
- Improve performance overhead by reducing controller polling.

## 0.2.1

- Add support for mixing `null` values in the `steps` array for direct control assignments. This new feature allows automatic interpolation between step values without having to manually calculate the steps. Eg: `[0, null, null, 1]` will result in the following actual step list: `[0, 0.33, 0.66, 1]`. This is useful for levers where you want a combination of semi-free range and stepped values. (ie: North American suppression steps mixed with percentage based free range)

## 0.2.0

- Add "relative" option for direct control actions outside of direct control assignments. This allows relative value setting (ie: set the value to be -0.4 below the current value).

## 0.1.7

- Update usb_id check to be case insensitive
- Updated `Tick` and `None` checks using `Fname`(thanks to [@UE4SS](https://github.com/UE4SS))
