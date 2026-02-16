import { useState } from "react";
import {
  ImportSharedProfile,
  LoadConfiguration,
} from "../../../wailsjs/go/main/App";
import { main } from "../../../wailsjs/go/models";
import { alert } from "../../utils/alert";
import { BrowserOpenURL } from "../../../wailsjs/runtime/runtime";

type Props = {
  profile: main.Interop_SharedProfile;
};

export const ExploreTabProfile = ({ profile }: Props) => {
  const [downloading, setIsDownloading] = useState(false);
  const handleDownload = () => {
    setIsDownloading(true);
    ImportSharedProfile(profile)
      .then(() =>
        LoadConfiguration().then(() => alert("Profile Downloaded", "info")),
      )
      .catch((err) => alert(String(err), "error"))
      .finally(() => setIsDownloading(false));
  };

  const handleOpenProfileAuthorUrl = () => {
    if (profile.Author?.Url) {
      BrowserOpenURL(profile.Author.Url);
    }
  };

  return (
    <li className="list-row grid-rows-[max-content,max-content]">
      <div className="list-col-grow flex flex-col gap-2">
        <div>
          <div>{profile.Name}</div>
          {!!profile.Author && (
            <div className="text-sm text-base-content/50">
              {"Created by "}
              {!!profile.Author.Url ? (
                <button className="link" onClick={handleOpenProfileAuthorUrl}>
                  {profile.Author.Name}
                </button>
              ) : (
                profile.Author.Name
              )}
            </div>
          )}
        </div>
        <div className="flex justify-start items-center gap-2 flex-wrap">
          {!!profile.AutoSelect && (
            <div className="badge badge-sm badge-soft badge-info tooltip tooltip-bottom">
              <div className="tooltip-content max-w-3xs ml-10">
                This profile has been configured to support auto-detection.
                This is enabled per locomotive class-type and may not be
                enabled on all variants of the same locomotive.
              </div>
              Supports Auto-Detection
            </div>
          )}
          {!!profile.ContainsCalibration && (
            <div className="badge badge-sm badge-soft badge-info tooltip tooltip-bottom">
              <div className="tooltip-content max-w-3xs ml-10">
                This profile embeds SDL and calibration data which means you
                may not need to manually calibrate your controller after
                importing
              </div>
              Fully Configured
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
      </div>
      <div>
        <button
          className="btn btn-sm btn-primary"
          disabled={downloading}
          onClick={handleDownload}
        >
          {downloading && (
            <span className="loading loading-spinner text-primary"></span>
          )}
          Download
        </button>
      </div>
    </li>
  );
};
