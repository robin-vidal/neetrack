import { useCallback, useEffect, useMemo, useState, type FormEvent } from "react";

import { createUser, deleteUser, fetchUsers, updateUser, UsersError } from "../../api/users";
import type { ManagedUser } from "../../types";
import { Modal } from "../common/Modal";
import { EditIcon, PlusIcon, SearchIcon, TrashIcon } from "./icons";

interface UsersPageProps {
  currentUserId: number;
}

type UserFilter = "all" | "admins" | "members";
type UserDialogState = { mode: "create" } | { mode: "edit"; user: ManagedUser };

/** UsersPage presents account access and the user management actions. */
export function UsersPage({ currentUserId }: UsersPageProps) {
  const [users, setUsers] = useState<ManagedUser[]>([]);
  const [loading, setLoading] = useState(true);
  const [pending, setPending] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [query, setQuery] = useState("");
  const [filter, setFilter] = useState<UserFilter>("all");
  const [dialog, setDialog] = useState<UserDialogState | null>(null);
  const [deleting, setDeleting] = useState<ManagedUser | null>(null);

  const refresh = useCallback(async () => {
    try {
      setUsers(await fetchUsers());
      setError(null);
    } catch (cause) {
      setError(cause instanceof UsersError ? cause.message : "Le serveur ne répond pas.");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  const counts = useMemo(() => ({
    all: users.length,
    admins: users.filter((user) => user.is_admin).length,
    members: users.filter((user) => !user.is_admin).length,
  }), [users]);

  const visibleUsers = useMemo(() => {
    const normalizedQuery = query.trim().toLocaleLowerCase("fr-FR");
    return users.filter((user) => {
      const matchesRole = filter === "all" || (filter === "admins" ? user.is_admin : !user.is_admin);
      return matchesRole && user.name.toLocaleLowerCase("fr-FR").includes(normalizedQuery);
    });
  }, [filter, query, users]);

  const saveUser = async (input: { name: string; password: string; isAdmin: boolean }) => {
    if (!dialog) return;
    setPending(true);
    setError(null);
    try {
      if (dialog.mode === "create") {
        await createUser({ name: input.name, password: input.password, is_admin: input.isAdmin });
      } else {
        await updateUser(dialog.user.id, {
          name: input.name,
          ...(input.password ? { password: input.password } : {}),
          is_admin: input.isAdmin,
        });
        if (input.password && dialog.user.id === currentUserId) {
          window.location.reload();
          return;
        }
      }
      setDialog(null);
      await refresh();
    } catch (cause) {
      setError(cause instanceof UsersError ? cause.message : "Le serveur ne répond pas.");
    } finally {
      setPending(false);
    }
  };

  const confirmDelete = async () => {
    if (!deleting) return;
    setPending(true);
    setError(null);
    try {
      await deleteUser(deleting.id);
      setDeleting(null);
      await refresh();
    } catch (cause) {
      setError(cause instanceof UsersError ? cause.message : "Le serveur ne répond pas.");
      setDeleting(null);
    } finally {
      setPending(false);
    }
  };

  return (
    <section className="admin-users-page" aria-labelledby="admin-users-title">
      <div className="admin-page-heading">
        <div>
          <p className="admin-eyebrow">ACCÈS <span>/</span> COMPTES</p>
          <h1 id="admin-users-title">Utilisateurs</h1>
          <p className="admin-page-description">Crée les comptes et contrôle qui peut administrer Neetrack.</p>
        </div>
        <button type="button" className="admin-primary-button" onClick={() => { setError(null); setDialog({ mode: "create" }); }}>
          <PlusIcon />Nouvel utilisateur
        </button>
      </div>

      <div className="admin-metrics" aria-label="Résumé des comptes">
        <article className="admin-metric-card">
          <span className="admin-metric-icon all"><UsersMetricIcon /></span>
          <div><span>Comptes au total</span><strong>{loading ? "—" : counts.all}</strong></div>
          <small>Accès à Neetrack</small>
        </article>
        <article className="admin-metric-card">
          <span className="admin-metric-icon admin"><ShieldMetricIcon /></span>
          <div><span>Administrateurs</span><strong>{loading ? "—" : counts.admins}</strong></div>
          <small>Gestion des comptes</small>
        </article>
        <article className="admin-metric-card">
          <span className="admin-metric-icon member"><PersonMetricIcon /></span>
          <div><span>Membres</span><strong>{loading ? "—" : counts.members}</strong></div>
          <small>Accès au suivi</small>
        </article>
      </div>

      <section className="admin-table-card" aria-label="Liste des utilisateurs">
        <div className="admin-table-header">
          <div><h2>Comptes</h2><p>Gère les identifiants et les rôles d’accès.</p></div>
          <label className="admin-search"><SearchIcon /><span className="sr-only">Rechercher un utilisateur</span><input value={query} onChange={(event) => setQuery(event.target.value)} placeholder="Rechercher un nom…" /></label>
        </div>
        <div className="admin-filter-row" role="group" aria-label="Filtrer les comptes">
          <button type="button" className={filter === "all" ? "selected" : ""} onClick={() => setFilter("all")}>Tous <span>{counts.all}</span></button>
          <button type="button" className={filter === "admins" ? "selected" : ""} onClick={() => setFilter("admins")}>Administrateurs <span>{counts.admins}</span></button>
          <button type="button" className={filter === "members" ? "selected" : ""} onClick={() => setFilter("members")}>Membres <span>{counts.members}</span></button>
        </div>

        {error && !dialog && <p className="admin-error" role="alert" aria-live="polite">{error}</p>}
        <div className="admin-table-scroll">
          <table className="admin-users-table">
            <thead><tr><th scope="col">Utilisateur</th><th scope="col">Rôle</th><th scope="col">Créé le</th><th scope="col"><span className="sr-only">Actions</span></th></tr></thead>
            <tbody>
              {loading ? <tr><td colSpan={4} className="admin-table-message">Chargement des comptes…</td></tr> : visibleUsers.map((user) => (
                <tr key={user.id}>
                  <td>
                    <div className="admin-account-cell">
                      <span className={`admin-avatar ${user.is_admin ? "is-admin" : ""}`}>{userInitials(user.name)}</span>
                      <div><strong>{user.name}</strong><span>{user.id === currentUserId ? "Votre compte" : `ID ${user.id}`}</span></div>
                    </div>
                  </td>
                  <td><span className={`admin-role ${user.is_admin ? "admin" : "member"}`}><i aria-hidden="true" />{user.is_admin ? "Administrateur" : "Membre"}</span></td>
                  <td className="admin-date">{formatCreatedAt(user.created_at)}</td>
                  <td>
                    <div className="admin-row-actions">
                      <button type="button" className="admin-icon-button" aria-label={`Modifier ${user.name}`} onClick={() => { setError(null); setDialog({ mode: "edit", user }); }}><EditIcon /></button>
                      {user.id !== currentUserId && <button type="button" className="admin-icon-button danger" aria-label={`Supprimer ${user.name}`} onClick={() => { setError(null); setDeleting(user); }}><TrashIcon /></button>}
                    </div>
                  </td>
                </tr>
              ))}
              {!loading && visibleUsers.length === 0 && <tr><td colSpan={4} className="admin-table-message"><strong>{query ? "Aucun résultat" : "Aucun compte dans ce filtre"}</strong><span>{query ? "Essaie un autre nom ou efface la recherche." : "Crée un compte pour commencer."}</span></td></tr>}
            </tbody>
          </table>
        </div>
        <div className="admin-table-footer">{visibleUsers.length} {visibleUsers.length > 1 ? "comptes affichés" : "compte affiché"}<span>{loading ? "Synchronisation…" : "Mis à jour à l’instant"}</span></div>
      </section>

      {dialog && <UserDialog state={dialog} currentUserId={currentUserId} pending={pending} error={error} onClose={() => setDialog(null)} onSave={saveUser} />}
      {deleting && <Modal title="Supprimer ce compte ?" onClose={() => setDeleting(null)} className="admin-confirm-dialog">
        <div className="admin-delete-symbol"><TrashIcon /></div>
        <p>Le compte <strong>{deleting.name}</strong> et tout son historique de suivi seront supprimés. Cette action est définitive.</p>
        {error && <p className="admin-error" role="alert" aria-live="polite">{error}</p>}
        <div className="admin-dialog-actions"><button type="button" className="admin-secondary-button" onClick={() => setDeleting(null)} disabled={pending}>Annuler</button><button type="button" className="admin-danger-button" onClick={() => void confirmDelete()} disabled={pending}>{pending ? "Suppression…" : "Supprimer le compte"}</button></div>
      </Modal>}
    </section>
  );
}

interface UserDialogProps {
  state: UserDialogState;
  currentUserId: number;
  pending: boolean;
  error: string | null;
  onClose: () => void;
  onSave: (input: { name: string; password: string; isAdmin: boolean }) => Promise<void>;
}

function UserDialog({ state, currentUserId, pending, error, onClose, onSave }: UserDialogProps) {
  const initialUser = state.mode === "edit" ? state.user : null;
  const [name, setName] = useState(initialUser?.name ?? "");
  const [password, setPassword] = useState("");
  const [isAdmin, setIsAdmin] = useState(initialUser?.is_admin ?? false);
  const isEditingSelf = initialUser?.id === currentUserId;

  const submit = (event: FormEvent) => {
    event.preventDefault();
    void onSave({ name, password, isAdmin });
  };

  return (
    <Modal title={state.mode === "create" ? "Créer un utilisateur" : "Modifier le compte"} onClose={onClose} className="admin-user-dialog">
      <form onSubmit={submit}>
        <p className="admin-dialog-description">{state.mode === "create" ? "Ajoute un accès au suivi Neetrack." : `Modifie le compte de ${initialUser?.name}.`}</p>
        <label className="admin-field">Nom d’utilisateur<input type="text" value={name} onChange={(event) => setName(event.target.value)} required maxLength={64} autoComplete="off" autoFocus /></label>
        <label className="admin-field">{state.mode === "create" ? "Mot de passe" : "Nouveau mot de passe"}<input type="password" value={password} onChange={(event) => setPassword(event.target.value)} required={state.mode === "create"} minLength={12} autoComplete="new-password" placeholder={state.mode === "edit" ? "Laisser vide pour conserver l’actuel" : "12 caractères minimum"} /></label>
        <label className={`admin-role-switch ${isEditingSelf ? "locked" : ""}`}>
          <input type="checkbox" checked={isAdmin} onChange={(event) => setIsAdmin(event.target.checked)} disabled={isEditingSelf} />
          <span className="admin-switch-track" aria-hidden="true"><i /></span>
          <span><strong>Administrateur</strong><small>{isEditingSelf ? "Ton propre rôle se modifie depuis un autre compte admin." : "Peut gérer les comptes et les accès."}</small></span>
        </label>
        {error && <p className="admin-error" role="alert" aria-live="polite">{error}</p>}
        <div className="admin-dialog-actions"><button type="button" className="admin-secondary-button" onClick={onClose} disabled={pending}>Annuler</button><button type="submit" className="admin-primary-button" disabled={pending}>{pending ? "Enregistrement…" : state.mode === "create" ? "Créer le compte" : "Enregistrer"}</button></div>
      </form>
    </Modal>
  );
}

function userInitials(name: string): string {
  const parts = name.trim().split(/\s+/);
  return parts.length > 1 ? `${parts[0][0]}${parts.at(-1)?.[0]}`.toLocaleUpperCase("fr-FR") : name.slice(0, 2).toLocaleUpperCase("fr-FR");
}

function formatCreatedAt(value: string): string {
  return new Intl.DateTimeFormat("fr-FR", { day: "numeric", month: "short", year: "numeric" }).format(new Date(value));
}

function UsersMetricIcon() {
  return <svg viewBox="0 0 24 24" fill="none" aria-hidden="true"><path d="M16 20v-1.5a3 3 0 0 0-3-3H6a3 3 0 0 0-3 3V20M9.5 11a3.5 3.5 0 1 0 0-7 3.5 3.5 0 0 0 0 7Zm6-6.5a3.5 3.5 0 0 1 0 6.5m5.5 9v-1.5a3 3 0 0 0-2.5-2.9" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"/></svg>;
}

function ShieldMetricIcon() {
  return <svg viewBox="0 0 24 24" fill="none" aria-hidden="true"><path d="M12 3 19 6v5c0 4.5-3 7.8-7 10-4-2.2-7-5.5-7-10V6l7-3Z" stroke="currentColor" strokeWidth="1.8" strokeLinejoin="round"/><path d="m9 12 2 2 4-4" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"/></svg>;
}

function PersonMetricIcon() {
  return <svg viewBox="0 0 24 24" fill="none" aria-hidden="true"><circle cx="12" cy="8" r="3.5" stroke="currentColor" strokeWidth="1.8"/><path d="M5 20a7 7 0 0 1 14 0" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round"/></svg>;
}
