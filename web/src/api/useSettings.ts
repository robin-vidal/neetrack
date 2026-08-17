import { useEffect, useState } from "react";

import { fetchSettings } from "./settings";
import type { Settings } from "../types";

type SettingsState =
  | { status: "loading"; settings: null }
  | { status: "ready"; settings: Settings }
  | { status: "error"; settings: null };

/** useSettings loads the current user's settings via GET /api/settings. */
export function useSettings() {
  const [state, setState] = useState<SettingsState>({ status: "loading", settings: null });

  useEffect(() => {
    let cancelled = false;

    void (async () => {
      try {
        const settings = await fetchSettings();
        if (!cancelled) setState({ status: "ready", settings });
      } catch {
        if (!cancelled) setState({ status: "error", settings: null });
      }
    })();

    return () => {
      cancelled = true;
    };
  }, []);

  return { state, setState };
}
