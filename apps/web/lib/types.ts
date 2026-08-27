// Re-exports the shared contract types (packages/contracts) so existing
// "@/lib/types" imports across the app keep working unchanged. The actual
// definitions live in packages/contracts/src/index.ts -- extracted there in
// Phase 3 so a future mobile app (Phase 6) has a single source of truth to
// depend on instead of re-declaring these shapes.
export * from "@absensi-next/contracts";
