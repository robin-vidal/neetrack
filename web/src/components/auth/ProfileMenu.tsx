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

  const initials = user.name.slice(0, 1).toLocaleUpperCase("fr-FR");

  return (
    <div className="profile-menu" ref={rootRef}>
      <button className="who" type="button" aria-expanded={open} aria-controls="profile-actions" onClick={() => setOpen((value) => !value)}>
        <span className="profile-avatar" aria-hidden="true">{initials}</span><b>{user.name}</b>
        <svg className={`profile-chevron ${open ? "open" : ""}`} viewBox="0 0 20 20" fill="none" aria-hidden="true"><path d="m5.5 7.5 4.5 4.5 4.5-4.5" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round" /></svg>
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
