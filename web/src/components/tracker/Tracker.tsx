import { useState, type ReactNode } from "react";

import type { Me } from "../../types";
import { Heatmap } from "./Heatmap";
import { PatternsList } from "./PatternsList";
import { Session } from "./Session";
import { Settings } from "./Settings";
import { StatsRow } from "./StatsRow";
import "./tracker.css";

interface TrackerProps {
  user: Me;
  logout: ReactNode;
}

type Tab = "patterns" | "session" | "settings";

/** Tracker renders the authenticated shell: header, stats, heatmap, tabs and
 * their content. */
export function Tracker({ user, logout }: TrackerProps) {
  const [tab, setTab] = useState<Tab>("patterns");

  return (
    <div className="nt">
      <div className="wrap">
        <div className="top">
          <div>
            <h1>Le suivi de {user.name}</h1>
          </div>
          {logout}
        </div>

        <StatsRow />
        <Heatmap />

        <div className="tabs" role="tablist">
          <button className="tab" role="tab" aria-selected={tab === "patterns"} onClick={() => setTab("patterns")}>
            Patterns
          </button>
          <button className="tab" role="tab" aria-selected={tab === "session"} onClick={() => setTab("session")}>
            Session
          </button>
          <button className="tab" role="tab" aria-selected={tab === "settings"} onClick={() => setTab("settings")}>
            Réglages
          </button>
        </div>

        {tab === "patterns" && <PatternsList />}
        {tab === "session" && <Session />}
        {tab === "settings" && <Settings />}
      </div>
    </div>
  );
}
