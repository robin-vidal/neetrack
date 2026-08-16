import { useEffect, useState } from "react";

import type { ActivityResponse } from "../types";
import { apiFetch } from "./client";

type ActivityState =
  | { status: "loading"; activity: null }
  | { status: "ready"; activity: ActivityResponse }
  | { status: "error"; activity: null };

const HEAT_WEEKS = 26;

/** useActivity loads the last HEAT_WEEKS weeks of activity via GET /api/activity. */
export function useActivity() {
  const [state, setState] = useState<ActivityState>({ status: "loading", activity: null });

  useEffect(() => {
    let cancelled = false;

    void (async () => {
      try {
        const res = await apiFetch(`/activity?weeks=${HEAT_WEEKS}`);
        if (!res.ok) {
          if (!cancelled) setState({ status: "error", activity: null });
          return;
        }
        const body = (await res.json()) as ActivityResponse;
        if (!cancelled) setState({ status: "ready", activity: body });
      } catch {
        if (!cancelled) setState({ status: "error", activity: null });
      }
    })();

    return () => {
      cancelled = true;
    };
  }, []);

  return state;
}
