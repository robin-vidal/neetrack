import type { ApiError, PatchSettingsRequest, Settings } from "../types";
import { apiFetch } from "./client";

export class SettingsError extends Error {
  code: string;

  constructor(body: ApiError) {
    super(body.message);
    this.code = body.error;
  }
}

/** fetchSettings loads the current user's settings via GET /api/settings. */
export async function fetchSettings(): Promise<Settings> {
  const res = await apiFetch("/settings");
  if (!res.ok) {
    const body = (await res.json()) as ApiError;
    throw new SettingsError(body);
  }
  return (await res.json()) as Settings;
}

/** patchSettings applies a partial update via PATCH /api/settings. */
export async function patchSettings(patch: PatchSettingsRequest): Promise<Settings> {
  const res = await apiFetch("/settings", {
    method: "PATCH",
    body: JSON.stringify(patch),
  });
  if (!res.ok) {
    const body = (await res.json()) as ApiError;
    throw new SettingsError(body);
  }
  return (await res.json()) as Settings;
}
