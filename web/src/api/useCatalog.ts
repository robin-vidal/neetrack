import { useEffect, useState } from "react";

import type { CatalogResponse } from "../types";
import { apiFetch } from "./client";

type CatalogState =
  | { status: "loading"; patterns: null }
  | { status: "ready"; patterns: CatalogResponse["patterns"] }
  | { status: "error"; patterns: null };

/** useCatalog loads the pattern roadmap and progress via GET /api/catalog. */
export function useCatalog() {
  const [state, setState] = useState<CatalogState>({ status: "loading", patterns: null });

  useEffect(() => {
    let cancelled = false;

    void (async () => {
      try {
        const res = await apiFetch("/catalog");
        if (!res.ok) {
          if (!cancelled) setState({ status: "error", patterns: null });
          return;
        }
        const body = (await res.json()) as CatalogResponse;
        if (!cancelled) setState({ status: "ready", patterns: body.patterns });
      } catch {
        if (!cancelled) setState({ status: "error", patterns: null });
      }
    })();

    return () => {
      cancelled = true;
    };
  }, []);

  return state;
}
