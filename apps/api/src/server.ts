import { serve } from "@hono/node-server";
import { Hono } from "hono";
import { cors } from "hono/cors";
import { createMiddleware } from "hono/factory";
import { logger } from "hono/logger";
import { auth, type SessionUser } from "./auth.ts";
import { env } from "./env.ts";

type AppEnv = { Variables: { user: SessionUser | null } };

const app = new Hono<AppEnv>();

app.use(logger());
app.use("/api/*", cors({ origin: env.webOrigin, credentials: true }));

app.on(["GET", "POST"], "/api/auth/*", (c) => auth.handler(c.req.raw));

app.use("/api/*", async (c, next) => {
  const session = await auth.api.getSession({ headers: c.req.raw.headers });
  c.set("user", session?.user ?? null);
  await next();
});

const requireUser = createMiddleware<AppEnv>(async (c, next) => {
  if (!c.get("user")) return c.json({ error: "Unauthorized" }, 401);
  await next();
});

app.get("/api/health", (c) => c.json({ ok: true }));

app.get("/api/me", requireUser, (c) => {
  const user = c.get("user")!;
  return c.json({ id: user.id, name: user.name, email: user.email });
});

serve({ fetch: app.fetch, port: env.apiPort }, (info) => {
  console.log(`API listening on http://localhost:${info.port}`);
});
