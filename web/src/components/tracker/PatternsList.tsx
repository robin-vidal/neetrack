import { useCatalog } from "../../api/useCatalog";

/** PatternsList renders one retention bar per pattern from GET /api/catalog. */
export function PatternsList() {
  const state = useCatalog();

  if (state.status === "loading") {
    return <p className="muted">Chargement…</p>;
  }

  if (state.status === "error") {
    return <p className="muted">Impossible de charger le catalogue.</p>;
  }

  return (
    <div>
      {state.patterns.map((p) => (
        <div className="row" key={p.name}>
          <span className="pname">{p.name}</span>
          {p.retrievability !== null ? (
            <span className="bar">
              <i
                style={{
                  width: `${Math.round(p.retrievability * 100)}%`,
                  background: p.due ? "var(--sig)" : "var(--ok)",
                }}
              />
            </span>
          ) : (
            <span className="bar empty" />
          )}
          <span className="pct mono">
            {p.retrievability !== null ? `${Math.round(p.retrievability * 100)}%` : "—"}
          </span>
          <span className="count mono">
            {p.done}/{p.total}
          </span>
        </div>
      ))}
    </div>
  );
}
