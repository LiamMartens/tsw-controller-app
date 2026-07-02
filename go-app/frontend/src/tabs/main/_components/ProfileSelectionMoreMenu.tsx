import { useCallback } from "react";
import { main } from "../../../../wailsjs/go/models";
import { unfocus } from "../../../utils/unfocus";

export type ProfileInfo = {
  Id: string;
  Name: string;
};

type Props = {
  controller: main.Interop_GenericController | null;
  profile: main.Interop_Profile | null;
  onReloadConfiguration: () => void;
  onBrowseConfiguration: () => void;
  onCreateProfile: (controller: main.Interop_GenericController | null) => void;
  onSavePofile: (
    profile: ProfileInfo,
    controller: main.Interop_GenericController | null,
  ) => void;
  onEditProfile: (profile: ProfileInfo) => void;
  onRemoveControllerOverride: (profile: ProfileInfo) => void;
  onDeleteProfile: (profile: ProfileInfo) => void;
};

export const ProfileSelectionMoreMenu = ({
  controller,
  profile,
  onReloadConfiguration,
  onBrowseConfiguration,
  onCreateProfile,
  onSavePofile,
  onEditProfile,
  onRemoveControllerOverride,
  onDeleteProfile,
}: Props) => {
  const handleReloadConfiguration = useCallback(() => {
    onReloadConfiguration();
    setTimeout(unfocus, 0);
  }, [onReloadConfiguration]);

  const handleBrowseConfiguration = useCallback(() => {
    onBrowseConfiguration();
    setTimeout(unfocus, 0);
  }, [onBrowseConfiguration]);

  const handleCreateProfile = useCallback(() => {
    onCreateProfile(controller);
    setTimeout(unfocus, 0);
  }, [controller, onCreateProfile]);

  const handleSaveProfile = useCallback(() => {
    if (!profile) return;
    onSavePofile(profile, controller);
    setTimeout(unfocus, 0);
  }, [controller, profile, onSavePofile]);

  const handleEditProfile = useCallback(() => {
    if (!profile) return;
    onEditProfile(profile);
  }, [profile, onEditProfile]);

  const handleRemoveControllverOverride = useCallback(() => {
    if (!profile) return;
    onRemoveControllerOverride(profile);
  }, [profile, onEditProfile]);

  const handleDeleteProfile = useCallback(() => {
    if (!profile) return;
    onDeleteProfile(profile);
  }, [profile, onDeleteProfile]);

  return (
    <div className="dropdown dropdown-end">
      <button
        tabIndex={0}
        role="button"
        className="btn"
        title="Configure Control"
      >
        More
      </button>
      <ul
        tabIndex={-1}
        className="dropdown-content menu bg-base-300 rounded-box z-1 w-52 p-2 shadow-sm"
      >
        <li>
          <button onClick={handleReloadConfiguration}>
            Reload configuration
          </button>
        </li>
        <li>
          <button onClick={handleBrowseConfiguration}>
            Browse configuration
          </button>
        </li>
        <li>
          <button onClick={handleCreateProfile}>Create new profile</button>
        </li>
        <li>
          <button
            disabled={!profile}
            onClick={handleSaveProfile}
            className="disabled:opacity-50 disabled:pointer-events-none"
          >
            Save profile for sharing
          </button>
        </li>
        <li>
          <button
            disabled={!profile}
            onClick={handleEditProfile}
            className="disabled:opacity-50 disabled:pointer-events-none"
          >
            Open profile in builder
          </button>
        </li>
        <li>
          <button
            disabled={
              !profile || !profile.Metadata.ContainsControllerConfiguration
            }
            onClick={handleRemoveControllverOverride}
            className="disabled:opacity-50 disabled:pointer-events-none"
          >
            Remove Controller Override
          </button>
        </li>
        <li>
          <button
            disabled={!profile || profile.Metadata.IsEmbedded}
            onClick={handleDeleteProfile}
            className="disabled:opacity-50 disabled:pointer-events-none"
          >
            Delete profile
          </button>
        </li>
      </ul>
    </div>
  );
};
