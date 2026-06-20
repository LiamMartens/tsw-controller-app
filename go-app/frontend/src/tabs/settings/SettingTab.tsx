import { useEffect, useMemo, useState } from "react";
import { DeepPartial, useForm } from "react-hook-form";
import {
  SetPreferredControlMode,
  SetTSWAPIKeyLocation,
  SetAlwaysOnTop,
  SetTheme,
  SelectCommAPIKeyFile,
  InstallUdevRules,
} from "../../../wailsjs/go/main/App";
import { alert } from "../../utils/alert";
import { BrowserOpenURL, Environment } from "../../../wailsjs/runtime/runtime";
import { updateTheme } from "../../utils/updateTheme";
import { useControllers, useEnvironment, useSettings } from "../../swr";

type FormValues = {
  tswApiKeyLocation: string;
  preferredControlMode: "direct_control" | "sync_control" | "api_control";
  alwaysOnTop: boolean;
  theme: "system" | "light" | "dark";
};

export const SettingsTab = () => {
  const { data: environment } = useEnvironment();
  const { data: currentSettings, mutate: mutateSettings } = useSettings();
  const { data: controllers } = useControllers();

  const { register, formState, reset, setValue, handleSubmit } =
    useForm<FormValues>({
      defaultValues: currentSettings,
    });

  const shouldShowUdevInstallPrompt = useMemo(
    () => environment.platform === "linux",
    [environment, controllers],
  );

  const handleSelectCommAPIKey = () => {
    SelectCommAPIKeyFile()
      .then((path) => {
        setValue("tswApiKeyLocation", path);
      })
      .catch((err) => alert(String(err), "error"));
  };

  const handleOpenForumLink = () => {
    BrowserOpenURL(
      "https://forums.dovetailgames.com/threads/train-sim-world-api-support.94488/",
    );
  };

  const handleInstallUdevRules = async () => {
    try {
      await InstallUdevRules();
      alert(
        "udev rules installed — your controller should now work without root.",
        "success",
      );
    } catch (err) {
      alert(String(err), "error");
    }
  };

  const handleSubmitSuccess = async (values: FormValues) => {
    const promises: Promise<void>[] = [];
    if (values.theme !== currentSettings.theme) {
      updateTheme(values.theme);
      promises.push(SetTheme(values.theme));
    }

    if (values.tswApiKeyLocation !== currentSettings.tswApiKeyLocation) {
      promises.push(SetTSWAPIKeyLocation(values.tswApiKeyLocation));
    }

    if (
      values.preferredControlMode &&
      values.preferredControlMode !== currentSettings.preferredControlMode
    ) {
      promises.push(SetPreferredControlMode(values.preferredControlMode));
    }

    if (values.alwaysOnTop !== currentSettings.alwaysOnTop) {
      promises.push(SetAlwaysOnTop(values.alwaysOnTop));
    }

    if (promises.length) {
      Promise.all(promises).then(() => {
        reset(values);
        mutateSettings(values);
        alert("Saved settings", "success");
      });
    }
  };

  return (
    <form
      className="grid grid-cols-1 grid-flow-row auto-rows-max gap-2"
      onSubmit={handleSubmit(handleSubmitSuccess)}
    >
      <fieldset className="fieldset">
        <label htmlFor="ui-theme" className="fieldset-legend">
          Theme
        </label>
        <select id="ui-theme" className="select w-full" {...register("theme")}>
          <option value="system">System</option>
          <option value="light">Light</option>
          <option value="dark">Dark</option>
        </select>
      </fieldset>
      <fieldset className="fieldset">
        <label htmlFor="preferred-control-mode" className="fieldset-legend">
          Preferred Control Mode
        </label>
        <select
          id="preferred-control-mode"
          className="select w-full"
          {...register("preferredControlMode")}
        >
          <option value="direct_control">Direct Control</option>
          <option value="sync_control">Sync Control</option>
          <option value="api_control">API Control</option>
        </select>
        <p className="fieldset-label whitespace-normal">
          Sets which control mode to prefer if multiple are defined
        </p>
      </fieldset>
      <fieldset className="fieldset">
        <label htmlFor="tsw-api-key-location" className="fieldset-legend">
          TSW API Key Location
        </label>
        <div className="grid grid-cols-[minmax(0,1fr)_max-content] gap-2">
          <input
            id="tsw-api-key-location"
            className="input w-full"
            {...register("tswApiKeyLocation")}
          />
          <button
            className="btn"
            type="button"
            onClick={handleSelectCommAPIKey}
          >
            Select File
          </button>
        </div>
        <p className="fieldset-label whitespace-normal">
          If the location has not been auto-detected you will need to enter it
          manually here. The API key is only requred for the "api_control"
          control mode.
        </p>
      </fieldset>
      <fieldset className="fieldset bg-base-100 border-base-300 rounded-box border p-4 w-full">
        <legend className="fieldset-legend">Other options</legend>
        <div className="flex gap-4">
          <label className="label">
            <input
              type="checkbox"
              className="checkbox"
              {...register("alwaysOnTop")}
            />
            Always on top
          </label>
        </div>
      </fieldset>

      <div role="alert" className="alert">
        <span>
          <strong>TSW API Notice</strong>
          <br />
          The API connection only works if -HTTPAPI is enabled in Train Sim
          World. You can find instructions in the linked PDF on the{" "}
          <button type="button" className="link" onClick={handleOpenForumLink}>
            online forum
          </button>
          .
        </span>
      </div>

      {shouldShowUdevInstallPrompt && (
        <div role="alert" className="alert">
          <div>
            <p className="mt-1">
              Some custom or complex controllers may require udev rules to be
              installed on linux to grant full HID device access. This requires
              elevated priviliges to install. (this may not be required for your
              set-up - currently this is only required for HID report control)
            </p>
            <button
              className="btn btn-primary btn-xs mt-3"
              type="button"
              onClick={handleInstallUdevRules}
            >
              Install udev rules
            </button>
          </div>
        </div>
      )}

      <div className="flex justify-end">
        <button
          type="submit"
          className="btn btn-primary"
          disabled={
            formState.disabled || !formState.isDirty || formState.isSubmitting
          }
        >
          Save
        </button>
      </div>
    </form>
  );
};
