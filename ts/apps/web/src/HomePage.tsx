import { useSignOut } from "./session.ts";

export function HomePage({ userName }: { userName: string }) {
  const signOut = useSignOut();
  return (
    <>
      <header className="topbar">
        <strong>Document Review</strong>
        <span>
          Signed in as {userName}
          <button className="secondary" onClick={() => signOut.mutate()} disabled={signOut.isPending}>
            Sign out
          </button>
        </span>
      </header>
      <main className="page" />
    </>
  );
}
