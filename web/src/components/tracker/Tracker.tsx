import type { ReactNode } from "react";

import type { Me } from "../../types";
import { PatternsList } from "./PatternsList";
import "./tracker.css";

interface TrackerProps {
  user: Me;
  logout: ReactNode;
}

/** Tracker renders the authenticated shell: header and the patterns roadmap. */
export function Tracker({ user, logout }: TrackerProps) {
  return (
    <div className="nt">
      <div className="wrap">
        <div className="top">
          <div>
            <div className="eyebrow">Neetcode 150 · rétention espacée</div>
            <h1>Le suivi de {user.name}</h1>
          </div>
          {logout}
        </div>

        <PatternsList />
      </div>
    </div>
  );
}
