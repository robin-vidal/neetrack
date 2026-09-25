import { useCallback, useEffect, useState, type FormEvent } from "react";

import { createUser, deleteUser, fetchUsers, updateUser, UsersError } from "../../api/users";
import type { ManagedUser } from "../../types";
import { Modal } from "../common/Modal";

interface UsersPanelProps {
  currentUserId: number;
}

/** UsersPanel provides administrator-only account management. */
export function UsersPanel({ currentUserId }: UsersPanelProps) {
  const [users, setUsers] = useState<ManagedUser[]>([]);
  const [loading, setLoading] = useState(true);
  const [pending, setPending] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [name, setName] = useState("");
  const [password, setPassword] = useState("");
  const [isAdmin, setIsAdmin] = useState(false);
  const [editing, setEditing] = useState<number | null>(null);
  const [editName, setEditName] = useState("");
  const [editPassword, setEditPassword] = useState("");
  const [editIsAdmin, setEditIsAdmin] = useState(false);
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

  const submitCreate = async (event: FormEvent) => {
    event.preventDefault();
    setPending(true);
    setError(null);
    try {
      await createUser({ name, password, is_admin: isAdmin });
      setName("");
      setPassword("");
      setIsAdmin(false);
      await refresh();
    } catch (cause) {
      setError(cause instanceof UsersError ? cause.message : "Le serveur ne répond pas.");
    } finally {
      setPending(false);
    }
  };

  const beginEdit = (user: ManagedUser) => {
    setEditing(user.id);
    setEditName(user.name);
    setEditPassword("");
    setEditIsAdmin(user.is_admin);
    setError(null);
  };

  const submitEdit = async (event: FormEvent) => {
    event.preventDefault();
    if (editing === null) return;
    setPending(true);
    setError(null);
    try {
      await updateUser(editing, { name: editName, ...(editPassword ? { password: editPassword } : {}), is_admin: editIsAdmin });
      if (editPassword && editing === currentUserId) {
        window.location.reload();
        return;
      }
      setEditing(null);
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
    <section className="users-panel" aria-labelledby="users-heading">
      <div className="users-heading">
        <div>
          <h2 id="users-heading">Utilisateurs</h2>
          <p className="muted">Gère les comptes qui peuvent accéder à Neetrack.</p>
        </div>
      </div>

      <form className="panel users-form" onSubmit={submitCreate}>
        <h3>Ajouter un utilisateur</h3>
        <label>Nom d'utilisateur<input type="text" value={name} onChange={(event) => setName(event.target.value)} required maxLength={64} autoComplete="off" /></label>
        <label>Mot de passe<input type="password" value={password} onChange={(event) => setPassword(event.target.value)} required minLength={12} autoComplete="new-password" /></label>
        <label className="users-check"><input type="checkbox" checked={isAdmin} onChange={(event) => setIsAdmin(event.target.checked)} /> Administrateur</label>
        <button className="btn" type="submit" disabled={pending}>Créer le compte</button>
      </form>

      {error && <p className="users-error" role="alert" aria-live="polite">{error}</p>}
      {loading ? <div className="panel"><p className="muted">Chargement…</p></div> : users.length === 0 ? <div className="panel"><p className="muted">Aucun utilisateur.</p></div> : (
        <div className="users-list">
          {users.map((user) => (
            <article className="panel users-card" key={user.id}>
              {editing === user.id ? (
                <form className="users-edit" onSubmit={submitEdit}>
                  <label>Nom d'utilisateur<input type="text" value={editName} onChange={(event) => setEditName(event.target.value)} required maxLength={64} autoComplete="off" /></label>
                  <label>Nouveau mot de passe<input type="password" value={editPassword} onChange={(event) => setEditPassword(event.target.value)} minLength={12} autoComplete="new-password" placeholder="Laisser vide pour ne pas le changer" /></label>
                  <label className="users-check"><input type="checkbox" checked={editIsAdmin} onChange={(event) => setEditIsAdmin(event.target.checked)} /> Administrateur</label>
                  <div className="users-actions"><button className="btn" type="submit" disabled={pending}>Enregistrer</button><button className="btn ghost" type="button" onClick={() => setEditing(null)} disabled={pending}>Annuler</button></div>
                </form>
              ) : (
                <div className="users-card-line">
                  <div><strong>{user.name}</strong>{user.id === currentUserId && <span className="users-self"> (toi)</span>}<div className="muted">{user.is_admin ? "Administrateur" : "Utilisateur"} · Créé le {new Date(user.created_at).toLocaleDateString("fr-FR")}</div></div>
                  <div className="users-actions"><button className="btn ghost" type="button" onClick={() => beginEdit(user)}>Modifier</button>{user.id !== currentUserId && <button className="btn ghost users-delete" type="button" onClick={() => setDeleting(user)}>Supprimer</button>}</div>
                </div>
              )}
            </article>
          ))}
        </div>
      )}

      {deleting && <Modal title="Supprimer cet utilisateur ?" onClose={() => setDeleting(null)}><p className="muted">Le compte « {deleting.name} » et toutes ses données de suivi seront supprimés. Cette action est définitive.</p><div className="modal-actions"><button className="btn ghost" type="button" onClick={() => setDeleting(null)} disabled={pending}>Annuler</button><button className="btn users-delete" type="button" onClick={() => void confirmDelete()} disabled={pending}>Supprimer</button></div></Modal>}
    </section>
  );
}
