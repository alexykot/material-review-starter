import { useState, type FormEvent } from "react";
import { authClient } from "./authClient.ts";

export function LoginPage() {
  const [email, setEmail] = useState("alice@example.com");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();
    setSubmitting(true);
    setError(null);
    const { error } = await authClient.signIn.email({ email, password });
    if (error) setError(error.message ?? "Sign in failed");
    setSubmitting(false);
  }

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
        <button type="submit" disabled={submitting}>
          {submitting ? "Signing in…" : "Sign in"}
        </button>
      </form>
    </main>
  );
}
