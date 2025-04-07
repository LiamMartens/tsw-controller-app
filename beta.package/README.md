# Setting up the beta program

There are 2 parts to this software; the actual software that listens to your controller and the UE4SS mod that receives the commands and sends them to the game. This guide will help you set everything up to work with your controller and locos.

## Summary: Different control types
There are 3 base ways of controlling the trains in game.
1. **Normal keybinds**; these are just key actions that are triggered by certain controller actions. Eg, when a lever exceeds a threshold or a button is pushed. This is useful for normal component interactions such as headlights, doors, ...
2. **Direct Control**; this is a way of directly controlling the components in the game by sending the lever values to the game itself. This is the most accurate control but is not supported in all trains. Some trains, especially older ones, become unstable when trying to interact with it using this method.
3. **Sync Control**; this is a middle ground between normal keybinds and direct control. All the train interaction is done using keybindings (such as `a` and `d` for increasing and decreasing the power handle), but the game will send the current value back to the program. This means the program can activate and release the keys depending on the target value which is still more accurate then setting up manual keybindings to be triggered at regular intervals, but less accurate than the direct control method. However this control method could be more stable if you are experiencing problems.

## Calibrating/configuring a new controller

To start you will need to configure and calibrate your controller. The configuration is used to map each button and lever to a common name which will be used in the train profiles. The calibration is used to determine the min and max values of your levers/axes. By default you will find a configuration for the TCA Quadrant Boeing edition since that's the controller I have. If you are using the same controller you can use that SDL mapping and just keep the calibration file. If you have a different controller you will need to keep both generated files after running the calibration.

To start the calibration mode you will need to run the `tsw5-gamepad.{linux/windows}` from a command prompt or terminal with the `calibrate` command. This will look something like this: `./tsw5-gamepad calibrate`. This will start up the calibration mode. In this mode you will need to move each lever and press each button on the controller. This will prompt you to enter the common name for the component. In the default profiles the common names are called `Lever1-3`, `Button1-5` and then there are some special states like `Dial1CW` for the upper dial of the TCA Quadrant clockwise rotation. You can look in the `tca_quadrant_boeing.json` SDL mapping file to see all the common names. As you are setting up each lever you should also make sure to move each lever to it's min and max positions so it can be calibrated correctly.

Once you are done, you can press Q and hit `[Enter]` to exit calibration mode and write the new configuration files. You should now see a new file in the `app/config/sdl_mappings` directory. This is the mapping of your controls to the common names. If you already had a configuration for the same controller ID (such as the TCA Quadrant Boeing Edition) you should make sure you only keep 1 configuration. Additionally, you will also see a new file in the `app/config/calibration` directory. This contains the min and max calibration values of your levers. You should once again make sure there is only 1 calibration file for each controller ID.

**Note**: You can customize the calibration file with some additional options like `invert` (to invert the lever values) and `easing_curve` to change the lever behavior either to be more linear, less linear etc.. You can check out the `tca_quadrant_boeing.json` calibration file for some examples. Additionally it can be a good idea to adjust the max and min values in the calibration file as they are the absolute extremes which are sometimes not easily reached in normal gameplay. For example, I have my controller configured at 2000 below and 2000 above the max and min values respectively in order to reach the 1.0 value more consistently.

That's all the required configuration for your controller.

## Installing the mod

### Install UE4SS

