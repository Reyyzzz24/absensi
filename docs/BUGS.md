# BUGS.md — Bug Tracker (Phase 4)

> Per CLAUDE.md §6, a phase is "0 bug" only when this file has no open bug with severity ≥ minor. Every entry must say what broke, how it was found, and how it was verified fixed.

Severity: 🔴 Kritis · 🟠 Tinggi · 🟡 Sedang · ⚪ Rendah/kosmetik

## Open bugs

### 🟡 Parity matrix #7 — Telegram check-in/out notification never ported
- **Ditemukan:** 2026-08-26, Phase 4 parity-matrix sweep (grep untuk "telegram" di `services/api` — nol hasil selain satu komentar di `internal/config/config.go` tentang tidak menghardcode kredensial).
- **Bukan bug kode** — ini item yang tercatat *sudah diputuskan* (D-6: "pakai token/API Telegram lama dulu sampai akses BotFather didapat dari programmer sebelumnya"), tapi **implementasinya sendiri sepertinya tidak pernah ditulis** di backend Go manapun. Legacy mengirim notifikasi Telegram di setiap check-in/check-out; sistem baru sama sekali tidak punya kode yang memanggil Telegram API.
- **❓ Perlu keputusan pemilik project** (bukan sesuatu yang boleh diam-diam ditambahkan/dilewati): apakah notifikasi Telegram memang sengaja ditunda ke luar scope Phase 4 (mis. dianggap bagian dari fitur non-inti), atau ini terlewat dan harus ditambahkan sebelum Phase 4 bisa dianggap selesai untuk item ini. Severity 🟡 karena bukan celah keamanan/data — hanya hilangnya satu channel notifikasi operasional.

## Closed bugs

_None yet — no product bugs have been found and fixed during Phase 4 so far. Issues encountered while writing the test suite itself (wrong FK references, timezone comparison in a test assertion, cross-package test-database races) were test-authoring mistakes, not product defects, and are documented as code comments in the relevant `_test.go` files / `internal/testutil/db.go` instead of here._

## Testing done so far (2026-08-26)

Automated Go tests added under `services/api/internal/**/*_test.go`, run against a real Postgres (`internal/testutil.NewDB`, see that package's doc comment for the required `-p 1` flag). All 25 tests pass as of this writing (`go test -p 1 ./...`).

Coverage, mapped to `docs/PARITY_MATRIX.md` items:

- **#2** (bcrypt hash portability) — `usecase/auth`: login succeeds against a pre-existing bcrypt hash with no rehash step.
- **#3** (login rate limiting, A4/D-4) — `middleware`: blocks after N attempts, keyed per client, resets after the window.
- **#8** (geofencing actually enforced, A1/D-1) — `usecase/attendance`: check-in outside the radius is rejected; inside is accepted with recorded distance.
- **#9/#10** (overnight shift closes yesterday's open row, D-14) — `usecase/attendance`: a check-in the next calendar day closes yesterday's still-open overnight row instead of starting a new one.
- **#11** (idempotency/cooldown, D-9/D-23) — `usecase/attendance`: an immediate re-check-in after checkout is rejected until the cooldown passes.
- **#13** (early-leave flag, B4) — `usecase/attendance`: checkout before shift end is flagged `is_early_leave`.
- **#14** (izin/sakit approval + effect on recap, D-15/D-16) — `usecase/leave`: submissions default to pending, review transitions state once and only once; `usecase/recap`: an *approved* leave day is classified `izin`/`sakit` (never `alpha`), a still-*pending* one is not excused.
- **#15** (task IDOR, A5/D-5) — `usecase/task`: a non-owner's update attempt returns the same "not found" as a nonexistent task (no existence leak) and the task is provably untouched; `ListOwn` never returns another employee's tasks.
- **#22** (admin RBAC, D-7) — `middleware`: a plain `admin` token is rejected (403) on a superadmin-only route; a `superadmin` token is accepted.

Also covered, not tied to a specific parity-matrix row:
- D-1/D-21 bypass paths (WFH, approved field assignment) still record geofence data but don't block check-in.
- D-24 refresh-token revocation: logging out one session's refresh token doesn't affect a second independent session's token; logout is idempotent and tolerates garbage/missing tokens.
- D-3 regression guard: editing an employee's other fields never resets their password; an explicit new password does rehash correctly; deactivation/reactivation round-trips.

## Not yet covered (tracked, not silently skipped)

- Frontend E2E (Playwright) — none written yet.
- Full input-validation audit across every handler (parity matrix #5) — only spot-checked via the tests above.
- Photo storage collision/path-traversal (#6) — `storage.local.go`'s `ReadPhoto` path-traversal guard has no dedicated test yet.
- Scheduled job D-17 (flag no-checkout) — verified manually live in a previous session (see PROGRESS.md Phase 5), no automated test.
- Performance/N+1 sanity pass on reports/recap and monitoring endpoints — not yet done.
- HTTP-handler-level tests (the suite above tests usecases directly, not the chi routes/JSON (de)serialization layer).
