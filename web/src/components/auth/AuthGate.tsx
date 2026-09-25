import type { ReactNode } from "react";

import { useSession } from "../../api/useSession";
import type { Me } from "../../types";
import "./auth.css";
import { LoginScreen } from "./LoginScreen";
import { ProfileMenu } from "./ProfileMenu";

interface AuthGateProps {
  children: (session: { user: Me; profileMenu: (onManageUsers: () => void) => ReactNode }) => ReactNode;
}

/** AuthGate resolves the session and renders the login screen, an
 * unreachable-server notice, or the authenticated children. */
export function AuthGate({ children }: AuthGateProps) {
  const { state, refresh } = useSession();

  switch (state.status) {
    case "loading":
      return (
        <div className="nt nt-auth">
          <div className="auth-wrap">
            <p className="muted">Chargement…</p>
          </div>
        </div>
      );

    case "unreachable":
      return (
        <div className="nt nt-auth">
          <div className="auth-wrap">
            <h1>Serveur injoignable</h1>
            <p className="muted">L'API ne répond pas sur /api. Vérifie le service, puis recharge.</p>
            <button className="btn" onClick={refresh}>
              Réessayer
            </button>
          </div>
        </div>
      );

    case "out":
      return <LoginScreen onAuthed={refresh} />;

    case "in":
      return (
        <>
          {children({
            user: state.user,
            profileMenu: (onManageUsers) => <ProfileMenu user={state.user} onDone={refresh} onManageUsers={onManageUsers} />,
          })}
        </>
      );
  }
}
