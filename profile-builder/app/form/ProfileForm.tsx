import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import type { profile_builder_schema } from "./types";
import { createEmptyProfile } from "./utils";
import { ProfileHeader } from "./ProfileHeader";
import { profileSchema } from "./schema/profileSchema";
import { t } from "../utils";
import { useEffect } from "react";
import { ControlsList } from "./ControlsList";

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

  useEffect(() => {
    form.watch(console.log);
  }, [form]);

  return (
    <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-6">
      <div className="card bg-base-100 border border-base-300 shadow-xl">
        <div className="card-body">
          <h2 className="card-title">{t("Profile Information")}</h2>
          <ProfileHeader form={form} />
        </div>
      </div>

      <div className="card bg-base-100 border border-base-300 shadow-xl">
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
  );
};
