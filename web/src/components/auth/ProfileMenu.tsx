import { useEffect, useRef, useState } from "react";

import type { Me } from "../../types";
import { apiFetch } from "../../api/client";

interface ProfileMenuProps {
  user: Me;
  onDone: () => void;
  onManageUsers: () => void;
}

/** ProfileMenu exposes admin navigation and logout from the profile control. */
export function ProfileMenu({ user, onDone, onManageUsers }: ProfileMenuProps) {
  const [open, setOpen] = useState(false);
  const [pending, setPending] = useState(false);
  const rootRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;

    const closeOutside = (event: PointerEvent) => {
      if (!rootRef.current?.contains(event.target as Node)) setOpen(false);
    };
    const closeOnEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") setOpen(false);
    };
    document.addEventListener("pointerdown", closeOutside);
    document.addEventListener("keydown", closeOnEscape);
    return () => {
      document.removeEventListener("pointerdown", closeOutside);
      document.removeEventListener("keydown", closeOnEscape);
    };
  }, [open]);

  const logout = async () => {
    if (pending) return;
    setPending(true);
    try {
      await apiFetch("/logout", { method: "POST" });
    } finally {
      setPending(false);
      setOpen(false);
      onDone();
    }
  };

  const manageUsers = () => {
    setOpen(false);
    onManageUsers();
  };

  return (
    <div className="profile-menu" ref={rootRef}>
      <button className="who" type="button" aria-expanded={open} aria-controls="profile-actions" onClick={() => setOpen((value) => !value)}>
        <b>{user.name}</b><span aria-hidden="true">⌄</span>
      </button>
      {open && (
        <div className="profile-dropdown" id="profile-actions" aria-label="Menu du profil">
          {user.is_admin && <button type="button" onClick={manageUsers}>Gérer les utilisateurs</button>}
          <button type="button" onClick={() => void logout()} disabled={pending}>{pending ? "Déconnexion…" : "Se déconnecter"}</button>
        </div>
      )}
    </div>
  );
}
