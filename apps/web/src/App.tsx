import { authClient } from "./authClient.ts";
import { HomePage } from "./HomePage.tsx";
import { LoginPage } from "./LoginPage.tsx";

export function App() {
  const { data: session, isPending } = authClient.useSession();

  if (isPending) return <main className="centered">Loading…</main>;
  if (!session) return <LoginPage />;
  return <HomePage userName={session.user.name} />;
}
