# TSW Controller App

This program allows you to use any joystick to directly control various train simulator games with any controller (Train Sim World 5/6, Train Simulator Classic, Wonders of Sodor, Running Train). This is NOT a raildriver compatibility layer, rather it interfaces directly with the game.

## Feature Highlights
### Controller Specific Profile Selection
You can select a profile for each controller allowing for a complex multi controller set-up with different active profiles.  
  
![Controller Specific Profiles](https://i.postimg.cc/pXT0Gwr7/controller-specific-profiles.png)  
  
### Cab Debugger
The cab debugger gives a real time status of the current in-game locomotive as the in game controls are changing. This is useful for configuring new train profiles and checking the relevant values.  
  
![Cab Debugger](https://i.postimg.cc/N0YjY73f/Highlights_-_Cab_Debugger.png)  

### Visual Calibration
The new 1.0.0 version also brings a completely overhauled visual calibration mode making it easier than ever to calibrate and configure existing or new controllers  
  
![Visual Calibration](https://i.postimg.cc/pV9hskxB/Highlights_-_UI_Calibration.png)  
  
### Profile Builder
A graphical profile builder is now available online to help with configuring new profiles if you are not comfortable creating the JSON profiles  
  
![Open Profile Builder](https://i.postimg.cc/SNFsDhgY/Highlights_-_Open_Profile_Builder.png)  
![Profile Builder](https://i.postimg.cc/VNbp737x/Highlights_-_Profile_Builder.png)  

## Shared Profiles
You can submit profiles for other users to use and download right from the app  
  
![Shared Profiles](https://i.postimg.cc/SK77NdjG/Highlights-Shared-Profiles.png)  

You can find some demos below:

- [Acela Demo](https://f001.backblazeb2.com/file/tsw-controller-app-demos/acela-demo.mp4)
- [Birmingham Cross-City Demo](https://f001.backblazeb2.com/file/tsw-controller-app-demos/birmingham-cross-city.mp4)
- [Class 101 Demo](https://f001.backblazeb2.com/file/tsw-controller-app-demos/class101.mp4)
- [LIRR Demo](https://f001.backblazeb2.com/file/tsw-controller-app-demos/lirr.mp4)
- [Running Train Demo (beta)](https://f001.backblazeb2.com/file/tsw-controller-app-demos/runningtrain/running-train-demo.mp4)

## Installation

### Automatic installation
To install the mod and program just head to the [releases page](https://github.com/LiamMartens/tsw-controller-app/releases) and download the latest installer for your platform. Once you launch the app you will just need to use the "Install mod" action to install the latest mod into Train Sim World or Train Simulator Classic game.  

### Manual installation (TSW)
You can also manually install if you alread have your own UE4SS installed and want to use your existing installation. To do so you will need to download the respective binary for you platform as well as the UE4SS mod and manually place the mod files into the UE4SS directory.

### Manual installation (TSC)
You can also manually install the Train Simulator Classic mod. To do so you will need to download the respective binary for you platform as well as the TSC mod and manually place the mod files in the game directory.

**Note linux users**  
SDL and Webkit2 4.1 are required for this app to work and will need to be installed.  
> From version 1.16.x and up SDL3 is bundled with the binary and does not need to be manually installed anymore

**For Arch users**
On Arch based systems it is recommended to use SDL3 in favor of SDL2 (see [this article](https://wiki.archlinux.org/title/SDL)). You will have to get both the SDL3 and SDL2-compat packages.
[SDL3](https://archlinux.org/packages/extra/x86_64/sdl3/) and [SDL2-compat](https://archlinux.org/packages/extra/x86_64/sdl2-compat/) are in the ['extra' package repository](https://wiki.archlinux.org/title/Official_repositories#extra) and can be installed with pacman
```
sudo pacman -S sdl3 sdl2-compat
```

**For Ubuntu users**
On Ubuntu/Debian you can install it using apt  
```
sudo apt install -y  libsdl2-2.0-0 libwebkit2gtk-4.1-0
```

## Contributing

If you feel like contributing I will happily accept contributions! Some useful contributions would be

- Train configuration improvements.
- Controller SDL mappings and calibrations
- New loco configs

## Links
[Documentation](https://tsw-controller-app.vercel.app/docs)  
[Profile Builder](https://tsw-controller-app.vercel.app/profile-builder)  
[Profile Examples](./shared-profiles)  
[Forum Discussion](https://forums.dovetailgames.com/threads/i-created-a-mod-software-to-directly-control-in-game-trains-using-a-joystick-no-raildriver.90609/#post-999423)  
[Reddit Discussion](https://www.reddit.com/r/trainsimworld/comments/1jqt103/i_created_a_modsoftware_to_directly_control_in/)  
[TrainSimCommunity Post](https://www.trainsimcommunity.com/mods/c3-train-sim-world/c75-utilities/i6396-custom-controller-mapper-control-train-sim-world-with-any-joystick-or-analog-controller-no-rail-driver-required)
