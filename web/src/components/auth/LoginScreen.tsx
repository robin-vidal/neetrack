import { useEffect, useRef, useState, type FormEvent } from "react";

import { apiFetch } from "../../api/client";
import type { LoginRequest } from "../../types";
import "./auth.css";

const MESSAGES: Record<number | "offline", string> = {
  401: "Identifiant ou mot de passe incorrect.",
  429: "Trop de tentatives. Attends une minute avant de réessayer.",
  offline: "Le serveur ne répond pas. Vérifie qu'il tourne, puis réessaie.",
};

interface LoginScreenProps {
  onAuthed: () => void;
}

/** LoginScreen renders the credentials form and posts to /api/login. */
export function LoginScreen({ onAuthed }: LoginScreenProps) {
  const [name, setName] = useState("");
  const [password, setPassword] = useState("");
  const [pending, setPending] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const first = useRef<HTMLInputElement>(null);

  useEffect(() => {
    first.current?.focus();
  }, []);

  const submit = async (e: FormEvent) => {
    e.preventDefault();
    if (pending) return;
    if (!name.trim() || !password) {
      setError("Remplis les deux champs.");
      return;
    }
    setPending(true);
    setError(null);
    try {
      const body: LoginRequest = { name: name.trim(), password };
      const res = await apiFetch("/login", {
        method: "POST",
        body: JSON.stringify(body),
      });
      if (res.ok) {
        setPassword("");
        onAuthed();
        return;
      }
      setError(MESSAGES[res.status] ?? `Le serveur a répondu ${res.status}.`);
    } catch {
      setError(MESSAGES.offline);
    } finally {
      setPending(false);
    }
  };

  return (
    <div className="nt nt-auth">
      <div className="auth-wrap">
        <div className="eyebrow">Neetcode 150 · rétention espacée</div>
        <h1>Connexion</h1>
        <p className="muted">Ton suivi, tes cartes et ton historique sont liés à ce compte.</p>

        <form className="auth-form" onSubmit={submit} noValidate>
          <label htmlFor="name">Identifiant</label>
          <input
            id="name"
            ref={first}
            name="username"
            autoComplete="username"
            autoCapitalize="none"
            spellCheck={false}
            value={name}
            onChange={(e) => {
              setName(e.target.value);
              setError(null);
            }}
            aria-invalid={!!error}
          />

          <label htmlFor="password">Mot de passe</label>
          <input
            id="password"
            name="password"
            type="password"
            autoComplete="current-password"
            value={password}
            onChange={(e) => {
              setPassword(e.target.value);
              setError(null);
            }}
            aria-invalid={!!error}
            aria-describedby={error ? "auth-error" : undefined}
          />

          {error && (
            <p className="auth-error" id="auth-error" role="alert">
              {error}
            </p>
          )}

          <button className="btn" type="submit" disabled={pending}>
            {pending ? "Connexion…" : "Se connecter"}
          </button>
        </form>
      </div>
    </div>
  );
}
