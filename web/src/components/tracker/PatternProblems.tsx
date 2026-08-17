import { useEffect, useState } from "react";

import { fetchPatternProblems } from "../../api/catalog";
import type { PatternProblem } from "../../types";
import { LinkEditModal } from "./LinkEditModal";
import { DIFF_LABEL, MASTERY_LABEL, leetcodeUrl, neetcodeUrl } from "./sessionFormat";

interface PatternProblemsProps {
  pattern: string;
}

/** PatternProblems lists one pattern's problems with mastery dots, lazily
 * loaded when its row is expanded. */
export function PatternProblems({ pattern }: PatternProblemsProps) {
  const [problems, setProblems] = useState<PatternProblem[] | null>(null);
  const [error, setError] = useState(false);
  const [editingSlug, setEditingSlug] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    void (async () => {
      try {
        const loaded = await fetchPatternProblems(pattern);
        if (!cancelled) setProblems(loaded);
      } catch {
        if (!cancelled) setError(true);
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [pattern]);

  if (error) {
    return (
      <div className="kids">
        <p className="muted">Impossible de charger les problèmes.</p>
      </div>
    );
  }

  if (!problems) {
    return (
      <div className="kids">
        <p className="muted">Chargement…</p>
      </div>
    );
  }

  const editing = editingSlug ? problems.find((p) => p.slug === editingSlug) : undefined;

  return (
    <div className="kids">
      {problems.map((prob) => {
        const [label, pillClass] = DIFF_LABEL[prob.difficulty];
        return (
          <div className="kid" key={prob.slug}>
            <span className="dots" title={MASTERY_LABEL[prob.mastery]}>
              {[1, 2, 3, 4].map((i) => (
                <i key={i} className={prob.mastery === 1 && i === 1 ? "warn" : i <= prob.mastery ? "on" : ""} />
              ))}
            </span>
            <a href={neetcodeUrl(prob.neetcode_slug)} target="_blank" rel="noreferrer">
              {prob.title}
            </a>
            <button
              className="lc"
              style={{ background: "none", border: "none", padding: 0, cursor: "pointer" }}
              onClick={() => setEditingSlug(prob.slug)}
              title="Corriger le lien NeetCode"
            >
              lien
            </button>
            <a className="lc" href={leetcodeUrl(prob.slug)} target="_blank" rel="noreferrer">
              LC
            </a>
            <span className={`pill ${pillClass}`}>{label}</span>
          </div>
        );
      })}

      {editing && (
        <LinkEditModal
          problemSlug={editing.slug}
          currentNeetcodeSlug={editing.neetcode_slug}
          onClose={() => setEditingSlug(null)}
          onSaved={(neetcodeSlug) => {
            setProblems((prev) =>
              prev ? prev.map((p) => (p.slug === editing.slug ? { ...p, neetcode_slug: neetcodeSlug } : p)) : prev,
            );
          }}
        />
      )}
    </div>
  );
}
