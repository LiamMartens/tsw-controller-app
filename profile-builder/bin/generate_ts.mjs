import fs from "node:fs";
import ts from "typescript";
import schema from "../app/schema.json" with { type: "json" };
import openapiTS, { astToString } from "openapi-typescript";

/** @type {(node: ts.Node) => node is ts.InterfaceDeclaration} */
const isInterfaceDeclaration = (node) => node.kind == ts.SyntaxKind.InterfaceDeclaration;

/** @type {(node: ts.Node) => node is ts.PropertySignature} */
const isPropertySignature = (node) => node.kind == ts.SyntaxKind.PropertySignature;

const ast = await openapiTS({
  openapi: "3.1",
  components: {
    schemas: {
      schema,
    },
  },
});

/** @type {ts.InterfaceDeclaration} */
const components = ast.find(
  (node) => isInterfaceDeclaration(node) && node.name.text === "components",
);

/** @type {ts.PropertySignature} */
const schemas = components.members.find(
  (member) => isPropertySignature(member) && member.name.text === "schemas",
);

components.name = ts.factory.createIdentifier("profile_builder_schema");
components.members = schemas.type.members[0].type.members;

fs.writeFileSync("./app/profile_builder_schema.d.ts", astToString(components));
