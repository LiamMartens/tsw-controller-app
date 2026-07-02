import { UseFormReturn, useWatch } from "react-hook-form";
import { main } from "../../../wailsjs/go/models";
import { useMemo } from "react";
import { ProfileSelectionListItem } from "./_components/ProfileSelectionListItem";
import { unfocus } from "../../utils/unfocus";
import {
  ProfileInfo,
  ProfileSelectionMoreMenu,
} from "./_components/ProfileSelectionMoreMenu";

type Props = {
  form: UseFormReturn<{
    profiles: Partial<Record<string, main.Interop_SelectedProfileInfo>>;
  }>;
  controllers: main.Interop_GenericController[];
  profiles: main.Interop_Profile[];
  onReloadConfiguration: () => void;
  onBrowseConfiguration: () => void;
  onCreateProfile: (controller: main.Interop_GenericController | null) => void;
  onSaveProfile: (
    profile: ProfileInfo,
    controller: main.Interop_GenericController | null,
  ) => void;
  onEditProfile: (profile: ProfileInfo) => void;
  onRemoveControllerOverride: (profile: ProfileInfo) => void;
  onDeleteProfileForController: (profile: ProfileInfo) => void;
};

export function MainTabProfileSelector({
  form,
  controllers,
  profiles,
  onReloadConfiguration,
  onBrowseConfiguration,
  onCreateProfile,
  onSaveProfile,
  onEditProfile,
  onRemoveControllerOverride,
  onDeleteProfileForController,
}: Props) {
  const { control } = form;
  const selectedProfiles = useWatch({
    name: "profiles",
    control,
  });
  const selectedValue = useMemo(() => {
    const selectedControllerProfiles = controllers.map(
      (controller) => selectedProfiles?.[controller.UniqueID],
    );
    const hasMixedValues = selectedControllerProfiles.every(
      (profile, index, list) =>
        index === 0 || list[index - 1]?.Id !== profile?.Id,
    );
    if (hasMixedValues) return "mixed";
    for (const profileInfo of selectedControllerProfiles) {
      const profile =
        (profileInfo &&
          profiles.find((profile) => profile.Id == profileInfo.Id)) ??
        null;
      if (profile) return profile;
    }
    return undefined;
  }, [controllers, selectedProfiles]);
  const supportedProfiles = useMemo(
    () => profiles?.filter((profile) => !profile.DeviceID),
    [profiles],
  );
  const unsupportedProfiles = useMemo(
    () =>
      profiles?.filter(
        /*
        for selecting profiles across controllers we will always hide embedded profiles or any profile which is for a specific device;
        embedded profiles should really only be used for specific controller support
        */
        (profile) => profile.Metadata.IsEmbedded || !!profile.DeviceID,
      ),
    [profiles],
  );

  const selectProfileForAll = (profile: ProfileInfo | null) => {
    for (const controller of controllers) {
      form.setValue(`profiles.${controller.UniqueID}`, profile ?? undefined);
    }
  };

  return (
    <fieldset className="fieldset w-full">
      <label htmlFor="select-profile" className="fieldset-legend">
        Select profile for all connected controllers
      </label>

      <div className="flex flex-row gap-2 items-center">
        <div className="grow dropdown dropdown-start">
          <button
            id="select-profile"
            tabIndex={0}
            role="button"
            className="select w-full"
          >
            {selectedValue
              ? selectedValue === "mixed"
                ? "Mixed"
                : selectedValue.Name
              : "Auto-detect"}
          </button>
          <div className="dropdown-content shadow-sm max-h-[50dvh] overflow-auto w-full">
            <ul className="menu w-full bg-base-300 rounded-box p-2">
              <li key="auto-detect">
                <button
                  className="grid grid-cols-1 grid-flow-row auto-rows-max gap-2"
                  onClick={() => {
                    selectProfileForAll(null);
                    setTimeout(unfocus, 0);
                  }}
                >
                  <div>Auto-detect</div>
                </button>
              </li>
              {supportedProfiles.map((profile) => (
                <ProfileSelectionListItem
                  key={profile.Id}
                  profile={profile}
                  onSelect={(profile) => selectProfileForAll(profile)}
                />
              ))}
              {!!unsupportedProfiles.length && (
                <div>
                  <div className="divider">Unsupported Profiles</div>
                  {unsupportedProfiles.map((profile) => (
                    <ProfileSelectionListItem
                      key={profile.Id}
                      profile={profile}
                      disabled="Not supported by all controllers"
                    />
                  ))}
                </div>
              )}
            </ul>
          </div>
        </div>

        <ProfileSelectionMoreMenu
          controller={null}
          profile={selectedValue === "mixed" ? null : (selectedValue ?? null)}
          onReloadConfiguration={onReloadConfiguration}
          onBrowseConfiguration={onBrowseConfiguration}
          onCreateProfile={onCreateProfile}
          onSavePofile={onSaveProfile}
          onEditProfile={onEditProfile}
          onRemoveControllerOverride={onRemoveControllerOverride}
          onDeleteProfile={onDeleteProfileForController}
        />
      </div>
    </fieldset>
  );
}
