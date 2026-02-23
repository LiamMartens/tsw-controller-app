import { ExploreTabProfile } from "./ExploreTabProfile";
import { Suspense, useMemo } from "react";
import { main } from "../../../wailsjs/go/models";
import { BrowserOpenURL } from "../../../wailsjs/runtime/runtime";
import { useControllers, useSharedProfiles } from "../../swr";
import { useForm } from "react-hook-form";

const ExploreTabContent = () => {
  const { register, watch } = useForm<{ query: string }>({
    defaultValues: { query: "" },
  });

  const { data: controllers } = useControllers();
  const { data: sharedProfiles } = useSharedProfiles();

  const handleShare = () => {
    BrowserOpenURL(
      "https://github.com/LiamMartens/tsw-controller-app/issues/new?title=NEW+PROFILE",
    );
  };

  const searchQuery = watch("query");
  const [supportedSharedProfiles, unsupportedSharedProfiles] = useMemo(() => {
    const controllerDeviceIDs = new Set(
      controllers?.filter((c) => !c.IsVirtual).map((c) => c.DeviceID) ?? [],
    );
    const supportedSharedProfiles: main.Interop_SharedProfile[] = [];
    const unsupportedSharedProfiles: main.Interop_SharedProfile[] = [];
    sharedProfiles.forEach((p) => {
      const isSearchMatch =
        searchQuery.trim().length === 0 ||
        p.Name.toLowerCase().includes(searchQuery.toLowerCase());
      if (!isSearchMatch) return;

      (controllerDeviceIDs.has(p.DeviceID)
        ? supportedSharedProfiles
        : unsupportedSharedProfiles
      ).push(p);
    });
    return [supportedSharedProfiles, unsupportedSharedProfiles] as const;
  }, [controllers, sharedProfiles, searchQuery]);

  return (
    <div className="flex flex-col gap-4">
      <div role="alert" className="alert alert-info alert-soft">
        <span>
          Want to share a profile with the world? Submit an "issue" request with
          your profile on Github
          <button className="link ml-2" onClick={handleShare}>
            Submit now
          </button>
        </span>
      </div>
      <div>
        <input
          className="input w-full"
          placeholder="Search for profile(s)"
          {...register("query")}
        />
      </div>
      <div>
        <p className="text-md">Supported Controller Profiles</p>
        <p className="text-sm mb-4 text-gray-400">
          These profiles are available for your currently connected
          controller(s)
        </p>
        {!!supportedSharedProfiles.length && (
          <ul className="list bg-base-100 rounded-box shadow-md">
            {supportedSharedProfiles?.map((profile) => (
              <ExploreTabProfile key={profile.Name} profile={profile} />
            ))}
          </ul>
        )}
        {!supportedSharedProfiles.length && (
          <p className="text-center py-16 text-gray-400">
            No shared profiles for your controller(s)
          </p>
        )}
      </div>

      {!!unsupportedSharedProfiles.length && (
        <div>
          <p className="text-md">Unsupported Controller Profiles</p>
          <p className="text-sm mb-4 text-gray-400">
            These profiles are configured for different controllers and may need
            manual re-configuration
          </p>
          <ul className="list bg-base-100 rounded-box shadow-md">
            {unsupportedSharedProfiles?.map((profile) => (
              <ExploreTabProfile key={profile.Name} profile={profile} />
            ))}
          </ul>
        </div>
      )}
    </div>
  );
};

export const ExploreTab = () => {
  return (
    <Suspense
      fallback={
        <div className="flex justify-center py-6">
          <span className="loading loading-spinner text-primary"></span>
        </div>
      }
    >
      <ExploreTabContent />
    </Suspense>
  );
};
