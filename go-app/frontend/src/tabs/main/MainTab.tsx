import { lt as semverLt } from "semver";
import {
  LoadConfiguration,
  SelectProfile,
  ClearProfile,
  GetSelectedProfiles,
  InstallTrainSimWorldMod,
  OpenConfigDirectory,
  SetLastInstalledModVersion,
  OpenProfileBuilder,
  DeleteProfile,
  OpenNewProfileBuilder,
  OpenNewProfileBuilderForDeviceID,
  InstallTrainSimClassicMod,
  InstallWondersOfSodorMod,
  SaveProfileForSharing,
  SaveProfileForSharingWithControllerInformation,
  ImportProfile,
} from "../../../wailsjs/go/main/App";
import { useCallback, useEffect, useRef, useState } from "react";
import { BrowserOpenURL, EventsOn } from "../../../wailsjs/runtime/runtime";
import { events } from "../../events";
import { useForm } from "react-hook-form";
import { MainTabControllerProfileSelector } from "./MainTabControllerProfileSelector";
import { main } from "../../../wailsjs/go/models";
import { alert } from "../../utils/alert";
import { confirm } from "../../utils/confirm";
import { ProfileInfo } from "./_components/ProfileSelectionMoreMenu";
import { MainTabProfileSelector } from "./MainTabProfileSelector";
import { ConnectRemoteControllerModal } from "./_components/ConnectRemoteControllerModal";
import {
  useControllers,
  useLastInstalledModVersion,
  useProfiles,
  useVersion,
} from "../../swr";

type FormValues = {
  profiles: Partial<Awaited<ReturnType<typeof GetSelectedProfiles>>>;
};

