# domain-manager Development Guidelines

Auto-generated from all feature plans. Last updated: 2025-11-07

## Active Technologies
- Go 1.22+ + go-chi/chi v5, client-go, modernc.org/sqlite (002-multi-ingress)
- SQLite (WAL mode) — `system_settings` 表已有 `default_ingress_class` 和 `ingress_annotations` 欄位 (002-multi-ingress)
- Go 1.22+ + go-chi/chi v5, client-go, cert-manager (Helm dependency) (003-cloudflare-dns01)
- SQLite (system_settings) + K8s Secret (token) + K8s ClusterIssuer (003-cloudflare-dns01)
- Go 1.22+ + go-chi/chi v5, bcrypt (golang.org/x/crypto) (004-admin-settings)
- SQLite — admin_accounts 表已有 username, password_hash, email 欄位 (004-admin-settings)

- (001-k8s-domain-manager)

## Project Structure

```text
src/
tests/
```

## Commands

# Add commands for 

## Code Style

: Follow standard conventions

## Recent Changes
- 004-admin-settings: Added Go 1.22+ + go-chi/chi v5, bcrypt (golang.org/x/crypto)
- 003-cloudflare-dns01: Added Go 1.22+ + go-chi/chi v5, client-go, cert-manager (Helm dependency)
- 002-multi-ingress: Added Go 1.22+ + go-chi/chi v5, client-go, modernc.org/sqlite


<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->

<!-- Knowie: Project Knowledge -->
## Project Knowledge

This project maintains structured knowledge in `knowledge/`:

- **Principles** (`knowledge/principles.md`): Core axioms and derived development principles — the project's non-negotiable rules.
- **Vision** (`knowledge/vision.md`): Goals, current state, architecture decisions, and roadmap.
- **Experience** (`knowledge/experience.md`): Distilled lessons from past development — patterns, pitfalls, and takeaways.

Read these files at the start of any task to understand the project's *why* and constraints.
Additional context may be found in `knowledge/research/`, `knowledge/design/`, and `knowledge/history/`.
<!-- /Knowie -->
