#!/usr/bin/env node
// API Gateway requires model names to be alphanumeric. swaggo emits
// package-qualified names (e.g. "auth.CreateTokenRequest"), so strip every
// non-alphanumeric character from schema names and their $refs.
const fs = require("fs");

const file = process.argv[2];
if (!file) {
  console.error("usage: node sanitize-openapi.js <openapi.json>");
  process.exit(1);
}

const sanitize = (name) => name.replace(/[^A-Za-z0-9]/g, "");

const doc = JSON.parse(fs.readFileSync(file, "utf8"));

const schemas = doc.components && doc.components.schemas;
if (schemas) {
  const renamed = {};
  for (const key of Object.keys(schemas)) {
    const clean = sanitize(key);
    if (Object.prototype.hasOwnProperty.call(renamed, clean)) {
      console.error(`collision: "${key}" -> "${clean}" already in use`);
      process.exit(1);
    }
    renamed[clean] = schemas[key];
  }
  doc.components.schemas = renamed;
}

const PREFIX = "#/components/schemas/";
const walk = (node) => {
  if (Array.isArray(node)) {
    node.forEach(walk);
  } else if (node && typeof node === "object") {
    for (const [k, v] of Object.entries(node)) {
      if (k === "$ref" && typeof v === "string" && v.startsWith(PREFIX)) {
        node[k] = PREFIX + sanitize(v.slice(PREFIX.length));
      } else {
        walk(v);
      }
    }
  }
};
walk(doc);

// Marca automáticamente como proxy/stream (sin variables) toda operación cuyo
// contrato produzca `text/event-stream` (swag @Produce). La librería
// openapi-to-apigateway-lib (apigw-rest) lee `x-apigw-integration` y genera
// el x-amazon-apigateway-integration acorde (HTTP_PROXY + responseTransferMode
// STREAM). Timeout al máximo permitido por AWS en modo STREAM (900000ms =
// 15 min) por default, configurable via STREAM_INTEGRATION_TIMEOUT_MS.
const streamTimeoutMs = Number(process.env.STREAM_INTEGRATION_TIMEOUT_MS || 900000);
const httpMethods = ["get", "post", "put", "delete", "patch"];
for (const pathItem of Object.values(doc.paths || {})) {
  if (!pathItem) continue;
  for (const method of httpMethods) {
    const operation = pathItem[method];
    const responses = operation && operation.responses;
    if (!responses) continue;

    const isStream = Object.values(responses).some(
      (response) => response && response.content && "text/event-stream" in response.content,
    );

    if (isStream) {
      operation["x-apigw-integration"] = {
        proxy: true,
        stream: true,
        timeoutInMillis: streamTimeoutMs,
      };
    }
  }
}

fs.writeFileSync(file, JSON.stringify(doc, null, 4) + "\n");
console.log(`sanitized model names in ${file}`);
