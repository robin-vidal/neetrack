import type { ApiError, CreateAttemptRequest, CreateAttemptResponse } from "../types";
import { apiFetch } from "./client";

export class AttemptError extends Error {
  code: string;

  constructor(body: ApiError) {
    super(body.message);
    this.code = body.error;
  }
}

/** submitAttempt posts one attempt via POST /api/attempts. */
export async function submitAttempt(req: CreateAttemptRequest): Promise<CreateAttemptResponse> {
  const res = await apiFetch("/attempts", {
    method: "POST",
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    const body = (await res.json()) as ApiError;
    throw new AttemptError(body);
  }
  return (await res.json()) as CreateAttemptResponse;
}
