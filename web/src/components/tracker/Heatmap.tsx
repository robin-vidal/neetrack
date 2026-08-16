import { useActivity } from "../../api/useActivity";
import { HEAT_WEEKS, heatLevel, heatWeeks } from "./heatmapGrid";

/** Heatmap renders the activity calendar for the last HEAT_WEEKS weeks. */
export function Heatmap() {
  const state = useActivity();

  if (state.status === "loading") {
    return (
      <div className="heat">
        <p className="muted">Chargement…</p>
      </div>
    );
  }

  if (state.status === "error") {
    return (
      <div className="heat">
        <p className="muted">Impossible de charger l'activité.</p>
      </div>
    );
  }

  const weeks = heatWeeks(state.activity.days);

  return (
    <div className="heat">
      <div className="heat-scroll">
        <div className="heat-inner">
          <div className="heat-months">
            {weeks.map((week, i) => {
              const month = week[0].date.getMonth();
              const prevMonth = i ? weeks[i - 1][0].date.getMonth() : -1;
              const label =
                month !== prevMonth && i < weeks.length - 1
                  ? week[0].date.toLocaleDateString("fr-FR", { month: "short" }).replace(".", "")
                  : "";
              return <span key={week[0].key}>{label}</span>;
            })}
          </div>
          <div className="heat-body">
            <div className="heat-days">
              <span>lun</span>
              <span />
              <span>mer</span>
              <span />
              <span>ven</span>
              <span />
              <span />
            </div>
            {weeks.map((week) => (
              <div className="heat-col" key={week[0].key}>
                {week.map((day) => (
                  <i
                    key={day.key}
                    className={day.future ? "void" : `l${heatLevel(day.count)}`}
                    title={
                      day.future
                        ? ""
                        : `${day.count} exercice${day.count > 1 ? "s" : ""} · ${day.date.toLocaleDateString("fr-FR", { day: "numeric", month: "long" })}`
                    }
                  />
                ))}
              </div>
            ))}
          </div>
        </div>
      </div>
      <div className="heat-foot">
        <span>Moins</span>
        <i />
        <i style={{ background: "#C7E4DD" }} />
        <i style={{ background: "#8ACCBC" }} />
        <i style={{ background: "#3FA391" }} />
        <i style={{ background: "#0E7C6B" }} />
        <span>Plus</span>
        <span style={{ marginLeft: "auto" }}>{HEAT_WEEKS} dernières semaines</span>
      </div>
    </div>
  );
}
