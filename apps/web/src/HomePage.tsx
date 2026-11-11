import { authClient } from "./authClient.ts";

export function HomePage({ userName }: { userName: string }) {
  return (
    <>
      <header className="topbar">
        <strong>Document Review</strong>
        <span>
          Signed in as {userName}
          <button className="secondary" onClick={() => authClient.signOut()}>
            Sign out
          </button>
        </span>
      </header>
      <main className="page" />
    </>
  );
}