To get started you will need to install UE4SS into Train Sim World. The instructions to do so can be found on [their website](https://docs.ue4ss.com/dev/installation-guide.html).

This mod will require the experimental latest version. As of the time of writing this version was released on December 29, 2024 ([https://github.com/UE4SS-RE/RE-UE4SS/releases/tag/experimental-latest](https://github.com/UE4SS-RE/RE-UE4SS/releases/tag/experimental-latest)).

**Note** important note, UE4SS won't work with TSW unless you specify the engine major/minor versions. To do so you will need to open up the `UE4SS-Settings.ini` file and edit the `[EngineVersionOverride]` section as follows:
```
[EngineVersionOverride]
MajorVersion = 4
MinorVersion = 26
```

### Install the mod

Once you have the `ue4ss` directory in the right place you can simply copy the `ue4ss/Mods/TSWControllerMod` directory into the `ue4ss` Mods directory as directed by UE4SS's guides. That's all the required installation on UE4SS.

## Enable steam controller support and clear the bindings

To ensure the game doesn't pick up your controller natively you will need to enable Steam Input and clear out all the default assigned controls in the game specific controller configuration. If you don't do this the game will make a feable attempt at interacting with the controller. To make your throttle work with steam input you will likely also need to do an initial calibration in the Steam Settings > Controller menu. Since this is a non "standard" controller steam will not fully detect it until calibrated. You only need to set the A and B buttons during calibration everything else can be skipped.   
![Steam Game Options](https://i.ibb.co/Dfvj9WSB/Screenshot-from-2025-02-10-07-56-09.png)  
![Steam Game Options](https://i.ibb.co/YBdkZ74V/Screenshot-from-2025-02-09-22-11-04.png)  
![Empty Configuration](https://i.ibb.co/DPF6P3GM/Screenshot-from-2025-02-09-22-11-40.png)  
![Empty Configuration](https://i.ibb.co/0Rh3kLnV/Screenshot-from-2025-02-09-22-11-52.png)  
![Empty Configuration](https://i.ibb.co/rGTFTpn4/Screenshot-from-2025-02-09-22-11-55.png)  
![Empty Configuration](https://i.ibb.co/3yrXxCvf/Screenshot-from-2025-02-09-22-11-57.png)

## Running the program

Now you are ready to go so you can fire up the game and run the `tsw5-gamepad` program as normal. This will open up the UI where you can select the train profile to use.
**Note**: It is a good idea to switch the profile to `None` if you are going to interact with your controller but don't want anything to trigger.

## Advanced: Setting up a new train profile

It is possible to customize or set-up your own train profiles. To start you can copy an existing profile and rename it. Have a look at the existing profiles to determine how to configure your own settings like keypresses, holds etc... The configuration is very versatile.

The most advanced configurations are the `direct_control` and `sync_control` assignment options. These are the options that allow you to have more direct control over the inputs the game. In order to configure these you will need to know the internal names of the train components and their min,max values to set it up as desired. You can follow the rough steps below to do so:

1. Enable the UE4SS debugging consoles by setting the `ConsoleEnabled`, `GuiConsoleEnabled`, `GuiConsoleVisible` to 1 in the `UE4SS-settings.ini` file.

2. Now startup the game and go to the training center and spawn the loco you want to configure.

3. Switch to the UE4SS debugging tools and go to the "Dumpers" menu. Click "Generate LUA Types".

4. Now you will have to open up a code editor inside the `ue4ss` directory and you are going to need to look for the loco class which inherits the `ARailVehicle` class in Lua. You will be looking for something like this: `---@class ARVM_BPE_AMTK_Acela_PowerCarBase_C : ARailVehicle` (this is from the Acela pak).  
![Code Example](https://i.ibb.co/zVzXMNkc/Screenshot-from-2025-02-09-22-23-30.png)  

6. In this class you will find a bunch of properties (called `@field`) - each corresponding to in game controls. Here you should look for the control you need; usually this is of the type `UIrregularLeverComponent`, for the Acela it's called `---@field ThrottleLever UIrregularLeverComponent`. That will be the name of the control you need to add to your `direct_control` assignment.

7. Lastly, you can play around with the lever in game using your normal controls and monitor the output in the `UE4SS` console window. This will help you figure out the min/max values of the control. They are normally always 0-1, however I like to set-up my brake levers to only go from 0-max brake instead of handle off or emergency and add additional controls to reach emergency manually. This makes driving the train easier imo. (eg: On the BR423/425 the default profile is set up to go from 100% power to Max Brake and the trigger and button on the lever are used to manually reach emergency braking when required)

That should be all; you can just close and re-open the program (you don't need to restart the game) to load the new profile and you can test out whether it works as expected. Rinse and repeat until you have everything configured correctly.

## Advanced: Adding controller specific config overrides
If you want to override the config for your specific controller you can create a new profile with the same name, but adding a `"usb_id": ""` key in the config. This key specifies the controller this config is relevant for and will override the general profile.

