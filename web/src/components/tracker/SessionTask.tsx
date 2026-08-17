import { useEffect, useState } from "react";

import { AttemptError, submitAttempt } from "../../api/attempts";
import type { CreateAttemptResponse, SessionItem } from "../../types";
import { LinkEditModal } from "./LinkEditModal";
import { DIFF_LABEL, FALLBACK_GRADES, TARGET_SECONDS, leetcodeUrl, mmss, neetcodeUrl } from "./sessionFormat";

interface SessionTaskProps {
  item: SessionItem;
  onGraded: (result: CreateAttemptResponse) => void;
  onLinkCorrected: (problemSlug: string, neetcodeSlug: string) => void;
}

/** SessionTask runs one problem's timer and submits its attempt. */
export function SessionTask({ item, onGraded, onLinkCorrected }: SessionTaskProps) {
  const [seconds, setSeconds] = useState(0);
  const [running, setRunning] = useState(false);
  const [usedHint, setUsedHint] = useState(false);
  const [sawSolution, setSawSolution] = useState(false);
  const [pending, setPending] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<CreateAttemptResponse | null>(null);
  const [editingLink, setEditingLink] = useState(false);

  useEffect(() => {
    if (!running) return;
    const id = setInterval(() => setSeconds((s) => s + 1), 1000);
    return () => clearInterval(id);
  }, [running]);

  const submit = async (rating?: number) => {
    if (pending) return;
    setPending(true);
    setError(null);
    try {
      const res = await submitAttempt({
        problem_slug: item.slug,
        duration_seconds: seconds > 0 ? seconds : undefined,
        used_hint: usedHint,
        saw_solution: sawSolution,
        rating,
      });
      setRunning(false);
      setResult(res);
      onGraded(res);
    } catch (e) {
      setError(e instanceof AttemptError ? e.message : "Le serveur ne répond pas.");
    } finally {
      setPending(false);
    }
  };

  const [label, pillClass] = DIFF_LABEL[item.difficulty];
  const target = TARGET_SECONDS[item.difficulty];

  return (
    <div className={`task${result ? " done" : ""}`}>
      <div style={{ display: "flex", alignItems: "baseline", gap: 10, flexWrap: "wrap" }}>
        <span className={`tag ${item.reason}`}>
          {item.reason === "review" && item.retrievability !== null
            ? `Révision · ${Math.round(item.retrievability * 100)}%`
            : "Nouveau"}
        </span>
        <span className="muted" style={{ fontSize: 12 }}>
          {item.pattern}
        </span>
        <span className={`pill ${pillClass}`}>{label}</span>
      </div>

      <div style={{ display: "flex", alignItems: "baseline", gap: 10, marginTop: 6, flexWrap: "wrap" }}>
        <a
          href={neetcodeUrl(item.neetcode_slug)}
          target="_blank"
          rel="noreferrer"
          style={{ fontSize: 16, fontWeight: 500, color: "inherit" }}
        >
          {item.title} ↗
        </a>
        <button className="lc" style={{ background: "none", border: "none", padding: 0, cursor: "pointer" }} onClick={() => setEditingLink(true)}>
          lien
        </button>
        <a className="lc" href={leetcodeUrl(item.slug)} target="_blank" rel="noreferrer">
          LC
        </a>
      </div>

      {editingLink && (
        <LinkEditModal
          problemSlug={item.slug}
          currentNeetcodeSlug={item.neetcode_slug}
          onClose={() => setEditingLink(false)}
          onSaved={(neetcodeSlug) => onLinkCorrected(item.slug, neetcodeSlug)}
        />
      )}

      {result ? (
        <p className="muted" style={{ margin: "10px 0 0" }}>
          {seconds > 0 ? `${mmss(seconds)} · ` : ""}
          noté « {FALLBACK_GRADES.find((g) => g[0] === result.rating)?.[1] ?? result.rating} » · {result.pattern}{" "}
          revient dans {result.interval_days} j
        </p>
      ) : (
        <>
          <div className="clock">
            <b className={seconds > target ? "over" : ""}>{mmss(seconds)}</b>
            <button className="mini" onClick={() => setRunning((r) => !r)} disabled={pending}>
              {running ? "Pause" : seconds ? "Reprendre" : "Démarrer"}
            </button>
            <span className="muted" style={{ fontSize: 12 }}>
              cible {mmss(target)}
            </span>
          </div>

          {seconds > 0 && (
            <div style={{ display: "flex", gap: 8, marginTop: 10, flexWrap: "wrap" }}>
              <button className="mini" aria-pressed={usedHint} onClick={() => setUsedHint((v) => !v)} disabled={pending}>
                {usedHint ? "✓ Indice utilisé" : "Indice utilisé"}
              </button>
              <button
                className="mini"
                aria-pressed={sawSolution}
                onClick={() => setSawSolution((v) => !v)}
                disabled={pending}
              >
                {sawSolution ? "✓ Solution vue" : "Solution vue"}
              </button>
              <button className="btn" onClick={() => submit()} disabled={pending}>
                {pending ? "…" : "Valider"}
              </button>
            </div>
          )}

          {seconds === 0 && (
            <div className="grades">
              {FALLBACK_GRADES.map(([rating, gradeLabel, hint]) => (
                <button className="grade" key={rating} onClick={() => submit(rating)} disabled={pending}>
                  <b>{gradeLabel}</b>
                  <span>{hint}</span>
                </button>
              ))}
            </div>
          )}

          {error && (
            <p className="muted" role="alert" style={{ margin: "10px 0 0", color: "var(--sig)" }}>
              {error}
            </p>
          )}
        </>
      )}
    </div>
  );
}
