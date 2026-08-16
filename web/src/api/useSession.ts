import { useCallback, useEffect, useState } from "react";

import type { Me } from "../types";
import { apiFetch } from "./client";

type SessionState =
  | { status: "loading"; user: null }
  | { status: "in"; user: Me }
  | { status: "out"; user: null }
  | { status: "unreachable"; user: null };

/** useSession resolves the current session by calling GET /api/me. */
export function useSession() {
  const [state, setState] = useState<SessionState>({ status: "loading", user: null });

  const refresh = useCallback(async () => {
    try {
      const res = await apiFetch("/me");
      if (res.ok) {
        const user = (await res.json()) as Me;
        setState({ status: "in", user });
      } else if (res.status === 401) {
        setState({ status: "out", user: null });
      } else {
        setState({ status: "unreachable", user: null });
      }
    } catch {
      setState({ status: "unreachable", user: null });
    }
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  return { state, refresh };
}
