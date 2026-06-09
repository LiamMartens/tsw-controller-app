import { Controller, UseFormReturn } from "react-hook-form";
import { main } from "../../../wailsjs/go/models";
import { useCallback, useMemo } from "react";
import { ProfileSelectionListItem } from "./_components/ProfileSelectionListItem";
import { unfocus } from "../../utils/unfocus";
import {
  ProfileInfo,
  ProfileSelectionMoreMenu,
} from "./_components/ProfileSelectionMoreMenu";
import clsx from "clsx";
import { useSettings } from "../../swr";

type Props = {
  form: UseFormReturn<{
    profiles: Partial<Record<string, main.Interop_SelectedProfileInfo>>;
  }>;
  controller: main.Interop_GenericController;
  profiles: main.Interop_Profile[];
  onReloadConfiguration: () => void;
  onBrowseConfiguration: () => void;
  onCreateProfile: (controller: main.Interop_GenericController | null) => void;
  onSaveProfile: (
    profile: ProfileInfo,
    controller: main.Interop_GenericController | null,
  ) => void;
  onEditProfile: (profile: ProfileInfo) => void;
  onDeleteProfileForController: (profile: ProfileInfo) => void;
};

export function MainTabControllerProfileSelector({
  form,
  controller,
  profiles,
  onReloadConfiguration,
  onBrowseConfiguration,
  onCreateProfile,
  onSaveProfile,
  onEditProfile,
  onDeleteProfileForController,
}: Props) {
  const { watch, control } = form;
  const { data: settings } = useSettings();
  const selectedProfileInfo = watch(`profiles.${controller.UniqueID}`);
  const selectedProfile = useMemo(
    () =>
      (selectedProfileInfo &&
        profiles.find((profile) => profile.Id == selectedProfileInfo?.Id)) ||
      null,
    [selectedProfileInfo],
  );
  const supportedCustomProfiles = useMemo(
    () =>
      profiles?.filter(
        (profile) =>
          !profile.Metadata.IsEmbedded &&
          (!profile.DeviceID || profile.DeviceID === controller.DeviceID),
      ),
    [profiles],
  );
  const supportedEmbeddedProfiles = useMemo(
    () =>
      profiles?.filter(
        (profile) =>
          profile.Metadata.IsEmbedded &&
          (!profile.DeviceID || profile.DeviceID === controller.DeviceID),
      ),
    [profiles],
  );
  const unsupportedProfiles = useMemo(
    () =>
      profiles?.filter(
        (profile) =>
          /* don't display embedded profiles if they are unsupported */
          !profile.Metadata.IsEmbedded &&
          profile.DeviceID &&
          profile.DeviceID !== controller.DeviceID,
      ),
    [profiles],
  );

  return (
    <fieldset className="fieldset w-full">
      <div className="flex flex-col gap-0">
        <label
          htmlFor={`controller_${controller.UniqueID}`}
          className="fieldset-legend pb-1"
        >
          {controller.Name} ({controller.DeviceID})
        </label>

        <p className="label">Unique ID: {controller.UniqueID}</p>
      </div>

      <div className="flex flex-row gap-2 items-center">
        <Controller
          control={control}
          name={`profiles.${controller.UniqueID}`}
          render={({ field }) => (
            <div className="grow dropdown dropdown-start disabled">
              <button
                id={`controller_${controller.UniqueID}`}
                tabIndex={0}
                role="button"
                className={clsx(
                  "select w-full",
                  !controller.IsConfigured && "pointer-events-none opacity-50",
                )}
              >
                {selectedProfile?.Name ?? "Auto-Detect"}
              </button>
              <div className="dropdown-content shadow-sm max-h-[50dvh] overflow-auto w-full">
                <ul className="menu w-full bg-base-300 rounded-box p-2">
                  <li key="auto-detect">
                    <button
                      className="grid grid-cols-1 grid-flow-row auto-rows-max gap-2"
                      onClick={() => {
                        field.onChange(undefined);
                        setTimeout(unfocus, 0);
                      }}
                    >
                      <div>Auto-detect</div>
                    </button>
                  </li>
                  {supportedCustomProfiles.map((profile) => (
                    <ProfileSelectionListItem
                      key={profile.Id}
                      profile={profile}
                      onSelect={field.onChange}
                    />
                  ))}
                  {!!supportedEmbeddedProfiles.length && (
                    <>
                      <div className="divider">Built-In Profiles</div>
                      {supportedEmbeddedProfiles.map((profile) => (
                        <ProfileSelectionListItem
                          key={profile.Id}
                          profile={profile}
                          onSelect={field.onChange}
                        />
                      ))}
                    </>
                  )}
                  {!!unsupportedProfiles.length && (
                    <div>
                      <div className="divider">Unsupported Profiles</div>
                      {unsupportedProfiles.map((profile) => (
                        <ProfileSelectionListItem
                          key={profile.Id}
                          profile={profile}
                          disabled="Not supported by controller"
                        />
                      ))}
                    </div>
                  )}
                </ul>
              </div>
            </div>
          )}
        />

        <ProfileSelectionMoreMenu
          controller={controller}
          profile={selectedProfile ?? null}
          onReloadConfiguration={onReloadConfiguration}
          onBrowseConfiguration={onBrowseConfiguration}
          onCreateProfile={onCreateProfile}
          onSavePofile={onSaveProfile}
          onEditProfile={onEditProfile}
          onDeleteProfile={onDeleteProfileForController}
        />
      </div>

      {!controller.IsConfigured && (
        <div className="alert alert-warning alert-soft p-2">
          <span className="text-xs">
            This controller has not been configured yet. You can download a
            profile which has the "Fully Configured" tag or manually configure
            the controller in the "Calibration" tab.
          </span>
        </div>
      )}
    </fieldset>
  );
}
