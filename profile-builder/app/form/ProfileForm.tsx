import { FormProvider, useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import type { profile_builder_schema } from "./types";
import { createEmptyProfile } from "./utils";
import { ProfileHeader } from "./ProfileHeader";
import { ControlsList } from "./ControlsList";
import { profileSchema } from "./schema/profileSchema";
import { t } from "../utils";

interface ProfileFormProps {
  initialProfile?: profile_builder_schema | null;
  onSave: (profile: profile_builder_schema) => void;
}

export const ProfileForm = ({ initialProfile, onSave }: ProfileFormProps) => {
  const form = useForm<z.infer<typeof profileSchema>>({
    resolver: zodResolver(profileSchema),
    defaultValues: initialProfile || createEmptyProfile(),
  });

  const handleSubmit = (data: profile_builder_schema) => {
    onSave(data);
  };

  return (
    <FormProvider {...form}>
      <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-6">
        <div className="card bg-base-100 shadow-xl">
          <div className="card-body">
            <h2 className="card-title">{t("Profile Information")}</h2>
            <ProfileHeader form={form} />
          </div>
        </div>
        <div className="card bg-base-100 shadow-xl">
          <div className="card-body">
            <h2 className="card-title">{t("Controls")}</h2>
            <ControlsList form={form} />
          </div>
        </div>
        <div className="flex gap-4">
          <button type="submit" className="btn btn-primary">
            {t("Save Profile")}
          </button>
        </div>
      </form>
    </FormProvider>
  );
};
