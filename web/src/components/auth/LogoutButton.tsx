import { useState, type ReactNode } from "react";

import { apiFetch } from "../../api/client";

interface LogoutButtonProps {
  onDone: () => void;
  children?: ReactNode;
}

/** LogoutButton posts to /api/logout, then triggers a session refresh. */
export function LogoutButton({ onDone, children = "Se déconnecter" }: LogoutButtonProps) {
  const [pending, setPending] = useState(false);

  const logout = async () => {
    if (pending) return;
    setPending(true);
    try {
      await apiFetch("/logout", { method: "POST" });
    } finally {
      setPending(false);
      onDone();
    }
  };

  return (
    <button className="who" onClick={logout} disabled={pending}>
      {pending ? "…" : children}
    </button>
  );
}
