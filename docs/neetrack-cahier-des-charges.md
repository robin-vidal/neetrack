# Neetrack — cahier des charges

Outil personnel de suivi Neetcode 150 avec révision espacée, multi-utilisateur, self-hosted.
Monorepo Go + React, déployé par Docker Compose.

Ce document est prescriptif : les décisions listées ont déjà été arbitrées. Ne les rediscute pas
en cours d'implémentation, signale-les si tu penses qu'elles sont fausses.

---

## 1. Décisions de conception à respecter

Ces choix ne sont pas arbitraires, ils sont le résultat de l'analyse préalable. Les réimplémenter
« plus simplement » casse le produit.

**La carte de révision porte sur le *pattern*, pas sur le problème.** Il y a ~18 patterns
(Arrays & Hashing, Two Pointers, Sliding Window…) et 150 problèmes. L'état FSRS
(stabilité, difficulté, échéance) est stocké par `(user, pattern)`. Raison : re-servir le même
problème mesure la mémoire de la solution, pas la capacité à résoudre ; et résoudre un problème
de sliding window renforce tous les autres, ce que FSRS ne modélise pas s'il travaille par problème.

**Quand un pattern est dû, on sert un problème jamais vu de ce pattern**, à défaut le moins
récemment tenté. Jamais le même problème par défaut.

