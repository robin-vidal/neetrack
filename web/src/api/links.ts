import type { ApiError, UpsertLinkRequest } from "../types";
import { apiFetch } from "./client";

export class LinkError extends Error {
  code: string;

  constructor(body: ApiError) {
    super(body.message);
    this.code = body.error;
  }
}

/** upsertLink corrects a problem's NeetCode slug via PUT /api/links/{slug}. */
export async function upsertLink(problemSlug: string, neetcodeSlug: string): Promise<void> {
  const req: UpsertLinkRequest = { neetcode_slug: neetcodeSlug };
  const res = await apiFetch(`/links/${encodeURIComponent(problemSlug)}`, {
    method: "PUT",
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    const body = (await res.json()) as ApiError;
    throw new LinkError(body);
  }
}
