import { useEffect, useState } from "react";

import type { StatsResponse } from "../types";
import { apiFetch } from "./client";

type StatsState =
  | { status: "loading"; stats: null }
  | { status: "ready"; stats: StatsResponse }
  | { status: "error"; stats: null };

/** useStats loads the current user's practice statistics via GET /api/stats. */
export function useStats() {
  const [state, setState] = useState<StatsState>({ status: "loading", stats: null });

  useEffect(() => {
    let cancelled = false;

    void (async () => {
      try {
        const res = await apiFetch("/stats");
        if (!res.ok) {
          if (!cancelled) setState({ status: "error", stats: null });
          return;
        }
        const body = (await res.json()) as StatsResponse;
        if (!cancelled) setState({ status: "ready", stats: body });
      } catch {
        if (!cancelled) setState({ status: "error", stats: null });
      }
    })();

    return () => {
      cancelled = true;
    };
  }, []);

  return state;
}
