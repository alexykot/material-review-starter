import { useState, type FormEvent } from "react";
import { isUnauthenticated } from "./api.ts";
import { useSignIn } from "./session.ts";

export function LoginPage() {
  const [email, setEmail] = useState("alice@example.com");
  const [password, setPassword] = useState("");
  const signIn = useSignIn();

  function handleSubmit(event: FormEvent) {
    event.preventDefault();
    signIn.mutate({ email, password });
  }

  const error = signIn.error
    ? isUnauthenticated(signIn.error)
      ? "Invalid email or password"
      : "Sign in failed"
    : null;

  return (
    <main className="centered">
      <form className="card" onSubmit={handleSubmit}>
        <h1>Document Review</h1>
        <label>
          Email
          <input type="email" value={email} onChange={(e) => setEmail(e.target.value)} required />
        </label>
        <label>
          Password
          <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} required />
        </label>
        {error && <p className="error">{error}</p>}
        <button type="submit" disabled={signIn.isPending}>
          {signIn.isPending ? "Signing in…" : "Sign in"}
        </button>
      </form>
    </main>
  );
}
