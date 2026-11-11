import { betterAuth } from "better-auth";
import { drizzleAdapter } from "better-auth/adapters/drizzle";
import { db } from "./db/client.ts";
import * as schema from "./db/schema.ts";
import { env } from "./env.ts";

export const auth = betterAuth({
  baseURL: env.betterAuthUrl,
  trustedOrigins: [env.webOrigin],
  database: drizzleAdapter(db, { provider: "pg", schema }),
  emailAndPassword: { enabled: true },
});

export type SessionUser = typeof auth.$Infer.Session.user;
