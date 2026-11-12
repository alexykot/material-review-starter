import { HomePage } from "./HomePage.tsx";
import { LoginPage } from "./LoginPage.tsx";
import { useMe } from "./session.ts";

export function App() {
  const { data: user, isPending, error } = useMe();

  if (isPending) return <main className="centered">Loading…</main>;
  if (error) return <main className="centered">Could not reach the API: {error.message}</main>;
  if (!user) return <LoginPage />;
  return <HomePage userName={user.name} />;
}
