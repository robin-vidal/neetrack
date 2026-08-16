import type { ComposeSessionRequest, SessionItem } from "../types";
import { apiFetch } from "./client";

/** composeSession asks the server to compose a session via POST /api/sessions. */
export async function composeSession(req: ComposeSessionRequest): Promise<SessionItem[]> {
  const res = await apiFetch("/sessions", {
    method: "POST",
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    throw new Error(`POST /api/sessions failed: ${res.status}`);
  }
  return (await res.json()) as SessionItem[];
}
