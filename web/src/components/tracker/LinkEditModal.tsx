import { useState, type FormEvent } from "react";

import { LinkError, upsertLink } from "../../api/links";
import { Modal } from "../common/Modal";

interface LinkEditModalProps {
  problemSlug: string;
  currentNeetcodeSlug: string;
  onClose: () => void;
  onSaved: (neetcodeSlug: string) => void;
}

/** LinkEditModal corrects a problem's NeetCode slug via PUT /api/links/{slug}. */
export function LinkEditModal({ problemSlug, currentNeetcodeSlug, onClose, onSaved }: LinkEditModalProps) {
  const [value, setValue] = useState(currentNeetcodeSlug);
  const [pending, setPending] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const submit = async (e: FormEvent) => {
    e.preventDefault();
    const trimmed = value.trim();
    if (!trimmed || pending) return;

    setPending(true);
    setError(null);
    try {
      await upsertLink(problemSlug, trimmed);
      onSaved(trimmed);
      onClose();
    } catch (err) {
      setError(err instanceof LinkError ? err.message : "Le serveur ne répond pas.");
    } finally {
      setPending(false);
    }
  };

  return (
    <Modal title="Corriger le lien NeetCode" onClose={onClose}>
      <form onSubmit={submit} noValidate>
        <label htmlFor="neetcode-slug">Slug ou URL NeetCode</label>
        <input
          id="neetcode-slug"
          type="text"
          value={value}
          onChange={(e) => setValue(e.target.value)}
          aria-invalid={!!error}
          aria-describedby={error ? "neetcode-slug-error" : undefined}
        />

        {error && (
          <p id="neetcode-slug-error" role="alert" style={{ color: "var(--sig)", fontSize: 13, margin: "-8px 0 16px" }}>
            {error}
          </p>
        )}

        <div className="modal-actions">
          <button type="button" className="btn ghost" onClick={onClose} disabled={pending}>
            Annuler
          </button>
          <button type="submit" className="btn" disabled={pending || !value.trim()}>
            {pending ? "…" : "Enregistrer"}
          </button>
        </div>
      </form>
    </Modal>
  );
}
