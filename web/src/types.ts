/** Me mirrors the GET /api/me response. */
export interface Me {
  id: number;
  name: string;
  is_admin: boolean;
}

/** ManagedUser mirrors an account summary returned by the admin API. */
export interface ManagedUser {
  id: number;
  name: string;
  is_admin: boolean;
  created_at: string;
}

/** LoginRequest mirrors the POST /api/login request body. */
export interface LoginRequest {
  name: string;
  password: string;
}

/** ApiError mirrors the error envelope returned by every /api route. */
export interface ApiError {
  error: string;
  message: string;
}

/** PatternProgress mirrors one entry of the GET /api/catalog response. */
export interface PatternProgress {
  name: string;
  total: number;
  done: number;
  due: boolean;
  retrievability: number | null;
}

/** CatalogResponse mirrors the GET /api/catalog response. */
export interface CatalogResponse {
  patterns: PatternProgress[];
}

export type Difficulty = "easy" | "medium" | "hard";

/** ComposeSessionRequest mirrors the POST /api/sessions request body. */
export interface ComposeSessionRequest {
  size: number;
  new_pct: number;
}

/** SessionItem mirrors one entry of the POST /api/sessions response. */
export interface SessionItem {
  slug: string;
  title: string;
  pattern: string;
  difficulty: Difficulty;
  reason: "review" | "new";
  retrievability: number | null;
  neetcode_slug: string;
}

/** CreateAttemptRequest mirrors the POST /api/attempts request body. */
export interface CreateAttemptRequest {
  problem_slug: string;
  duration_seconds?: number;
  used_hint: boolean;
  saw_solution: boolean;
  rating?: number;
}

/** CreateAttemptResponse mirrors the POST /api/attempts response. */
export interface CreateAttemptResponse {
  rating: number;
  pattern: string;
  next_due_at: string;
  interval_days: number;
}

/** ActivityDay mirrors one entry of the GET /api/activity response. */
export interface ActivityDay {
  date: string;
  count: number;
}

/** ActivityResponse mirrors the GET /api/activity response. */
export interface ActivityResponse {
  days: ActivityDay[];
  streak: number;
}

/** UpsertLinkRequest mirrors the PUT /api/links/{slug} request body. */
export interface UpsertLinkRequest {
  neetcode_slug: string;
}

/** Settings mirrors the GET /api/settings response. */
export interface Settings {
  target_retention: number;
  target_date: string | null;
  block_size: number;
  open_max: number;
  new_pct: number;
  session_size: number;
}

/** PatchSettingsRequest mirrors the PATCH /api/settings request body. Only
 * included fields are updated; target_date accepts "" to clear it. */
export type PatchSettingsRequest = Partial<Settings>;

/** PatternProblem mirrors one entry of the GET /api/catalog/problems response. */
export interface PatternProblem {
  slug: string;
  title: string;
  difficulty: Difficulty;
  neetcode_slug: string;
  mastery: number;
}
