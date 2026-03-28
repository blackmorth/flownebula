# FlowNebula — Roadmap d'amélioration (issue-ready)

Ce document transforme l'audit en backlog concret, orienté livrables.

## Objectif

Amener FlowNebula vers un profiler "production-grade" en priorisant :
1. qualité des données,
2. fiabilité du pipeline,
3. scalabilité analytique,
4. expérience d'analyse avancée.

## Livrables que l'on peut ajouter immédiatement

### 1) Data quality score (MVP)

**But:** rendre visible la fiabilité réelle d'une session profilée.

**Changements proposés:**
- étendre les compteurs côté agent :
  - `dropped_by_probe`,
  - `dropped_by_agent`,
  - `rejected_by_server`.
- persister ces compteurs dans le payload session côté serveur.
- afficher un score synthétique dans l'UI session (ex: 0-100) + détail des pertes.

**Critères d'acceptation:**
- un utilisateur voit un score de qualité sur chaque session,
- les pertes sont attribuées explicitement à la source,
- une session avec pertes significatives est visuellement signalée.

### 2) Compatibilité protocole v1/v2

**But:** faire évoluer la sonde sans casser l'existant.

**Changements proposés:**
- introduire un header v2 avec `version`, `flags`, `seq`, `mono_ts`, `wall_ts`.
- accepter v1 et v2 côté serveur, puis convertir vers un modèle canonique interne.
- exposer la version réellement ingérée dans l'API de session.

**Critères d'acceptation:**
- uploads v1 fonctionnels sans régression,
- uploads v2 acceptés avec enrichissement temporel,
- tests d'intégration couvrant les deux versions.

### 3) Sampling adaptatif de bout en bout

**But:** éviter la saturation sans perdre toute visibilité.

**Changements proposés:**
- calcul d'un `effective_sample_rate` côté agent selon : taille de queue, erreurs upload, CPU agent.
- canal de contrôle runtime pour renvoyer ce taux à la probe.
- métriques Prometheus dédiées au contrôle de charge.

**Critères d'acceptation:**
- sous charge, le taux baisse automatiquement,
- après stabilisation, le taux remonte progressivement,
- baisse mesurable du taux de drop en stress test.

### 4) Mesures I/O réelles (PDO / cURL / FS)

**But:** sortir des `io_wait`/`network` à zéro.

**Changements proposés:**
- hooks ciblés sur appels externes critiques,
- ventilation des latences bloquantes : DB, HTTP, filesystem,
- exposition claire dans l'Overview + Timeline.

**Critères d'acceptation:**
- au moins trois catégories I/O peuplées en données réelles,
- corrélation visuelle entre latence endpoint et I/O wait,
- overhead mesuré et documenté.

## Plan 8 semaines (jalons)

### S1-S2
- Protocole v2 (header + parsing).
- Compatibilité serveur v1/v2.
- Compteurs qualité consolidés.

### S3-S4
- Sampling adaptatif.
- Canal de contrôle probe/agent.
- Score qualité affiché en UI.

### S5-S6
- Rollups analytiques (p50/p95/p99).
- Index requêtes `(service, endpoint, release, created_at)`.
- Requêtes comparatives baseline vs candidate.

### S7-S8
- Diff profil avancé (absolu/relatif par noeud).
- Heuristiques automatiques (N+1, fan-out, récursions).
- Corrélation OpenTelemetry (`trace_id`, `span_id`).

## KPIs de suivi

- **Qualité:** % sessions complètes, % events perdus, % sessions rejetées.
- **Overhead:** coût p95 de la probe (CPU et latence).
- **Scalabilité:** events/s, latence p95 d'upload, débit de persistance.
- **Valeur produit:** temps moyen vers root cause, régressions détectées avant prod.

## Découpage en tickets (proposition)

- EPIC-01: Data Quality & Reliability
- EPIC-02: Protocol Evolution v2
- EPIC-03: Adaptive Sampling Controller
- EPIC-04: I/O Runtime Instrumentation
- EPIC-05: Rollups & Comparative Analytics
- EPIC-06: Advanced Profiler Insights

Chaque epic peut être décliné en stories de 1 à 3 jours pour livraisons incrémentales.
