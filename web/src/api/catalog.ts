import type { PatternProblem } from "../types";
import { apiFetch } from "./client";

/** fetchPatternProblems loads one pattern's problems via GET /api/catalog/problems. */
export async function fetchPatternProblems(pattern: string): Promise<PatternProblem[]> {
  const res = await apiFetch(`/catalog/problems?pattern=${encodeURIComponent(pattern)}`);
  if (!res.ok) {
    throw new Error(`GET /api/catalog/problems failed: ${res.status}`);
  }
  return (await res.json()) as PatternProblem[];
}