export const MainTab = () => {
  const { data: version } = useVersion();
  const {
    data: lastInstalledModVersion,
    mutate: refetchLastInstalledModVersion,
  } = useLastInstalledModVersion();
  const { data: profiles, mutate: refetchProfiles } = useProfiles();
  const { data: controllers, mutate: refetchControllers } = useControllers();
  const [connectRemoteControllerOpen, setConnectRemoteControllerOpen] = useState(false);

  const form = useForm<FormValues>({
    defaultValues: { profiles: {} },
  });
  const { watch, getValues } = form;

  const openInWindow = useCallback((url: string) => {
    BrowserOpenURL(url);
  }, []);

  const trySyncSelectedProfiles = useCallback(() => {
    const profiles = getValues("profiles");
    for (const guid in profiles) {
      if (profiles[guid]) {
        SelectProfile(guid, profiles[guid].Id).catch(() => {
          ClearProfile(guid);
          form.setValue(`profiles.${guid}`, undefined);
        });
      } else {
        ClearProfile(guid);
      }
    }
  }, [form]);

  const handleReloadConfiguration = () => {
    LoadConfiguration().then(trySyncSelectedProfiles);
  };

  const handleBrowseConfig = () => {
    OpenConfigDirectory();
  };

  const handleCreateProfile = (
    controller: main.Interop_GenericController | null,
  ) => {
    if (!controller) OpenNewProfileBuilder();
    else OpenNewProfileBuilderForDeviceID(controller.DeviceID);
  };

  const handleEditProfile = (profile: ProfileInfo) => {
    OpenProfileBuilder(profile.Id).catch((err) => alert(String(err), "error"));
  };

  const handleDeleteProfile = (profile: ProfileInfo) => {
    confirm({
      id: "confirm-delete",
      title: "Confirm delete profile?",
      message: "Are you sure you want to delete this profile?",
      actions: ["Cancel", "Confirm"],
      onConfirm: () => {
        DeleteProfile(profile.Id)
          .then(() => {
            const profiles = form.getValues("profiles");
            for (const guid in profiles) {
              form.setValue(`profiles.${guid}`, undefined);
              ClearProfile(guid);
            }
            LoadConfiguration();
          })
          .catch((reason) => alert(String(reason), "error"));
      },
    });
  };

  const handleSaveProfileForSharing = (
    profile: ProfileInfo,
    controller: main.Interop_GenericController | null,
  ) => {
    (controller
      ? SaveProfileForSharingWithControllerInformation(
          profile.Id,
          controller.UniqueID,
        )
      : SaveProfileForSharing(profile.Id)
    ).catch((err) => alert(String(err), "error"));
  };

  const handleConnectRemoteController = () => {
    setConnectRemoteControllerOpen(true);
  };

  const handleInstallTrainSimWorldMod = () => {
    InstallTrainSimWorldMod()
      .then(() => refetchLastInstalledModVersion(version))
      .catch((err) => alert(String(err), "error"));
  };

  const handleInstallTrainSimClassicMod = () => {
    InstallTrainSimClassicMod()
      .then(() => refetchLastInstalledModVersion(version))
      .catch((err) => alert(String(err), "error"));
  };

  const handleInstallWondorsOfSodorMod = () => {
    InstallWondersOfSodorMod()
      .then(() => refetchLastInstalledModVersion(version))
      .catch((err) => alert(String(err), "error"));
  };

  const handleImportProfile = () => {
    ImportProfile()
      .then(() => LoadConfiguration())
      .catch((err) => alert(String(err), "error"));
  };

  const handleIgnoreModInstallWarning = () => {
    SetLastInstalledModVersion(version).then(() => {
      refetchLastInstalledModVersion(version);
    });
  };

  useEffect(() => {
    return watch(trySyncSelectedProfiles).unsubscribe;
  }, [trySyncSelectedProfiles]);

  useEffect(() => {
    GetSelectedProfiles().then((profiles) => form.reset({ profiles }));
  }, [form]);

  useEffect(() => {
    return EventsOn(events.profiles_updated, () => {
      refetchProfiles();
    });
  }, []);

  useEffect(() => {
    return EventsOn(events.joydevices_updated, () => {
      refetchControllers();
    });
  }, []);

  return (
    <div className="grid grid-cols-1 grid-flow-row auto-rows-max gap-2">
      <div role="alert" className="alert alert-info alert-soft">
        <span>
          Want a quick start guide on how to create a profile from scratch?{" "}
          <button
            className="link"
            onClick={() =>
              openInWindow("https://tsw-controller-app.vercel.app/docs")
            }
          >
            Check out the online documentation
          </button>
        </span>
      </div>
      <div>
        {controllers && controllers.length > 1 && (
          <MainTabProfileSelector
            form={form}
            profiles={profiles ?? []}
            controllers={controllers}
            onBrowseConfiguration={handleBrowseConfig}
            onCreateProfile={handleCreateProfile}
            onReloadConfiguration={handleReloadConfiguration}
            onSaveProfile={handleSaveProfileForSharing}
            onEditProfile={handleEditProfile}
            onDeleteProfileForController={handleDeleteProfile}
          />
        )}
        {controllers.map((c) => (
          <div key={c.UniqueID}>
            <MainTabControllerProfileSelector
              controller={c}
              profiles={profiles ?? []}
              form={form}
              onBrowseConfiguration={handleBrowseConfig}
              onCreateProfile={handleCreateProfile}
              onReloadConfiguration={handleReloadConfiguration}
              onSaveProfile={handleSaveProfileForSharing}
              onEditProfile={handleEditProfile}
              onDeleteProfileForController={handleDeleteProfile}
            />
          </div>
        ))}
      </div>
      <p className="text-xs text-base-content/50">
        Note: for auto-detection to work it has to be supported by the profile.
      </p>
      <button
        className="btn btn-sm grow"
        onClick={handleConnectRemoteController}
      >
        + Connect Virtual/Remote Controller
      </button>

      <div className="divider"></div>
      {/* steam://controllerconfig/2967990/3576092503 */}
      <div className="grid grid-cols-2 gap-2">
        <div className="dropdown grow">
          <div tabIndex={0} role="button" className="btn btn-sm w-full">
            Install/Reinstall Game Mod
          </div>
          <ul
            tabIndex={-1}
            className="dropdown-content menu bg-base-300 rounded-box z-1 w-52 p-2 shadow-sm"
          >
            <li>
              <button onClick={handleInstallTrainSimWorldMod}>
                Install TSW mod
              </button>
            </li>
            <li>
              <button onClick={handleInstallTrainSimClassicMod}>
                Install TS Classic mod
              </button>
            </li>
            <li>
              <button onClick={handleInstallWondorsOfSodorMod}>
                Install Wonders of Sodor
              </button>
            </li>
          </ul>
        </div>
        <button className="btn btn-sm grow" onClick={handleImportProfile}>
          Import profile (.tswprofile)
        </button>
      </div>
      {!lastInstalledModVersion && (
        <div role="alert" className="alert alert-soft alert-warning">
          <span>
            It looks like you have not installed the Train Sim World or Train Simulator Clasic mod yet,
            make sure you install the mod first for the best experience.
          </span>
          <div>
            <button
              className="btn btn-sm"
              onClick={handleIgnoreModInstallWarning}
            >
              Ignore
            </button>
          </div>
        </div>
      )}
      {lastInstalledModVersion &&
        semverLt(lastInstalledModVersion, version) && (
          <div role="alert" className="alert alert-soft alert-warning">
            <span>
              It looks like the app has updated since the last time you
              installed the mod, make sure to reinstall the updated mod version
              before starting the game.
            </span>
            <div>
              <button
                className="btn btn-sm"
                onClick={handleIgnoreModInstallWarning}
              >
                Ignore
              </button>
            </div>
          </div>
        )}
      <div role="alert" className="alert">
        <span>
          <strong>Mod Installation Notice</strong>
          <br />
          The mod is not required to install but recommended for full
          compatibility. Without the mod you will not be able to use the
          "direct_control" or "sync_control" control modes. You will have access
          to the "api_control" control mode (Train Sim World only), and any
          regular key bind assignments as long as the TSW API key is configured
          properly (see Settings).
        </span>
      </div>
      <div role="alert" className="alert">
        <span>
          <strong>Controller Setup Notice</strong>
          <br />
          For this app to correctly work you will need to make sure Train Sim
          World is not able to process the controller input. You can achieve
          this by configuring your controller using in Steam using{" "}
          <button
            className="inline link"
            onClick={() =>
              openInWindow(
                "https://github.com/LiamMartens/tsw-controller-app/blob/main/STEAM_INPUT_SETUP.md",
              )
            }
          >
            Steam Input
          </button>{" "}
          and applying the "Disabled Controller" layout preset for the game (see
          "Steam Input" link). Alternatively, you can also use a software like{" "}
          <button
            className="inline link"
            onClick={() =>
              openInWindow("https://ds4-windows.com/download/hidhide/")
            }
          >
            HidHide
          </button>{" "}
          to hide the controller from the game altogether
        </span>
      </div>

      <ConnectRemoteControllerModal
        open={connectRemoteControllerOpen}
        onClose={() => setConnectRemoteControllerOpen(false)}
      />
    </div>
  );
};
