import type { ReactNode } from "react";

import type { Me } from "../../types";
import { UsersPage } from "./UsersPage";
import { ArrowLeftIcon, UsersIcon } from "./icons";
import "./admin.css";

interface AdminDashboardProps {
  user: Me;
  onReturn: () => void;
  profileMenu: ReactNode;
}

/** AdminDashboard provides the reusable navigation frame for administration pages. */
export function AdminDashboard({ user, onReturn, profileMenu }: AdminDashboardProps) {
  return (
    <div className="nt admin-frame">
      <aside className="admin-sidebar" aria-label="Navigation d'administration">
        <div className="admin-brand">
          <div className="admin-brand-mark" aria-hidden="true">N</div>
          <div><strong>neetrack</strong><span>ESPACE ADMIN</span></div>
        </div>

        <div className="admin-nav-group">
          <p>GESTION</p>
          <div className="admin-nav-item active" aria-current="page"><UsersIcon /><span>Utilisateurs</span></div>
        </div>

        <div className="admin-sidebar-bottom">
          <div className="admin-sidebar-user">
            <span className="admin-initials">{user.name.slice(0, 1).toLocaleUpperCase("fr-FR")}</span>
            <div><strong>{user.name}</strong><span>Administrateur</span></div>
          </div>
          <button type="button" className="admin-return" onClick={onReturn}><ArrowLeftIcon />Mon suivi</button>
        </div>
      </aside>

      <main className="admin-main">
        <header className="admin-topbar">
          <div className="admin-breadcrumb"><span>Administration</span><i aria-hidden="true">/</i><strong>Utilisateurs</strong></div>
          {profileMenu}
        </header>
        <div className="admin-content">
          <UsersPage currentUserId={user.id} />
        </div>
      </main>
    </div>
  );
}
