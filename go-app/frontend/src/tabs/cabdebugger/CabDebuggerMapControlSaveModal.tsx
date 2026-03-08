import { useEffect, useMemo } from "react";
import { Modal, ModalCloseReason, ModalContentProps } from "../../components";
import { ProfileSchema } from "../../profile-schema/schema";
import { useProfiles } from "../../swr";
import { useForm } from "react-hook-form";
import z from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import { alert } from "../../utils/alert";
import {
  LoadConfiguration,
  SaveControlMapping,
} from "../../../wailsjs/go/main/App";
import { config } from "../../../wailsjs/go/models";

type Props = {
  openState: ProfileSchema["controls"][number] | null;
  onClose: () => void;
};

type ContentProps = ModalContentProps<ProfileSchema["controls"][number]>;

type FormValues = z.infer<typeof formSchema>;
const formSchema = z.object({
  profile: z.string(),
  name: z.string().min(1, "Name is required"),
});

export class ProfileSavedModalCloseReason implements ModalCloseReason {}

const CabDebuggerMapControlSaveModalContent = ({
  openState: controlMapping,
  onClose,
}: ContentProps) => {
  const { data: profiles } = useProfiles();
  const form = useForm<FormValues>({
    mode: "all",
    resolver: zodResolver(formSchema),
    defaultValues: {
      profile: "",
      name: "",
    },
  });

  const selectedProfileId = form.watch("profile");
  const selectedProfile = useMemo(
    () => profiles.find((profile) => profile.Id === selectedProfileId) ?? null,
    [selectedProfileId],
  );

  const availableProfiles = useMemo(
    () => profiles.filter((p) => !p.Metadata.IsEmbedded),
    [profiles],
  );

  const handleSave = async () => {
    await form.handleSubmit(async (values) => {
      try {
        const jsonStr = JSON.stringify({
          name: values.name,
          controls: [controlMapping],
        });
        await SaveControlMapping({
          ProfileJSON: jsonStr,
          ExistingPath: selectedProfile?.Metadata.Path ?? "",
        });
        await LoadConfiguration();
        onClose(new ProfileSavedModalCloseReason());
      } catch (err) {
        alert(`Could not save profile (${err})`, "error");
      }
    })();
  };

  useEffect(() => {
    if (selectedProfile) {
      form.setValue("name", selectedProfile?.Name ?? "", {
        shouldDirty: true,
        shouldTouch: true,
        shouldValidate: true,
      });
    }
  }, [form, selectedProfile]);

  return (
    <div className="flex flex-col gap-2">
      <h3 className="font-bold text-base">Save Mapping</h3>
      <div>
        <fieldset className="fieldset">
          <legend className="fieldset-legend">Select Profile</legend>
          <select className="select w-full" {...form.register("profile")}>
            <option value="">New profile</option>
            {availableProfiles.map((profile) => (
              <option key={profile.Id} value={profile.Id}>
                {profile.Name}
              </option>
            ))}
          </select>
          <span className="label">
            Select the profile to save this mapping to
          </span>
        </fieldset>
        <fieldset className="fieldset">
          <legend className="fieldset-legend">Profile Name</legend>
          <input
            type="text"
            className="input w-full"
            disabled={!!selectedProfileId}
            {...form.register("name")}
          />
          {form.formState.errors.name?.message && (
            <span className="label text-error">
              {form.formState.errors.name.message}
            </span>
          )}
        </fieldset>
      </div>
      <div className="flex gap-2 justify-end">
        <form method="dialog">
          <button className="btn btn-sm">Cancel</button>
        </form>
        <button
          className="btn btn-sm"
          disabled={!form.formState.isValid}
          onClick={handleSave}
        >
          Save
        </button>
      </div>
    </div>
  );
};

export const CabDebuggerMapControlSaveModal = ({
  openState,
  onClose,
}: Props) => {
  return (
    <Modal
      openState={openState ?? false}
      Component={CabDebuggerMapControlSaveModalContent}
      onClose={onClose}
    />
  );
};
