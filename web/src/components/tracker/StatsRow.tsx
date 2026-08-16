import { useCatalog } from "../../api/useCatalog";
import { useActivity } from "../../api/useActivity";

/** StatsRow shows the four headline numbers: solved problems, started
 * patterns, patterns due for review, and the current streak. */
export function StatsRow() {
  const catalog = useCatalog();
  const activity = useActivity();

  if (catalog.status !== "ready") {
    return null;
  }

  const totalProblems = catalog.patterns.reduce((sum, p) => sum + p.total, 0);
  const solved = catalog.patterns.reduce((sum, p) => sum + p.done, 0);
  const started = catalog.patterns.filter((p) => p.retrievability !== null).length;
  const due = catalog.patterns.filter((p) => p.due).length;
  const streak = activity.status === "ready" ? activity.activity.streak : 0;

  return (
    <div className="stats">
      <div className="stat">
        <b>
          {solved}
          <span style={{ fontSize: 16, color: "var(--ink-3)" }}>/{totalProblems}</span>
        </b>
        <span>Problèmes vus</span>
      </div>
      <div className="stat">
        <b>
          {started}
          <span style={{ fontSize: 16, color: "var(--ink-3)" }}>/{catalog.patterns.length}</span>
        </b>
        <span>Patterns amorcés</span>
      </div>
      <div className="stat">
        <b style={{ color: due ? "var(--sig)" : undefined }}>{due}</b>
        <span>Patterns à revoir</span>
      </div>
      <div className="stat">
        <b>
          {streak}
          <span style={{ fontSize: 16, color: "var(--ink-3)" }}> j</span>
        </b>
        <span>Série en cours</span>
      </div>
    </div>
  );
}
