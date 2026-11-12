import { Code, ConnectError, createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { AuthService } from "@app/api";

// Same-origin: Vite proxies /api to the Go backend in dev, and the load
// balancer routes /api to Cloud Run in production. The session cookie is
// HttpOnly, so the browser sends it automatically.
const transport = createConnectTransport({ baseUrl: "/api" });

export const authClient = createClient(AuthService, transport);

export function isUnauthenticated(error: unknown): boolean {
  return error instanceof ConnectError && error.code === Code.Unauthenticated;
}
