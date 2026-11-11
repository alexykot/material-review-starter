import { existsSync } from "node:fs";
import { fileURLToPath } from "node:url";

export const rootEnvPath = fileURLToPath(new URL("../../../.env", import.meta.url));

if (existsSync(rootEnvPath)) process.loadEnvFile(rootEnvPath);

function required(name: string): string {
  const value = process.env[name];
  if (!value) throw new Error(`Missing ${name} in .env (see .env.example)`);
  return value;
}

export const env = {
  databaseUrl: required("DATABASE_URL"),
  apiPort: Number(process.env.API_PORT ?? 8787),
  webOrigin: process.env.WEB_ORIGIN ?? "http://localhost:5180",
  betterAuthUrl: process.env.BETTER_AUTH_URL ?? "http://localhost:8787",
  s3: {
    endpoint: required("S3_ENDPOINT"),
    region: process.env.S3_REGION ?? "us-east-1",
    accessKeyId: required("S3_ACCESS_KEY_ID"),
    secretAccessKey: required("S3_SECRET_ACCESS_KEY"),
    bucket: required("S3_BUCKET"),
  },
  agentId: process.env.AGENT_ID || undefined,
  environmentId: process.env.ENVIRONMENT_ID || undefined,
};
