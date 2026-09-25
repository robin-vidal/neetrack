import type { ApiError, ManagedUser } from "../types";
import { apiFetch } from "./client";

/** UsersError represents a French API error returned by user management. */
export class UsersError extends Error {
  constructor(message: string) {
    super(message);
  }
}

async function readError(response: Response): Promise<UsersError> {
  try {
    const body = (await response.json()) as ApiError;
    return new UsersError(body.message);
  } catch {
    return new UsersError("Le serveur ne répond pas.");
  }
}

/** fetchUsers loads account summaries for an administrator. */
export async function fetchUsers(): Promise<ManagedUser[]> {
  const response = await apiFetch("/admin/users");
  if (!response.ok) throw await readError(response);
  const body = (await response.json()) as { users: ManagedUser[] };
  return body.users;
}

/** createUser creates an account through the admin API. */
export async function createUser(input: { name: string; password: string; is_admin: boolean }): Promise<void> {
  const response = await apiFetch("/admin/users", { method: "POST", body: JSON.stringify(input) });
  if (!response.ok) throw await readError(response);
}

/** updateUser applies account changes through the admin API. */
export async function updateUser(id: number, input: { name: string; password?: string; is_admin: boolean }): Promise<void> {
  const response = await apiFetch(`/admin/users/${id}`, { method: "PATCH", body: JSON.stringify(input) });
  if (!response.ok) throw await readError(response);
}

/** deleteUser removes an account through the admin API. */
export async function deleteUser(id: number): Promise<void> {
  const response = await apiFetch(`/admin/users/${id}`, { method: "DELETE" });
  if (!response.ok) throw await readError(response);
}
