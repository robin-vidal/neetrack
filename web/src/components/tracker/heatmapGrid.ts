import type { ActivityDay } from "../../types";

export const HEAT_WEEKS = 26;

export interface HeatDay {
  key: string;
  date: Date;
  future: boolean;
  count: number;
}

function dayKey(d: Date): string {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
}

/** heatWeeks lays activity days out into HEAT_WEEKS columns of 7 days each,
 * ending on the current week's Sunday. */
export function heatWeeks(days: ActivityDay[]): HeatDay[][] {
  const counts = new Map(days.map((d) => [d.date, d.count]));

  const today = new Date();
  today.setHours(0, 0, 0, 0);
  const end = new Date(today);
  end.setDate(end.getDate() + (6 - ((today.getDay() + 6) % 7)));
  const cursor = new Date(end);
  cursor.setDate(cursor.getDate() - (HEAT_WEEKS * 7 - 1));

  const weeks: HeatDay[][] = [];
  for (let w = 0; w < HEAT_WEEKS; w += 1) {
    const week: HeatDay[] = [];
    for (let d = 0; d < 7; d += 1) {
      const date = new Date(cursor);
      const key = dayKey(date);
      week.push({ key, date, future: date > today, count: counts.get(key) ?? 0 });
      cursor.setDate(cursor.getDate() + 1);
    }
    weeks.push(week);
  }
  return weeks;
}

export function heatLevel(count: number): 0 | 1 | 2 | 3 | 4 {
  if (count === 0) return 0;
  if (count === 1) return 1;
  if (count <= 2) return 2;
  if (count <= 4) return 3;
  return 4;
}
