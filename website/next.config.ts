import fs from "node:fs";
import $RefParser from "@apidevtools/json-schema-ref-parser";
import type { NextConfig } from "next";

async function config(phase: string, { defaultConfig }: { defaultConfig: NextConfig }) {
  const completeSchema = await $RefParser.dereference(
    "../profile-builder-schema/profile.schema.json",
  );
  const completeSchemaJSON = JSON.stringify(completeSchema);
  fs.writeFileSync("./public/profile-builder/schema.json", completeSchemaJSON);
  fs.writeFileSync(
    "./src/_profile-builder-json-schema/profile.complete.schema.json",
    completeSchemaJSON,
  );

  const nextConfig: NextConfig = {
    ...defaultConfig,
    output: "export",
  };

  return nextConfig;
}

export default config;
