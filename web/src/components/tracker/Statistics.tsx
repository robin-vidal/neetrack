import { useStats } from "../../api/useStats";
import type { StatsProblem, StatsResponse } from "../../types";
import { DIFF_LABEL, FALLBACK_GRADES, mmss, neetcodeUrl } from "./sessionFormat";

function percent(rate: number): string {
  return `${Math.round(rate * 100)} %`;
}

function hours(totalSeconds: number): string {
  const totalMinutes = Math.round(totalSeconds / 60);
  if (totalMinutes < 60) return `${totalMinutes} min`;
  const h = Math.floor(totalMinutes / 60);
  const m = (totalMinutes % 60).toString().padStart(2, "0");
  return `${h} h ${m}`;
}

function duration(seconds: number): string {
  return seconds > 0 ? mmss(seconds) : "—";
}

/** Statistics renders the practice statistics returned by GET /api/stats. */
export function Statistics() {
  const state = useStats();

  if (state.status === "loading") {
    return <p className="muted">Chargement…</p>;
  }

  if (state.status === "error") {
    return <p className="muted">Impossible de charger les statistiques.</p>;
  }

  if (state.stats.overview.attempts === 0) {
    return (
      <div className="panel empty-state">
        Aucune tentative pour l'instant. Termine une session pour voir tes statistiques.
      </div>
    );
  }

  return (
    <div>
      <Overview stats={state.stats} />
      <Ratings stats={state.stats} />
      <ByDifficulty stats={state.stats} />
      <ProblemList title="Meilleurs exercices" problems={state.stats.best} />
      <ProblemList title="Exercices à retravailler" problems={state.stats.worst} />
      <ByPattern stats={state.stats} />
    </div>
  );
}

function Overview({ stats }: { stats: StatsResponse }) {
  const o = stats.overview;
  const items: Array<[value: string, label: string]> = [
    [String(o.attempts), "Tentatives"],
    [hours(o.total_seconds), "Temps total"],
    [duration(o.average_seconds), "Temps moyen"],
    [percent(o.success_rate), "Taux de réussite"],
    [String(o.problems_attempted), "Problèmes tentés"],
    [String(o.problems_solved), "Problèmes réussis"],
    [String(o.active_days), "Jours actifs"],
    [`${o.longest_streak} j`, "Plus longue série"],
    [percent(o.hint_rate), "Avec indice"],
    [percent(o.solution_rate), "Solution regardée"],
  ];

  return (
    <section className="panel">
      <h2 className="st-h">Vue d'ensemble</h2>
      <div className="kpis">
        {items.map(([value, label]) => (
          <div className="stat" key={label}>
            <b>{value}</b>
            <span>{label}</span>
          </div>
        ))}
      </div>
    </section>
  );
}

function Ratings({ stats }: { stats: StatsResponse }) {
  const { again, hard, good, easy } = stats.ratings;
  const counts = [again, hard, good, easy];

  return (
    <section className="panel st-section">
      <h2 className="st-h">Répartition des notes</h2>
      <div className="grades" style={{ marginTop: 0 }}>
        {FALLBACK_GRADES.map(([rating, label], i) => (
          <div className="grade" key={rating}>
            <b>{label}</b>
            <span className="mono st-big">{counts[i]}</span>
            <span>{percent(counts[i] / stats.overview.attempts)}</span>
          </div>
        ))}
      </div>
    </section>
  );
}

function ByDifficulty({ stats }: { stats: StatsResponse }) {
  return (
    <section className="panel st-section">
      <h2 className="st-h">Par difficulté</h2>
      <div className="st-scroll">
        <table className="st-table">
          <thead>
            <tr>
              <th>Difficulté</th>
              <th>Tentatives</th>
              <th>Temps moyen</th>
              <th>Cible</th>
              <th>Réussite</th>
            </tr>
          </thead>
          <tbody>
            {stats.by_difficulty.map((d) => {
              const [label, pillClass] = DIFF_LABEL[d.difficulty];
              const over = d.average_seconds > d.target_seconds;
              return (
                <tr key={d.difficulty}>
                  <td>
                    <span className={`pill ${pillClass}`}>{label}</span>
                  </td>
                  <td className="mono">{d.attempts}</td>
                  <td className="mono" style={{ color: over ? "var(--sig)" : undefined }}>
                    {duration(d.average_seconds)}
                  </td>
                  <td className="mono">{mmss(d.target_seconds)}</td>
                  <td className="mono">{d.attempts > 0 ? percent(d.success_rate) : "—"}</td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </section>
  );
}

function ProblemList({ title, problems }: { title: string; problems: StatsProblem[] }) {
  if (problems.length === 0) return null;

  return (
    <section className="panel st-section">
      <h2 className="st-h">{title}</h2>
      <div className="st-scroll">
        <table className="st-table">
          <thead>
            <tr>
              <th>Exercice</th>
              <th>Note moy.</th>
              <th>Temps moyen</th>
              <th>Tentatives</th>
            </tr>
          </thead>
          <tbody>
            {problems.map((p) => {
              const [label, pillClass] = DIFF_LABEL[p.difficulty];
              return (
                <tr key={p.slug}>
                  <td>
                    <span className={`pill ${pillClass}`}>{label}</span>{" "}
                    <a href={neetcodeUrl(p.neetcode_slug)} target="_blank" rel="noreferrer">
                      {p.title}
                    </a>
                    <span className="st-sub">{p.pattern}</span>
                  </td>
                  <td className="mono">{p.average_rating.toFixed(1)}/4</td>
                  <td className="mono" style={{ color: p.pace > 1 ? "var(--sig)" : undefined }}>
                    {duration(p.average_seconds)}
                  </td>
                  <td className="mono">{p.attempts}</td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </section>
  );
}

function ByPattern({ stats }: { stats: StatsResponse }) {
  return (
    <section className="panel st-section">
      <h2 className="st-h">Par pattern</h2>
      <div className="st-scroll">
        <table className="st-table">
          <thead>
            <tr>
              <th>Pattern</th>
              <th>Tentatives</th>
              <th>Problèmes</th>
              <th>Note moy.</th>
              <th>Réussite</th>
              <th>Temps moyen</th>
            </tr>
          </thead>
          <tbody>
            {stats.by_pattern.map((p) => (
              <tr key={p.pattern}>
                <td>{p.pattern}</td>
                <td className="mono">{p.attempts}</td>
                <td className="mono">{p.problems_attempted}</td>
                <td className="mono">{p.average_rating.toFixed(1)}/4</td>
                <td className="mono">{percent(p.success_rate)}</td>
                <td className="mono">{duration(p.average_seconds)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  );
}
