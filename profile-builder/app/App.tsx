import { ProfileForm } from "./form/ProfileForm";
import type { profile_builder_schema } from "./types/profile_builder_schema";

export const App = () => {
  const handleSave = (profile: profile_builder_schema) => {
    console.log("Saved profile:", JSON.stringify(profile, null, 2));
    alert("Profile saved! Check console for output.");
  };

  return (
    <div className="container mx-auto p-4 max-w-4xl">
      <h1 className="text-2xl font-bold mb-6">
        TSW Controller Profile Builder
      </h1>
      <ProfileForm onSave={handleSave} />
    </div>
  );
};
