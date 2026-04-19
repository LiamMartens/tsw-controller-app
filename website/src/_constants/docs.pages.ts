import fs from "node:fs";

export const pages = {
  "creating-profile-quickstart": {
    title: "Creating A Profile Quickstart - TSW Controller App",
    contents: () => fs.readFileSync("../CREATING_PROFILE_QUICKSTART.md", { encoding: "utf8" }),
  },
  "using-virtual-controls": {
    title: "Using Virtual Controls - TSW Controller App",
    contents: () => fs.readFileSync("../USING_VIRTUAL_CONTROLS.md", { encoding: "utf8" }),
  },
  "setting-up-virtual-controller": {
    title: "Setting up a virtual controller device - TSW Controller App",
    contents: () => fs.readFileSync("../SETTING_UP_VIRTUAL_CONTROLLER.md", { encoding: "utf8" }),
  },
  "profile-explainer": {
    title: "Profile Explainer - TSW Controller App",
    contents: () => fs.readFileSync("../PROFILE_EXPLAINER.md", { encoding: "utf8" }),
  },
  "steam-input-setup": {
    title: "Steam Input Setup - TSW Controller App",
    contents: () => fs.readFileSync("../STEAM_INPUT_SETUP.md", { encoding: "utf8" }),
  },
  "proxy-mode": {
    title: "Proxy Mode - TSW Controller App",
    contents: () => fs.readFileSync("../PROXY_MODE.md", { encoding: "utf8" }),
  },
};
