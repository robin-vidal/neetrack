import { useEffect, useState } from "react";

import { patchSettings, SettingsError } from "../../api/settings";
import { useSettings } from "../../api/useSettings";
import type { Settings as SettingsType } from "../../types";

type FormState = {
  targetRetention: number;
  targetDate: string; // "" means no target date
  blockSize: number;
  openMax: number;
  includeHard: boolean;
};

function toFormState(s: SettingsType): FormState {
  return {
    targetRetention: s.target_retention,
    targetDate: s.target_date ?? "",
    blockSize: s.block_size,
    openMax: s.open_max,
    includeHard: s.include_hard,
  };
}

/** Settings lets the user tune the FSRS retention target, target date,
 * learning block size, how many new patterns a session may open and whether
 * hard problems are scheduled. */
export function Settings() {
  const { state, setState } = useSettings();
  const [form, setForm] = useState<FormState | null>(null);
  const [pending, setPending] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    if (state.status === "ready") setForm(toFormState(state.settings));
  }, [state]);

  if (state.status === "loading" || !form) {
    return (
      <div className="panel">
        <p className="muted">Chargement…</p>
      </div>
    );
  }

  if (state.status === "error") {
    return (
      <div className="panel">
        <p className="muted">Impossible de charger les réglages.</p>
      </div>
    );
  }

  const save = async () => {
    if (pending) return;
    setPending(true);
    setError(null);
    setSaved(false);
    try {
      const updated = await patchSettings({
        target_retention: form.targetRetention,
        target_date: form.targetDate,
        block_size: form.blockSize,
        open_max: form.openMax,
        include_hard: form.includeHard,
      });
      setState({ status: "ready", settings: updated });
      setSaved(true);
    } catch (e) {
      setError(e instanceof SettingsError ? e.message : "Le serveur ne répond pas.");
    } finally {
      setPending(false);
    }
  };

  return (
    <div className="panel">
      <label htmlFor="ret" style={{ fontWeight: 500 }}>
        Rétention cible
      </label>
      <p className="muted" style={{ marginTop: 4 }}>
        La probabilité de réussite visée au moment où un pattern revient. Plus haut = intervalles courts et charge
        lourde. Monte vers 0.95 à l'approche des entretiens.
      </p>
      <div style={{ display: "flex", alignItems: "center", gap: 14 }}>
        <input
          id="ret"
          type="range"
          min={70}
          max={99}
          value={Math.round(form.targetRetention * 100)}
          onChange={(e) => setForm({ ...form, targetRetention: Number(e.target.value) / 100 })}
          style={{ flex: 1 }}
        />
        <span className="mono" style={{ width: 46, textAlign: "right" }}>
          {form.targetRetention.toFixed(2)}
        </span>
      </div>

      <hr style={{ border: "none", borderTop: "1px solid var(--rule)", margin: "22px 0" }} />

      <label htmlFor="target-date" style={{ fontWeight: 500 }}>
        Date d'entretien
      </label>
      <p className="muted" style={{ marginTop: 4 }}>
        Si fixée, la rétention effective monte progressivement vers 0.95 dans les 90 jours qui précèdent.
      </p>
      <div style={{ display: "flex", alignItems: "center", gap: 10 }}>
        <input
          id="target-date"
          type="date"
          value={form.targetDate}
          onChange={(e) => setForm({ ...form, targetDate: e.target.value })}
          className="num mono"
          style={{ width: "auto" }}
        />
        {form.targetDate && (
          <button className="mini" onClick={() => setForm({ ...form, targetDate: "" })}>
            Effacer
          </button>
        )}
      </div>

      <hr style={{ border: "none", borderTop: "1px solid var(--rule)", margin: "22px 0" }} />

      <label htmlFor="blk" style={{ fontWeight: 500 }}>
        Taille du bloc d'apprentissage
      </label>
      <p className="muted" style={{ marginTop: 4 }}>
        Quand un pattern est découvert, il est servi en bloc : plusieurs problèmes d'affilée pour laisser le schéma
        se former. Les patterns déjà connus sont ensuite entrelacés, un problème à la fois.
      </p>
      <div style={{ display: "flex", alignItems: "center", gap: 14 }}>
        <input
          id="blk"
          type="range"
          min={1}
          max={6}
          value={form.blockSize}
          onChange={(e) => setForm({ ...form, blockSize: Number(e.target.value) })}
          style={{ flex: 1 }}
        />
        <span className="mono" style={{ width: 46, textAlign: "right" }}>
          {form.blockSize}
        </span>
      </div>

      <hr style={{ border: "none", borderTop: "1px solid var(--rule)", margin: "22px 0" }} />

      <label htmlFor="open-max" style={{ fontWeight: 500 }}>
        Patterns neufs par session
      </label>
      <p className="muted" style={{ marginTop: 4 }}>
        Nombre maximal de patterns jamais amorcés qu'une session peut ouvrir en même temps.
      </p>
      <div style={{ display: "flex", alignItems: "center", gap: 14 }}>
        <input
          id="open-max"
          type="range"
          min={1}
          max={3}
          value={form.openMax}
          onChange={(e) => setForm({ ...form, openMax: Number(e.target.value) })}
          style={{ flex: 1 }}
        />
        <span className="mono" style={{ width: 46, textAlign: "right" }}>
          {form.openMax}
        </span>
      </div>

      <hr style={{ border: "none", borderTop: "1px solid var(--rule)", margin: "22px 0" }} />

      <label htmlFor="exclude-hard" style={{ display: "flex", alignItems: "center", gap: 10, fontWeight: 500 }}>
        <input
          id="exclude-hard"
          type="checkbox"
          checked={!form.includeHard}
          onChange={(e) => setForm({ ...form, includeHard: !e.target.checked })}
        />
        Exclusion des difficiles
      </label>
      <p className="muted" style={{ marginTop: 4, marginBottom: 0 }}>
        Activé, les sessions ne proposent plus aucun problème difficile, ni en découverte ni en révision.
      </p>

      {error && (
        <p role="alert" style={{ color: "var(--sig)", fontSize: 13, margin: "22px 0 0" }}>
          {error}
        </p>
      )}
      {saved && !error && (
        <p role="status" className="muted" style={{ margin: "22px 0 0" }}>
          Réglages enregistrés.
        </p>
      )}

      <button className="btn" style={{ marginTop: 16 }} onClick={save} disabled={pending}>
        {pending ? "Enregistrement…" : "Enregistrer"}
      </button>
    </div>
  );
}
