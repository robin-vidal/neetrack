import React, { useState, useEffect, useCallback, useRef } from "react";

/*
 * Écran de login pour la version self-host.
 *
 * Contrat attendu côté serveur :
 *   GET  /api/me      → 200 {"name":"robin"} si session valide, sinon 401
 *   POST /api/login   → 204 + Set-Cookie de session, 401 si mauvais identifiants
 *   POST /api/logout  → 204 et cookie invalidé côté serveur
 *
 * Le cookie de session est httpOnly : ce fichier ne le lit jamais et ne stocke
 * aucun jeton. Le mot de passe ne vit que dans le champ contrôlé, le temps du
 * submit.
 */

const API = "/api";

async function call(path, options = {}) {
  return fetch(`${API}${path}`, {
    credentials: "same-origin",
    headers: { "Content-Type": "application/json", ...(options.headers || {}) },
    ...options,
  });
}

/* -------------------------------------------------------------- session --- */

export function useSession() {
  const [state, setState] = useState({ status: "loading", user: null });

  const refresh = useCallback(async () => {
    try {
      const res = await call("/me");
      if (res.ok) setState({ status: "in", user: await res.json() });
      else if (res.status === 401) setState({ status: "out", user: null });
      else setState({ status: "unreachable", user: null });
    } catch {
      setState({ status: "unreachable", user: null });
    }
  }, []);

  useEffect(() => { refresh(); }, [refresh]);

  return { ...state, refresh };
}

/* ---------------------------------------------------------------- login --- */

const MESSAGES = {
  401: "Identifiant ou mot de passe incorrect.",
  429: "Trop de tentatives. Attends une minute avant de réessayer.",
  offline: "Le serveur ne répond pas. Vérifie qu'il tourne, puis réessaie.",
};

export function LoginScreen({ onAuthed }) {
  const [name, setName] = useState("");
  const [password, setPassword] = useState("");
  const [pending, setPending] = useState(false);
  const [error, setError] = useState(null);
  const first = useRef(null);

  useEffect(() => { first.current?.focus(); }, []);

  const submit = async (e) => {
    e.preventDefault();
    if (pending) return;
    if (!name.trim() || !password) {
      setError("Remplis les deux champs.");
      return;
    }
    setPending(true);
    setError(null);
    try {
      const res = await call("/login", {
        method: "POST",
        body: JSON.stringify({ name: name.trim(), password }),
      });
      if (res.ok) {
        setPassword("");
        onAuthed?.();
        return;
      }
      setError(MESSAGES[res.status] || `Le serveur a répondu ${res.status}.`);
    } catch {
      setError(MESSAGES.offline);
    } finally {
      setPending(false);
    }
  };

  return (
    <div className="nt nt-auth">
      <style>{AUTH_CSS}</style>
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
            spellCheck="false"
            value={name}
            onChange={(e) => { setName(e.target.value); setError(null); }}
            aria-invalid={!!error}
          />

          <label htmlFor="password">Mot de passe</label>
          <input
            id="password"
            name="password"
            type="password"
            autoComplete="current-password"
            value={password}
            onChange={(e) => { setPassword(e.target.value); setError(null); }}
            aria-invalid={!!error}
            aria-describedby={error ? "auth-error" : undefined}
          />

          {error && <p className="auth-error" id="auth-error" role="alert">{error}</p>}

          <button className="btn" type="submit" disabled={pending}>
            {pending ? "Connexion…" : "Se connecter"}
          </button>
        </form>
      </div>
    </div>
  );
}

/* --------------------------------------------------------------- logout --- */

export function LogoutButton({ onDone, children = "Se déconnecter" }) {
  const [pending, setPending] = useState(false);

  const logout = async () => {
    if (pending) return;
    setPending(true);
    try {
      await call("/logout", { method: "POST" });
    } catch {
      /* le cookie est invalidé côté serveur ; on redescend quand même */
    } finally {
      setPending(false);
      onDone?.();
    }
  };

  return (
    <button className="who" onClick={logout} disabled={pending}>
      {pending ? "…" : children}
    </button>
  );
}

/* ------------------------------------------------------------------ gate --- */

export function AuthGate({ children }) {
  const { status, user, refresh } = useSession();

  if (status === "loading") {
    return (
      <div className="nt nt-auth">
        <style>{AUTH_CSS}</style>
        <div className="auth-wrap"><p className="muted">Chargement…</p></div>
      </div>
    );
  }

  if (status === "unreachable") {
    return (
      <div className="nt nt-auth">
        <style>{AUTH_CSS}</style>
        <div className="auth-wrap">
          <h1>Serveur injoignable</h1>
          <p className="muted">L'API ne répond pas sur {API}. Vérifie le service, puis recharge.</p>
          <button className="btn" onClick={refresh}>Réessayer</button>
        </div>
      </div>
    );
  }

  if (status === "out") return <LoginScreen onAuthed={refresh} />;

  return children({ user, logout: <LogoutButton onDone={refresh} /> });
}

/* ------------------------------------------------------------------- css --- */

const AUTH_CSS = `
.nt-auth { --ink:#141C26; --ink-2:#5A6875; --ink-3:#8C97A2; --paper:#EDF0F3;
  --card:#FFFFFF; --rule:#D8DEE4; --sig:#C11F44;
  background:var(--paper); color:var(--ink); min-height:100vh;
  display:flex; align-items:center; justify-content:center;
  font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,sans-serif;
  font-size:15px; line-height:1.5; padding:24px; }
.nt-auth *,.nt-auth *::before,.nt-auth *::after { box-sizing:border-box; }
.nt-auth .auth-wrap { width:100%; max-width:360px; background:var(--card);
  border:1px solid var(--rule); border-radius:12px; padding:32px 28px; }
.nt-auth .eyebrow { font-size:11px; letter-spacing:.16em; text-transform:uppercase;
  color:var(--ink-3); font-family:ui-monospace,SFMono-Regular,Menlo,monospace; }
.nt-auth h1 { font-size:24px; font-weight:600; letter-spacing:-.02em; margin:6px 0 8px; }
.nt-auth .muted { color:var(--ink-2); font-size:13px; margin:0 0 22px; }
.nt-auth .auth-form { display:flex; flex-direction:column; }
.nt-auth label { font-size:13px; font-weight:500; margin-bottom:6px; }
.nt-auth input { border:1px solid var(--rule); border-radius:7px; padding:10px 12px;
  font:inherit; font-size:15px; background:var(--paper); color:var(--ink);
  margin-bottom:16px; }
.nt-auth input:focus { outline:2px solid #3355C9; outline-offset:1px;
  border-color:transparent; }
.nt-auth input[aria-invalid="true"] { border-color:var(--sig); }
.nt-auth .auth-error { color:var(--sig); font-size:13px; margin:0 0 16px; }
.nt-auth .btn { background:var(--ink); color:var(--paper); border:none;
  border-radius:7px; padding:11px 18px; font:inherit; font-size:14px;
  font-weight:500; cursor:pointer; }
.nt-auth .btn:hover:not(:disabled) { background:#2A3542; }
.nt-auth .btn:disabled { opacity:.55; cursor:default; }
.nt-auth .btn:focus-visible { outline:2px solid #3355C9; outline-offset:2px; }
`;

export default LoginScreen;
