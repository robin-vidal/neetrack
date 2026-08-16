const API_BASE = "/api";

/** apiFetch calls a Neetrack API route with JSON headers and same-origin credentials. */
export function apiFetch(path: string, options: RequestInit = {}): Promise<Response> {
  return fetch(`${API_BASE}${path}`, {
    credentials: "same-origin",
    headers: { "Content-Type": "application/json", ...(options.headers ?? {}) },
    ...options,
  });
}
