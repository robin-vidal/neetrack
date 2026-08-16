import { useState } from "react";

import { composeSession } from "../../api/sessions";
import type { CreateAttemptResponse, SessionItem } from "../../types";
import { SessionTask } from "./SessionTask";

/** Session lets the user configure and run a review session. */
export function Session() {
  const [size, setSize] = useState(5);
  const [newPct, setNewPct] = useState(40);
  const [items, setItems] = useState<SessionItem[] | null>(null);
  const [gradedCount, setGradedCount] = useState(0);
  const [pending, setPending] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const start = async () => {
    if (pending) return;
    setPending(true);
    setError(null);
    try {
      const composed = await composeSession({ size, new_pct: newPct });
      setItems(composed);
      setGradedCount(0);
    } catch {
      setError("Le serveur ne répond pas.");
    } finally {
      setPending(false);
    }
  };

  if (!items) {
    return (
      <div className="panel">
        <div className="setup">
          <label htmlFor="n">Exercices</label>
          <input
            id="n"
            className="num mono"
            type="number"
            min={1}
            max={20}
            value={size}
            onChange={(e) => setSize(Number(e.target.value))}
          />
        </div>
        <div className="setup" style={{ marginTop: 16 }}>
          <label htmlFor="mix">Part de nouveaux</label>
          <input
            id="mix"
            type="range"
            min={0}
            max={100}
            step={10}
            value={newPct}
            onChange={(e) => setNewPct(Number(e.target.value))}
            style={{ flex: 1, minWidth: 140 }}
          />
          <span className="mono" style={{ width: 44, textAlign: "right" }}>
            {newPct}%
          </span>
        </div>

        {error && (
          <p className="muted" role="alert" style={{ margin: "14px 0 0", color: "var(--sig)" }}>
            {error}
          </p>
        )}

        <button className="btn" style={{ marginTop: 16 }} onClick={start} disabled={pending}>
          {pending ? "Composition…" : "Composer la session"}
        </button>
      </div>
    );
  }

  const remaining = items.length - gradedCount;

  const handleGraded = (_result: CreateAttemptResponse) => {
    setGradedCount((c) => c + 1);
  };

  return (
    <div>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 14 }}>
        <span className="muted mono">
          {gradedCount}/{items.length} noté{gradedCount > 1 ? "s" : ""}
        </span>
        <button className="btn ghost" onClick={() => setItems(null)}>
          {remaining ? "Abandonner" : "Terminer"}
        </button>
      </div>
      {items.map((item) => (
        <SessionTask key={item.slug} item={item} onGraded={handleGraded} />
      ))}
    </div>
  );
}
