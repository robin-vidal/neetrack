/** Me mirrors the GET /api/me response. */
export interface Me {
  name: string;
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
