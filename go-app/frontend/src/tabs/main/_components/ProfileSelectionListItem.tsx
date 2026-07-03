import { useCallback } from "react";
import { main } from "../../../../wailsjs/go/models";
import { unfocus } from "../../../utils/unfocus";
import clsx from "clsx";

type Props = {
  profile: main.Interop_Profile;
  disabled?: string;
  onSelect?: (profile: main.Interop_Profile) => void;
};

const updatedAtFormatter = new Intl.DateTimeFormat(undefined, {
  dateStyle: "medium",
  timeStyle: "medium",
});

export const ProfileSelectionListItem = ({
  profile,
  disabled,
  onSelect,
}: Props) => {
  const handleClick = useCallback(() => {
    onSelect?.(profile);
    setTimeout(unfocus, 0);
  }, [profile, onSelect]);

  return (
    <li
      className={clsx({
        "menu-disabled": !!disabled,
      })}
    >
      <button
        disabled={!!disabled}
        className="grid grid-cols-1 grid-flow-row auto-rows-max gap-1"
        onClick={handleClick}
      >
        <div>
          <div>{profile.Name}</div>
          <div className="text-base-content/30 text-xs">
            Last updated:{" "}
            {updatedAtFormatter.format(new Date(profile.Metadata.UpdatedAt))}
          </div>
          <div className="text-base-content/30 text-xs">
            {profile.Metadata.Path}
          </div>
          {!!disabled && (
            <span className="text-base-content/30 text-xs">{disabled}</span>
          )}
        </div>
        {!!profile.Metadata.Warnings.length &&
          profile.Metadata.Warnings.map((warning) => (
            <div
              key={warning}
              role="alert"
              className="alert alert-soft alert-warning p-1 text-xs"
            >
              {warning}
            </div>
          ))}

        <div className="flex flex-wrap gap-2 empty:hidden">
          {!!profile.AutoSelect && (
            <div className="badge badge-sm badge-soft badge-info">
              Supports Auto-Select
            </div>
          )}
          {!!profile.Metadata.IsEmbedded && (
            <div className="badge badge-sm badge-soft badge-info">Built-In</div>
          )}
          {profile.Metadata.ContainsControllerConfiguration && (
            <div
              className="tooltip tooltip-bottom tooltip-start"
              data-tip="This profile contains controller mapping information which will override your own controller configuration. To remove it use the 'More' menu."
            >
              <div
                role="alert"
                className="badge badge-sm badge-soft badge-info"
              >
                Contains Controller Mapping Override
              </div>
            </div>
          )}
          {profile.Apps?.map((app) => (
            <div
              key={`app-${app}`}
              className="badge badge-sm badge-soft badge-info"
            >
              {app}
            </div>
          ))}
        </div>
      </button>
    </li>
  );
};