**Acquisition groupée, révision entrelacée.** Un pattern jamais amorcé est servi en bloc
(N problèmes d'affilée dans la même session, N configurable, défaut 4) : le schéma ne se forme pas
sur un exemple unique. Un pattern déjà amorcé est plafonné à 1 nouveau problème par session.
Une session n'ouvre pas plus de `open_max` patterns neufs (défaut 1).

**Les tentatives sont un journal append-only ; l'état FSRS est dérivé.** On n'écrase jamais une
tentative. La carte est recalculable en rejouant le journal — indispensable pour changer de
scheduler ou de paramètres plus tard.

**La note est dérivée de la durée et des indices, pas auto-déclarée.** L'auto-évaluation est le
maillon faible. Le client envoie `duration_seconds`, `used_hint`, `saw_solution` ; le serveur
calcule le rating 1–4. Le client peut proposer une note, le serveur reste seul juge.

**Tout le scheduling est côté serveur.** Le frontend n'implémente aucune formule FSRS. Il affiche
ce que l'API renvoie.

**La rétention cible peut monter vers une date d'entretien.** Réglage `target_retention` +
`target_date` optionnelle : si la date est fixée, la rétention effective interpole linéairement
entre la valeur courante et 0.95 à l'approche de la date.

---

## 2. Stack

| Couche | Choix | Notes |
|---|---|---|
| Langage backend | Go 1.25+ | stdlib `net/http`, **aucun framework web** |
| Routeur | `http.ServeMux` | patterns méthode + wildcard (`"POST /api/login"`) |
| DB | PostgreSQL 17 | |
| Driver | `pgx/v5` + `pgxpool` | pas de `database/sql` |
| Requêtes | `sqlc` | SQL écrit à la main, types générés |
| Migrations | `goose`, embarquées via `embed.FS` | appliquées au démarrage |
| Scheduler | `github.com/open-spaced-repetition/go-fsrs/v3` | ne pas réimplémenter les formules |
| Logs | `log/slog`, JSON en prod | |
| Frontend | React 19 + TypeScript + Vite | |
| État serveur | TanStack Query | pas de Redux, pas de Zustand |
| Routing | react-router v7 | 3 routes suffisent |
| Style | CSS modules ou CSS plain | reprendre le design fourni tel quel |
| Runtime | Docker Compose | |

Pas de ORM. Pas de générateur de scaffolding. Pas de gRPC.

---

## 3. Arborescence

```
neetrack/
├── cmd/
│   └── neetrack/
│       └── main.go              // seul endroit qui câble les dépendances
├── internal/
│   ├── config/                  // parsing env, une struct, validée au boot
│   ├── catalog/                 // les 150 problèmes, JSON embarqué + accesseurs
│   ├── auth/                    // argon2id, sessions, middleware
│   ├── scheduler/               // wrapper FSRS, composition de session, notation
│   ├── store/                   // pgx, sqlc généré, migrations
│   └── httpapi/                 // handlers, middleware, routeur, DTO
├── migrations/                  // 0001_init.sql, …
├── web/
│   ├── src/
│   │   ├── api/                 // client fetch typé + hooks TanStack Query
│   │   ├── components/
│   │   ├── routes/
│   │   └── types.ts             // miroir des DTO Go
│   ├── index.html
│   ├── vite.config.ts
│   └── package.json
├── Dockerfile                   // multi-stage: node → go → runtime
├── compose.yaml                 // prod
├── compose.dev.yaml             // dev, avec hot reload
├── Makefile
├── sqlc.yaml
└── go.mod
```

Règles de structure :

- `internal/` uniquement. Rien dans `pkg/` — ce code n'est pas une bibliothèque.
- Aucun état global, aucun `init()`, aucun singleton. Tout est construit dans `main.go` et passé
  explicitement.
- Les handlers sont des méthodes sur un `Server` qui porte ses dépendances par interface.
- `internal/scheduler` ne connaît ni HTTP ni SQL. Il prend des structs de domaine et rend des
  structs de domaine. C'est la partie testable en priorité.

---

## 4. Modèle de données

```sql
CREATE TABLE users (
    id            BIGSERIAL PRIMARY KEY,
    name          TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE sessions (
    id         BYTEA PRIMARY KEY,              -- 32 octets aléatoires, hashés SHA-256
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX ON sessions (expires_at);

-- Journal append-only. Aucun UPDATE, aucun DELETE.
CREATE TABLE attempts (
    id               BIGSERIAL PRIMARY KEY,
    user_id          BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    problem_slug     TEXT NOT NULL,            -- clé du catalogue embarqué
    pattern          TEXT NOT NULL,            -- dénormalisé volontairement (le catalogue peut bouger)
    rating           SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 4),
    duration_seconds INTEGER,
    used_hint        BOOLEAN NOT NULL DEFAULT false,
    saw_solution     BOOLEAN NOT NULL DEFAULT false,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX ON attempts (user_id, created_at DESC);
CREATE INDEX ON attempts (user_id, problem_slug);

-- État dérivé, reconstructible en rejouant attempts.
CREATE TABLE pattern_cards (
    user_id        BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    pattern        TEXT NOT NULL,
    stability      DOUBLE PRECISION NOT NULL,
    difficulty     DOUBLE PRECISION NOT NULL,
    due_at         TIMESTAMPTZ NOT NULL,
    last_review_at TIMESTAMPTZ NOT NULL,
    reps           INTEGER NOT NULL DEFAULT 0,
    lapses         INTEGER NOT NULL DEFAULT 0,
    state          SMALLINT NOT NULL,
    PRIMARY KEY (user_id, pattern)
);
CREATE INDEX ON pattern_cards (user_id, due_at);

CREATE TABLE user_settings (
    user_id          BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    target_retention DOUBLE PRECISION NOT NULL DEFAULT 0.90
                     CHECK (target_retention BETWEEN 0.70 AND 0.99),
    target_date      DATE,
    block_size       SMALLINT NOT NULL DEFAULT 4 CHECK (block_size BETWEEN 1 AND 6),
    open_max         SMALLINT NOT NULL DEFAULT 1 CHECK (open_max BETWEEN 1 AND 3),
    new_pct          SMALLINT NOT NULL DEFAULT 40 CHECK (new_pct BETWEEN 0 AND 100),
    session_size     SMALLINT NOT NULL DEFAULT 5 CHECK (session_size BETWEEN 1 AND 20)
);

-- Neetcode a renommé la plupart des problèmes sur son site : les slugs ne se dérivent
-- pas des titres. L'utilisateur corrige un lien une fois, la correction persiste.
CREATE TABLE problem_links (
    user_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    problem_slug TEXT NOT NULL,
    neetcode_slug TEXT NOT NULL,
    PRIMARY KEY (user_id, problem_slug)
);
```

Le catalogue des 150 problèmes (`slug`, `title`, `pattern`, `difficulty`, `order`) est un JSON
embarqué dans le binaire via `embed.FS`, **pas une table**. Il ne change qu'avec une release.
Le fichier JSX fourni contient la liste complète et l'ordre du roadmap : reprends-la telle quelle.

---

## 5. Domaine — `internal/scheduler`

### 5.1 Notation dérivée

```go
type Outcome struct {
    Difficulty      catalog.Difficulty // Easy | Medium | Hard
    Duration        time.Duration
    UsedHint        bool
    SawSolution     bool
}

func Rate(o Outcome) fsrs.Rating
```

Règles, dans cet ordre :

1. `SawSolution` → `Again` (1)
2. `UsedHint` → `Hard` (2)
3. `Duration > 2 × cible` → `Hard` (2)
4. `Duration > cible` → `Good` (3)
5. `Duration ≤ cible / 2` → `Easy` (4)
6. sinon → `Good` (3)

Cibles : Easy 12 min, Medium 25 min, Hard 40 min. Constantes exportées, testables.
Si `Duration` est absente (chrono non lancé), accepter le rating envoyé par le client en repli.

### 5.2 Rétention effective

```go
func EffectiveRetention(base float64, targetDate *time.Time, now time.Time) float64
```

Sans `targetDate` : renvoie `base`. Avec : interpolation linéaire de `base` vers 0.95 sur les
90 derniers jours avant la date, bornée à 0.95 après.

### 5.3 Composition de session

```go
type Request struct {
    Size       int
    NewPct     int
    BlockSize  int
    OpenMax    int
}

type Item struct {
    Problem  catalog.Problem
    Reason   Reason  // ReasonReview | ReasonNew
    Retrievability float64 // rempli si ReasonReview
}

func Compose(cards []Card, attempts AttemptIndex, req Request, now time.Time) []Item
```

Algorithme :

1. Patterns dus = `due_at <= now`, triés par rétrievabilité croissante (le plus dégradé d'abord).
2. Quota révision = `round(Size × (1 - NewPct/100))`, nul s'il n'y a aucun pattern dû.
3. Pour chaque pattern dû, dans l'ordre, jusqu'au quota : prendre un problème jamais tenté de ce
   pattern ; à défaut le moins récemment tenté.
4. Remplir le reste avec des problèmes jamais tentés, dans l'ordre du roadmap, sous contrainte :
   au plus `BlockSize` par pattern non amorcé, au plus 1 par pattern amorcé, au plus `OpenMax`
   patterns neufs ouverts.
5. S'il reste des places, compléter par des révisions supplémentaires.
6. Trier : révisions d'abord, nouveaux ensuite.

`Compose` est une fonction pure. Elle doit être couverte par des tests table-driven :
utilisateur vierge, tout dû, rien dû, `NewPct` à 0 et à 100, catalogue épuisé.

### 5.4 Rejeu

```go
func Replay(attempts []Attempt, settings Settings) map[string]Card
```

Reconstruit toutes les cartes depuis le journal. Exposé par une commande CLI
(`neetrack replay --user=…`), pas par l'API. Sert à changer de paramètres FSRS sans perdre
l'historique.

---

## 6. API HTTP

Toutes les routes sont sous `/api`, JSON, et exigent une session valide sauf `login`.
Erreurs : `{"error":"code_machine","message":"phrase lisible"}` avec le bon statut HTTP.

| Méthode | Route | Corps / retour |
|---|---|---|
| `POST` | `/api/login` | `{name, password}` → 204 + cookie ; 401 sinon |
| `POST` | `/api/logout` | → 204, session supprimée en base |
| `GET` | `/api/me` | → `{name}` ou 401 |
| `GET` | `/api/catalog` | → patterns + problèmes + progression + rétrievabilité courante |
| `GET` | `/api/settings` | → réglages |
| `PATCH` | `/api/settings` | patch partiel des réglages |
| `POST` | `/api/sessions` | `{size, new_pct}` → liste d'items composée |
| `POST` | `/api/attempts` | `{problem_slug, duration_seconds, used_hint, saw_solution}` → `{rating, pattern, next_due_at, interval_days}` |
| `GET` | `/api/activity?weeks=26` | → `[{date, count}]` + `{streak}` |
| `PUT` | `/api/links/{slug}` | `{neetcode_slug}` → 204 |

Notes :

- `POST /api/sessions` ne persiste rien. La session de travail vit dans le state React ; seules
  les tentatives sont écrites. Une session abandonnée ne laisse aucune trace, ce qui est voulu.
- `POST /api/attempts` fait tout le travail : dériver la note, charger la carte, appliquer FSRS,
  écrire la tentative et la carte **dans une même transaction**.
- `GET /api/catalog` renvoie la rétrievabilité calculée au moment de la requête, pas stockée.

---

## 7. Authentification

- Hash `argon2id` via `golang.org/x/crypto/argon2`. Paramètres : `time=1`, `memory=64*1024`,
  `threads=4`, `keyLen=32`, sel aléatoire de 16 octets. Encodage PHC standard en base.
- Comparaison en temps constant (`subtle.ConstantTimeCompare`).
- Session : 32 octets `crypto/rand`, envoyés en clair dans le cookie, **stockés hashés SHA-256**
  en base. Durée 30 jours, prolongée à l'usage.
- Cookie : `HttpOnly`, `Secure` (désactivable en dev par config), `SameSite=Lax`, `Path=/`.
- Rotation de l'identifiant de session à la connexion.
- Rate limit sur `/api/login` : 5 tentatives / minute / IP, réponse 429. Un simple token bucket
  en mémoire suffit, pas de Redis.
- Message d'erreur identique pour identifiant inconnu et mot de passe faux (pas d'énumération).
- Pas d'inscription publique. Création de compte par CLI : `neetrack adduser <nom>`, mot de passe
  lu sur stdin sans écho. C'est un outil pour deux personnes.
- Middleware `RequireAuth` qui injecte l'utilisateur dans le `context.Context` via une clé typée
  privée.

---

## 8. Frontend

Le design est fourni : deux fichiers JSX (`neetrack.jsx` pour le tracker, `neetrack-auth.jsx`
pour le login). **Reprends la structure visuelle, le CSS et les libellés français tels quels.**
Ton travail est de les porter, pas de les redessiner.

Portage à faire :

- JSX → TSX, avec des types miroir des DTO Go dans `types.ts`.
- Supprimer toute la logique FSRS du client (`W`, `review`, `retrievability`, `composeSession`,
  `mastery`, `activity`, `streakOf`) — elle passe côté serveur.
- Remplacer `window.storage` par le client API + TanStack Query.
- Remplacer les `prompt()` par de vraies modales accessibles (correction de lien, pseudo).
- Garder le chrono en state local ; n'envoyer la durée qu'au moment de la notation.
- Garder les trois états du `AuthGate` : connecté, déconnecté, serveur injoignable.

Routes : `/` (patterns), `/session`, `/settings`. Le login est un écran plein, hors routeur.

Le pseudo affiché en haut vient de `GET /api/me`, plus d'un réglage local.

Accessibilité : focus visible partout, `aria-live` sur les erreurs, chrono lisible par lecteur
d'écran, `prefers-reduced-motion` respecté. Responsive jusqu'à 380 px — le heatmap défile
horizontalement, le reste s'empile.

---

## 9. Docker et workflow local

**`Dockerfile`** multi-stage :

1. `node:22-alpine` → `npm ci && npm run build` dans `web/`
2. `golang:1.25-alpine` → copie `web/dist` dans `internal/httpapi/assets/`, `go build -trimpath
   -ldflags="-s -w"`
3. `gcr.io/distroless/static-debian12` → binaire seul, utilisateur non-root

En production le binaire sert le frontend via `embed.FS` et un handler SPA (fallback `index.html`
sur les routes inconnues, sauf `/api/*`). Un seul conteneur, un seul port.

**`compose.yaml`** : `db` (postgres:17-alpine, volume nommé, healthcheck `pg_isready`) et `app`
(dépend de `db` en `service_healthy`, port 8080, variables d'env).

**`compose.dev.yaml`** : `db`, `api` (avec `air` pour le hot reload, code monté en volume),
`web` (`vite dev` sur 5173, proxy `/api` → `api:8080`). C'est le mode de travail par défaut.

**`Makefile`** : `dev`, `build`, `test`, `lint`, `migrate`, `sqlc`, `adduser`, `seed`.
Un `make dev` doit suffire à partir d'un clone vierge.

Config par variables d'environnement uniquement : `DATABASE_URL`, `LISTEN_ADDR`, `COOKIE_SECURE`,
`LOG_LEVEL`. Aucun fichier de config, aucun défaut caché dans le code — un env manquant fait
échouer le démarrage avec un message explicite.

---

## 10. CI/CD et déploiement

Cible : un VPS Contabo (Debian/Ubuntu, x86_64, mono-nœud). Deux utilisateurs, pas de contrainte
de haute disponibilité. Le déploiement doit être ennuyeux et réversible, pas sophistiqué.

### 10.1 Intégration continue

`.github/workflows/ci.yaml`, sur pull request et push sur `main` :

1. `actions/setup-go` avec cache de modules, `actions/setup-node` avec cache npm.
2. `make lint` (golangci-lint), `npm run typecheck` et `npm run lint` dans `web/`.
3. `make test` — les tests de `internal/store` utilisent testcontainers, disponible sur les
   runners `ubuntu-latest`.
4. `docker build` sans push, pour vérifier que le multi-stage passe.

Le job échoue si la couverture de `internal/scheduler` descend sous 85 %.

### 10.2 Image et registre

`.github/workflows/release.yaml`, sur push `main` :

- Build `linux/amd64` uniquement. Pas de multi-arch : le VPS est x86_64, le build ARM doublerait
  la durée pour rien.
- Push sur GHCR : `ghcr.io/<owner>/neetrack:<sha-court>` **et** `:latest`.
- `docker/build-push-action` avec cache GitHub Actions (`type=gha`).
- Le paquet GHCR est privé ; le VPS s'y authentifie avec un token à portée `read:packages`.

### 10.3 Déploiement

`.github/workflows/deploy.yaml`, déclenché après un `release` réussi (et manuellement via
`workflow_dispatch`, avec un tag en entrée pour le rollback).

Sur le VPS, dans `/opt/neetrack` :

```bash
echo "IMAGE_TAG=${SHA}" > .image
docker compose pull
docker compose up -d --remove-orphans
docker image prune -f
```

Points non négociables :

- **L'image est épinglée par tag SHA, pas par `latest`.** `compose.yaml` lit
  `image: ghcr.io/<owner>/neetrack:${IMAGE_TAG}`. Le rollback est alors une seule commande :
  remettre l'ancien SHA dans `.image` et relancer `up -d`.
- Le dépôt ne contient **jamais** de `.env`. Il vit en `/opt/neetrack/.env`, `chmod 600`,
  propriétaire root. Y figurent `DATABASE_URL`, `POSTGRES_PASSWORD`, `COOKIE_SECURE=true`,
  `IMAGE_TAG`.
- Secrets GitHub nécessaires : `SSH_HOST`, `SSH_USER`, `SSH_KEY` (clé dédiée au déploiement, sans
  passphrase, autorisée uniquement pour ce compte), `GHCR_TOKEN`.
- Les migrations tournent au démarrage du conteneur. Une migration ne doit donc jamais être
  destructive dans la même release que le code qui en dépend : ajouter une colonne dans une
  release, la remplir dans la suivante, supprimer l'ancienne dans une troisième. Sinon un rollback
  d'image tombe sur un schéma incompatible.
- Interruption de service attendue : quelques secondes, le temps que le conteneur redémarre.
  C'est accepté. Ne construis pas de blue-green ni de rolling update pour deux utilisateurs.

### 10.4 Reverse proxy et TLS

Un conteneur `caddy:alpine` dans le compose de production, qui termine le TLS et proxy vers
l'application. Caddyfile minimal :

```
neetrack.<domaine>.fr {
    reverse_proxy app:8080
}
```

Let's Encrypt est automatique. `COOKIE_SECURE=true` en production — sans TLS le cookie de session
ne serait pas protégé, et c'est le seul secret qui circule.

Le service `db` n'expose **aucun port** vers l'hôte. Postgres n'est joignable que sur le réseau
Docker interne.

### 10.5 Sauvegardes

Sans base sauvegardée, tout l'historique de révision disparaît au premier incident disque — et
c'est précisément l'historique qui rend l'état FSRS reconstructible.

- `pg_dump` quotidien via une tâche cron sur l'hôte ou un conteneur `ofelia`, vers
  `/opt/neetrack/backups/`, rétention 7 jours.
- Copie hors-site : `restic` ou `rclone` vers un stockage distinct du VPS. Un dump qui vit sur le
  disque qu'il est censé protéger ne protège rien.
- Un test de restauration documenté dans le README, à exécuter au moins une fois : restaurer le
  dump dans une base jetable et vérifier que l'application démarre dessus.

### 10.6 Durcissement du VPS

Contabo ne fournit ni pare-feu managé ni instantanés automatiques sur les offres de base : tout
est à ta charge.

- `ufw` : 22, 80, 443 en entrée, rien d'autre.
- SSH par clé uniquement, `PasswordAuthentication no`, `PermitRootLogin no`, `fail2ban` actif.
- Un utilisateur non-root membre du groupe `docker`, propriétaire de `/opt/neetrack`.
- `unattended-upgrades` pour les correctifs de sécurité système.
- Docker en version du dépôt officiel, pas celle de la distribution.

---

## 11. Conventions Go attendues

- Erreurs enveloppées avec `%w`, sentinelles exportées (`ErrNotFound`, `ErrInvalidCredentials`).
  Les handlers traduisent une erreur de domaine en statut HTTP, la couche store ne connaît pas HTTP.
- `context.Context` en premier paramètre, propagé jusqu'à pgx. Timeout par requête via middleware.
- Aucun `panic` en dehors de `main`. Middleware de recover qui log et rend 500.
- Arrêt propre : `signal.NotifyContext`, `server.Shutdown` avec 10 s de grâce, fermeture du pool.
- Interfaces définies côté consommateur, petites (1–3 méthodes). Pas d'interface « au cas où ».
- `golangci-lint` avec au minimum `errcheck`, `govet`, `staticcheck`, `revive`, `sqlclosecheck`.
- Tests : table-driven, `t.Parallel()` quand c'est sûr, `httptest` pour les handlers,
  `testcontainers-go` pour la couche store. Le domaine (`scheduler`) se teste sans base.

Couverture minimale attendue : `internal/scheduler` ≥ 85 %, `internal/auth` ≥ 80 %.
Le reste n'a pas de seuil.

---

## 12. Jalons

1. **Socle** — module Go, compose dev, migrations, `/api/health`, `make dev` fonctionnel.
2. **Auth** — argon2id, sessions, middleware, `adduser`, écran de login porté.
3. **Catalogue** — JSON embarqué, `GET /api/catalog`, écran patterns porté avec barres de rétention.
4. **Boucle de révision** — `POST /api/sessions`, `POST /api/attempts`, chrono, notation dérivée.
   C'est le cœur : rien ne sert d'aller plus loin avant que cette boucle soit testée.
5. **Activité** — heatmap, série, `GET /api/activity`.
6. **Réglages** — rétention, date cible, taille de bloc, `open_max`, correction de liens.
7. **Prod** — Dockerfile multi-stage, `embed.FS`, compose de production.
8. **Livraison** — CI sur PR, image GHCR, déploiement SSH, Caddy et TLS, sauvegardes. À faire
   dès que le jalon 4 est vert : déployer tôt une boucle qui marche vaut mieux que déployer tard
   un produit complet.

---

## 13. Hors périmètre

Inscription publique, réinitialisation de mot de passe, OAuth, e-mail, notifications,
classement entre utilisateurs, import LeetCode, mobile natif, i18n (le produit est en français),
et tout entraînement de paramètres FSRS personnalisés — `go-fsrs` n'embarque pas l'optimiseur, et
le volume de données d'un utilisateur ne le justifierait pas.

---

## 14. Critères d'acceptation

- `git clone && make dev` donne une application fonctionnelle sans étape manuelle.
- Un utilisateur créé par CLI peut se connecter, composer une session de 5, en noter 5, et voir
  les échéances de ses patterns bouger en conséquence.
- Deux utilisateurs simultanés ne voient jamais les données l'un de l'autre. Un test le vérifie.
- Supprimer `pattern_cards` et lancer `neetrack replay` reconstruit à l'identique l'état dérivé.
- `docker compose -f compose.yaml up` démarre un conteneur unique servant l'API et le frontend.
- Aucun secret, aucune URL de base, aucun mot de passe dans le dépôt.
- Un push sur `main` qui passe la CI aboutit à la nouvelle version en ligne sans intervention.
- Remettre le SHA précédent dans `/opt/neetrack/.env` et relancer `docker compose up -d` restaure
  la version antérieure en moins d'une minute.
- Un dump restauré dans une base vierge suffit à redémarrer l'application avec tout l'historique.
