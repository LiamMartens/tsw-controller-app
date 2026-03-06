"use client";

import useSWR from "swr";
import jsonSchema from "../_profile-builder-json-schema/profile.complete.schema.json";
import {
  ChangeEventHandler,
  Suspense,
  useCallback,
  useEffect,
  useRef,
  useState,
} from "react";
import { ErrorBoundary } from "react-error-boundary";

declare class JSONEditor {
  constructor(element: HTMLElement, options: Record<string, unknown>);
  getValue(): Record<string, unknown>;
  setValue(input: unknown): void;
  validate(): void;
  showValidationErrors(): void;
}

const useIsClientReady = () => {
  const [isClientReady, setIsClientRaedy] = useState(false);
  useEffect(() => {
    if (typeof window !== "undefined") {
      setIsClientRaedy(true);
    }
  }, []);
  return isClientReady;
};

const useJsonEditorLibrary = () => {
  return useSWR(
    [
      "lib",
      "https://cdn.jsdelivr.net/npm/@json-editor/json-editor@latest/dist/jsoneditor.min.js",
    ],
    async () => {
      return new Promise<HTMLScriptElement>((resolve, reject) => {
        const script =
          document.querySelector<HTMLScriptElement>("script#jsoneditor") ??
          document.createElement("script");
        script.id = "jsoneditor";
        script.onload = () => resolve(script);
        script.onerror = () => reject(new Error("Could not load JSON editor"));
        script.src =
          "https://cdn.jsdelivr.net/npm/@json-editor/json-editor@latest/dist/jsoneditor.min.js";
        document.body.appendChild(script);
      });
    },
    { suspense: true },
  );
};

const ProfileEditorContent = () => {
  useJsonEditorLibrary();

  const editorRef = useRef<JSONEditor | null>(null);

  const handleSave = () => {
    if (!editorRef.current) return;
    const value = editorRef.current.getValue();
    const blob = new Blob([JSON.stringify(value, null, 2)], {
      type: "application/json",
    });
    const url = URL.createObjectURL(blob);
    const downloadLink = document.createElement("a");
    downloadLink.download = `profile.tswprofile`;
    downloadLink.href = url;
    document.body.appendChild(downloadLink);
    downloadLink.click();
    downloadLink.remove();
    URL.revokeObjectURL(url);
  };

  const handleOpen: ChangeEventHandler<HTMLInputElement> = (event) => {
    if (
      !event.currentTarget.files?.length ||
      !event.currentTarget.files[0].name.match(/\.json|\.tswprofile$/)
    ) {
      return;
    }

    const [file] = event.currentTarget.files;
    const reader = new FileReader();
    reader.addEventListener("load", () => {
      const json = JSON.parse(reader.result?.toString() ?? "{}");
      editorRef.current?.setValue(json);
      editorRef.current?.validate();
      editorRef.current?.showValidationErrors();
    });
    reader.readAsText(file);
  };

  const handleContainerRef = useCallback((ref: HTMLElement | null) => {
    if (!ref || typeof JSONEditor === "undefined" || editorRef.current) return;

    const profile_data_raw = new URL(window.location.href).searchParams.get(
      "profile",
    );
    const profile_data = profile_data_raw
      ? JSON.parse(atob(profile_data_raw))
      : {};

    editorRef.current = new JSONEditor(ref, {
      schema: jsonSchema,
      display_required_only: true,
      keep_oneof_values: false,
      theme: "barebones",
      startval: profile_data,
    });
  }, []);

  return (
    <>
      <div id="editor" ref={handleContainerRef} />
      <div className="px-6 mx-auto max-w-4xl sticky bottom-4">
        <div className="bg-base-100 border-base-content/5 border rounded-lg shadow-xl">
          <div className="m-4 flex items-center gap-2">
            <button className="btn btn-primary" onClick={handleSave}>
              Save
            </button>
            <div>
              <label className="btn">
                Open
                <input
                  className="hidden"
                  type="file"
                  accept=".json,.tswprofile"
                  onChange={handleOpen}
                ></input>
              </label>
            </div>
          </div>
        </div>
      </div>
    </>
  );
};

export const ProfileEditor = () => {
  const isClientReady = useIsClientReady();

  return (
    <div className="px-6 mx-auto max-w-4xl">
      <ErrorBoundary
        fallback={
          <div role="alert" className="alert alert-error alert-soft">
            <span>Something went wrong. Please try again later.</span>
          </div>
        }
      >
        <Suspense
          fallback={
            <div className="py-32 flex justify-center items-center">
              <span className="loading loading-dots loading-sm"></span>
            </div>
          }
        >
          {!!isClientReady && <ProfileEditorContent />}
        </Suspense>
      </ErrorBoundary>
    </div>
  );
};
