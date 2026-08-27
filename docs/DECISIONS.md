# DECISIONS.md — Architecture Decision Log

> Diisi seiring pemilik project menjawab pertanyaan di `AUDIT_FINDINGS.md` §D dan seiring keputusan teknis di Phase 1 dikunci. Belum ada keputusan final — dokumen ini masih kosong menunggu Phase 0 gate.

## Format entri
```
### D-<nomor> — <judul singkat>
- Tanggal: YYYY-MM-DD
- Terkait: AUDIT_FINDINGS.md <ID temuan>
- Keputusan: ...
- Alasan: ...
- Dampak: ...
```

### D-1 — Geofencing ditegakkan, radius 100m
- Tanggal: 2026-08-26
- Terkait: AUDIT_FINDINGS.md A1
- Keputusan: Radius check-in ditegakkan di server, default **100 meter** dari lokasi kantor yang dikonfigurasi di `konfigurasi_lokasi` (bukan hardcoded lagi).
- Alasan: Menutup celah pemalsuan lokasi absensi.
- Dampak: Perlu desain penanganan kasus dinas luar/WFH yang sah — **belum diputuskan mekanismenya**, dibahas lagi di Phase 1/2 (opsi: flag "dinas luar" dengan approval, atau radius berbeda untuk WFH).

### D-2 — Route maintenance tanpa auth tidak diporting
- Tanggal: 2026-08-26
- Terkait: AUDIT_FINDINGS.md A2
- Keputusan: `/createSymlink`, `/foo1`, `/create-symlink`, `/create-symlinkk` tidak dibuat ulang di sistem baru. Operasi symlink/storage-link dilakukan via deploy script/CI (DevOps, Phase 5).

### D-3 — Bug reset password "12345" diperbaiki
- Tanggal: 2026-08-26
- Terkait: AUDIT_FINDINGS.md A3
- Keputusan: Update profil/data karyawan-user di sistem baru tidak pernah menyentuh password kecuali field password diisi eksplisit oleh admin.
- Dampak migrasi data: strategi reset password massal saat go-live belum diputuskan secara rinci — perlu dibahas saat Phase 1 (siapa yang diberitahu, bagaimana cara reset aman).

### D-4 — Rate limiting login ditambahkan
- Tanggal: 2026-08-26
- Terkait: AUDIT_FINDINGS.md A4
- Keputusan: Login (karyawan & admin) di API Go diberi rate limit per-IP dan per-akun.

### D-5 — IDOR pada task diperbaiki
- Tanggal: 2026-08-26
- Terkait: AUDIT_FINDINGS.md A5
- Keputusan: Endpoint edit/update task di sistem baru wajib scoping kepemilikan (`nik` pemilik = user login), kecuali akses admin dengan role yang berwenang.

### D-6 — Rotasi token Telegram bot
- Tanggal: 2026-08-26
- Terkait: AUDIT_FINDINGS.md A6, C2
- Keputusan: Token boleh dirotasi; pemegang akses BotFather adalah **programmer sebelumnya** (perlu koordinasi akses). **Untuk sementara, sistem baru tetap memakai bot/token/API Telegram yang lama** (belum dirotasi) sampai akses ke BotFather didapat — dicatat sebagai item follow-up, bukan blocker Phase 1.
- Tindak lanjut: cari cara mendapat akses BotFather dari programmer sebelumnya, atau buat bot baru dan migrasi chat.

### D-7 — Role-Based Access Control (RBAC) untuk admin
- Tanggal: 2026-08-26
- Terkait: AUDIT_FINDINGS.md A8
- Keputusan: Sistem baru menerapkan RBAC nyata untuk panel admin (bukan sekadar label `role_id` seperti sistem lama). Detail peran/permission spesifik dirancang di Phase 1 bersama pemilik project.

### D-8 — Unifikasi alur check-in
- Tanggal: 2026-08-26
- Terkait: AUDIT_FINDINGS.md B1
- Keputusan: Sistem baru punya **satu** alur check-in saja, menggantikan `PresensiController` (lama) dan `AbsensiController`/EOS (baru) yang berjalan paralel. Basis desain akan mengambil konsep shift-aware dari alur EOS (jadwal_kerja/jam_kerja) karena itu yang lebih maju, digabung dengan hal-hal yang masih relevan dari alur lama (WFH, dsb) — detail penggabungan dirancang di Phase 1.

### D-9 — Race condition check-in/out diperbaiki
- Tanggal: 2026-08-26
- Terkait: AUDIT_FINDINGS.md B2
- Keputusan: Sistem baru pakai unique constraint DB + transaction/row-lock untuk mencegah duplikasi/race condition pada check-in-out, menggantikan pola check-then-act lama.

### D-10 — Ambang keterlambatan: 09:15 default, bisa diatur admin, berbasis shift
- Tanggal: 2026-08-26
- Terkait: AUDIT_FINDINGS.md B3
- Keputusan: Ambang keterlambatan default **09:15**, tapi **dapat dikonfigurasi oleh admin/superadmin**, dan **berbeda per shift/jabatan** (karyawan dengan shift/jabatan tertentu punya jam masuk berbeda → ambang telat mengikuti jam masuk shift mereka masing-masing, bukan angka tunggal untuk semua orang).
- Dampak desain: `jam_kerja` perlu field ambang toleransi (mis. grace period dalam menit) yang bisa diedit admin per shift, bukan hardcoded di kode.

### D-11 — Foto check-in: nama file unik
- Tanggal: 2026-08-26
- Terkait: AUDIT_FINDINGS.md B8
- Keputusan: Sistem baru memberi nama file foto secara unik per event (timestamp/UUID), tidak lagi collision-prone.

### D-12 — Migrasi data: port data historis (bukan fresh start)
- Tanggal: 2026-08-26
- Terkait: AUDIT_FINDINGS.md C1
- Keputusan: Data historis **akan diporting** ke PostgreSQL baru (bukan fresh start). Akses dump database MySQL/MariaDB live akan disediakan oleh pemilik project.
- Tindak lanjut: jadwalkan pengambilan dump DB live; setelah didapat, rekonstruksi skema aktual (karena tidak ada file migration) menjadi bagian awal Phase 1.
- Belum diputuskan: penanganan baris data "cacat" (foto 0-byte, presensi duplikat dari race condition lama, jam_out overnight yang menggantung) — akan dibahas setelah dump DB tersedia dan bisa diperiksa langsung.

### D-13 — Overtime/pulang-cepat: hanya flag "pulang cepat", tanpa perhitungan lembur/payroll
- Tanggal: 2026-08-26
- Terkait: AUDIT_FINDINGS.md B4
- Keputusan: Sistem baru menambahkan flag **"pulang cepat"** (jam_out lebih awal dari `jam_pulang` shift karyawan) karena murah untuk dibangun begitu shift-aware lateness (D-10) ada. **Tidak** membangun perhitungan jam lembur/overtime dan payroll — itu proyek terpisah, ditunda ke luar scope migrasi ini kecuali diminta lagi nanti.

### D-14 — Logika shift overnight digeneralisasi (bukan hardcode kode shift tertentu)
- Tanggal: 2026-08-26
- Terkait: AUDIT_FINDINGS.md B5
- Keputusan: Deteksi "shift lintas tengah malam" diturunkan dari data shift itu sendiri (`jam_masuk > jam_pulang` pada `jam_kerja`), bukan hardcode nama kode shift (mis. `SH03`). Perilaku menutup baris kemarin (bukan insert baru) saat checkout tetap dipertahankan, berlaku untuk shift overnight manapun.

### D-15 — Approval workflow izin/sakit + exclude dari "tidak hadir"
- Tanggal: 2026-08-26
- Terkait: AUDIT_FINDINGS.md B6
- Keputusan: Ditambahkan alur approval minimal untuk `pengajuan_izin` (status: pending/approved/rejected, admin bisa approve/reject). Hari dengan izin/sakit yang **disetujui** dikecualikan dari status "tidak hadir" di rekap laporan kehadiran.

### D-16 — Form "buat sakit" digabung ke form izin
- Tanggal: 2026-08-26
- Terkait: AUDIT_FINDINGS.md B7
- Keputusan: Form terpisah "buat sakit" dihapus; digabung jadi satu form pengajuan dengan pilihan tipe (izin/sakit), sesuai struktur tabel `pengajuan_izin.status` (`'i'`/`'s'`) yang sudah ada.

### D-17 — Job harian untuk flag shift overnight yang tidak checkout
- Tanggal: 2026-08-26
- Terkait: AUDIT_FINDINGS.md B9
- Keputusan: Ditambahkan scheduled job harian yang menandai (bukan otomatis mengisi) baris presensi dengan `jam_out` masih NULL jauh melewati akhir shift (mis. sampai akhir hari berikutnya) sebagai "tidak checkout", untuk ditinjau admin. Tidak melakukan auto-fabrikasi jam checkout.

### D-18 — Resolusi konflik jadwal shift: jadwal_kerja (per-tanggal) menang atas konfigurasi_jamkerja (mingguan)
- Tanggal: 2026-08-26
- Terkait: AUDIT_FINDINGS.md B10
- Keputusan: Saat menentukan shift karyawan pada tanggal tertentu, `jadwal_kerja` (penugasan spesifik per-tanggal) selalu diprioritaskan bila ada entri; `konfigurasi_jamkerja` (default mingguan) hanya dipakai sebagai fallback ketika tidak ada entri spesifik di `jadwal_kerja` untuk tanggal tersebut.

---

**Semua 18 item checklist Phase 0 (`AUDIT_FINDINGS.md` §D) telah diputuskan.** Item terbuka yang masih perlu tindak lanjut operasional (bukan blocker desain): D-6 (akses BotFather untuk rotasi token — pakai token lama dulu), D-12 (jadwal pengambilan dump database live).

---

## Phase 1 — Skema & Kontrak API (Draft Awal)

### D-19 — Skema PostgreSQL provisional ditulis sebagai golang-migrate migrations
- Tanggal: 2026-08-26
- Terkait: Phase 1, semua keputusan D-1 s/d D-18
- Keputusan: Skema awal ditulis di `services/api/migrations/000001..000009_*.up/down.sql`, memetakan tabel legacy (`karyawan`→`employees`, `presensi`→`attendances`, `departemen`→`departments`, `jam_kerja`→`shifts`, `jadwal_kerja`→`work_schedules`, `konfigurasi_jamkerja`→`weekly_shift_defaults`, `konfigurasi_lokasi`→`office_locations`, `pengajuan_izin`→`leave_requests`, `tasks`→`tasks`) sekaligus mengimplementasikan seluruh keputusan D-1 s/d D-18 langsung di level skema (mis. `office_locations.radius_meters` benar-benar dipakai, `shifts.late_grace_minutes` per-shift, `shifts.is_overnight` generated column, `attendances` UNIQUE(employee_id, work_date) untuk mencegah race condition, `leave_requests.status` enum pending/approved/rejected, `leave_requests.type` enum izin/sakit menggantikan form terpisah).
- **Status: PROVISIONAL.** Nama kolom/tipe data asli di database live belum diverifikasi (tidak ada file migration Laravel — lihat D-12). Skema ini akan direvisi setelah dump database live diterima dan diperiksa (nama kolom, tipe data MySQL→Postgres, index/FK yang mungkin sudah ada, data yang harus dipetakan).
- Perubahan penamaan dari legacy: semua nama tabel/kolom diganti ke bahasa Inggris snake_case sesuai konvensi §9 CLAUDE.md (identifier kode dalam bahasa Inggris) — ini murni penamaan, bukan perubahan perilaku, tapi perlu diketahui saat mapping data historis di Phase 2.

### D-20 — Draf `openapi.yaml` awal
- Tanggal: 2026-08-26
- Terkait: Phase 1
- Keputusan: Kontrak API awal ditulis di `docs/openapi.yaml`, mencakup auth (karyawan+admin, JWT), attendance check-in/out **tunggal/unified** (D-8), leave requests dengan approval (D-15/D-16), tasks dengan ownership scoping (D-5), dashboard, reports/monitoring (RBAC-gated, D-7), dan config (office locations, shifts, jadwal). Endpoint maintenance/symlink legacy (A2/D-2) sengaja tidak diporting.
- **Status: DRAFT.** Response body detail untuk beberapa endpoint laporan (`/admin/reports/*`) belum dirinci penuh — akan dilengkapi begitu format laporan final disepakati bersama Frontend/QA di awal Phase 2/3.
- Belum diputuskan: mekanisme override radius untuk kasus dinas luar yang sah (terkait D-1) — endpoint check-in saat ini hanya membedakan `is_wfh`, belum ada jalur "dinas luar dengan approval". Perlu dibahas sebelum implementasi Phase 2 dimulai untuk endpoint ini.

### D-21 — Mekanisme "dinas luar": pre-approved assignment
- Tanggal: 2026-08-26
- Terkait: AUDIT_FINDINGS.md A1, DECISIONS.md D-1
- Keputusan: Dinas luar ditangani lewat **penugasan yang disetujui admin di muka** (mirip leave request) — admin menandai employee+tanggal tertentu sebagai "dinas luar" sebelum hari-H. Check-in pada tanggal yang sudah ditandai tsb **melewati pengecekan radius 100m** untuk hari itu. Bukan self-declare saat check-in, dan bukan tanpa mekanisme sama sekali.
- Dampak skema: perlu tabel/kolom baru untuk penugasan dinas luar (opsi: field `location_override` pada `attendances`/`work_schedules`, atau tabel terpisah `field_assignments` dengan `employee_id`, `date`, `approved_by`, `note`). Detail struktur ditentukan saat implementasi Phase 2; sementara dicatat sebagai kebutuhan skema tambahan di luar migration 000001-000009.
- Dampak kontrak API: perlu endpoint admin baru (mis. `POST /admin/config/field-assignments`) dan `attendance/check-in` perlu mengecek keberadaan penugasan dinas luar yang disetujui untuk employee+tanggal berjalan sebelum menegakkan radius.

### D-23 — Multi-cycle check-in/out per hari didukung (menggantikan model satu-baris-per-hari)
- Tanggal: 2026-08-26
- Terkait: LOGIC_SPEC.md §3/§5 (perilaku khusus departemen `GA`), ditemukan sebagai celah saat implementasi Phase 2
- Keputusan: Sistem baru **mendukung beberapa siklus check-in/out per karyawan per hari** (mis. rotasi shift satpam departemen `GA`), bukan membatasi satu pasang in/out per hari seperti asumsi awal skema `attendances`.
- Dampak skema: migration `000011_add_attendance_cycles` menambah kolom `attendances.cycle_number`, mengganti constraint unique dari `(employee_id, work_date)` menjadi `(employee_id, work_date, cycle_number)`.
- Dampak logika: siklus baru hanya bisa dimulai setelah cooldown minimum **5 menit** sejak checkout siklus sebelumnya (nilai rekayasa sementara menggantikan inkonsistensi legacy 30 menit vs 5 menit, D-8/D-9 — bisa dibuat configurable nanti bila diperlukan). Hanya siklus pertama hari itu yang dievaluasi terhadap ambang keterlambatan shift (D-10); siklus berikutnya tidak dianggap "terlambat".
- Catatan: keputusan ini melengkapi (bukan menggantikan) D-8 (unifikasi alur) dan D-9 (perbaikan race condition) — unifikasi tetap satu alur/endpoint check-in, hanya modelnya sekarang mendukung banyak siklus per hari.

### D-22 — Tech stack §2 CLAUDE.md dikonfirmasi final
- Tanggal: 2026-08-26
- Terkait: CLAUDE.md §2
- Keputusan: Seluruh default di §2 dikonfirmasi final tanpa perubahan — Go 1.22+/chi/GORM/golang-migrate/JWT untuk backend; Next.js App Router+TS+Tailwind+shadcn/ui+TanStack Query/Table untuk web; pnpm workspaces; Docker multi-stage+Compose; GitHub Actions untuk CI.

### D-24 — Revocation refresh-token saat logout (menutup simplifikasi Phase 2)
- Tanggal: 2026-08-26 (sesi re-orientasi lanjutan Phase 3)
- Terkait: catatan simplifikasi di `internal/usecase/auth/auth.go` sejak Phase 2 ("refresh token stateless tanpa revocation store — logout saat ini tidak benar-benar invalidasi token sebelum expired")
- Keputusan (implementasi teknis, bukan perubahan aturan bisnis baru — melengkapi celah yang sudah ditandai eksplisit sejak awal, bukan bug diam-diam yang baru ditemukan): refresh token JWT sekarang membawa klaim `jti` (random 16-byte hex via `crypto/rand`). `/auth/logout` men-decode token yang dikirim, lalu mencatat `jti`-nya ke tabel denylist `revoked_refresh_tokens` (migration `000012`). `/auth/refresh` menolak (`401`) token yang `jti`-nya ada di denylist. Access token (berumur pendek, default 15 menit) **tidak** dicek terhadap denylist — hanya refresh token yang direvoke; access token yang sudah terlanjur diterbitkan tetap valid sampai expiry alaminya setelah logout, ini standar dan disengaja (bukan celah).
- Dampak skema: migration baru `000012_create_revoked_refresh_tokens` (kolom `jti` unique, `expires_at` untuk kandidat cleanup job nanti — belum dijadwalkan, sama seperti D-17).
- Diverifikasi: dua sesi login independen (dua `jti` berbeda) — logout salah satu sesi tidak memengaruhi sesi lain (tidak over-broad); logout dua kali dengan token yang sama tetap `204` (idempotent via `ON CONFLICT DO NOTHING`); logout tanpa body/token tetap `204` (tidak pernah gagal untuk klien yang sesinya sudah hilang).

### ⚠️ DRIFT belum diputuskan — D-22 menyebut "pnpm workspaces", implementasi nyata pakai npm workspaces
- Tanggal ditemukan: 2026-08-26 (sesi re-orientasi, saat mengerjakan `packages/contracts`)
- Terkait: D-22 ("Seluruh default di §2 dikonfirmasi final tanpa perubahan ... pnpm workspaces ...")
- **Bukan keputusan baru — ini catatan drift yang butuh persetujuan pemilik project**, bukan perubahan yang diam-diam diterapkan. `apps/web` sudah berjalan dengan npm (`package-lock.json`, `RUN npm install` di `Dockerfile.dev`) sejak Phase 5, dan saat root `package.json` + workspace dibuat untuk `packages/contracts` di sesi ini, dilanjutkan dengan npm workspaces (`"workspaces": [...]` di `package.json`) supaya konsisten dengan apa yang sudah berjalan, **bukan** pnpm seperti tercatat di D-22.
- **❓ Perlu keputusan:** (a) revisi D-22 agar mencatat npm workspaces sebagai pilihan final (tidak ada dampak fungsional apapun untuk monorepo web-only saat ini), atau (b) migrasi seluruh `apps/web` + root ke pnpm sebelum Phase 6 (mobile) benar-benar dimulai, karena React Native/Expo tooling punya asumsi package-manager sendiri yang mungkin lebih cocok dengan salah satu pilihan. Tidak mendesak untuk Phase 3, tapi harus diputuskan sebelum Phase 6 dibuka.

### D-25 — Manajemen Hari Libur: tiga sumber digabung satu resolver, tanpa generate baris data
- Tanggal: 2026-08-27
- Terkait: fitur baru (tidak ada di sistem lama -- LOGIC_SPEC.md §6 mengonfirmasi legacy tidak punya tabel kalender libur sama sekali; kode shift `LBR` murni kosmetik, tidak pernah dipakai untuk exclude hari manapun)
- Keputusan: Status "hari libur" dihitung ON-DEMAND oleh satu resolver (`internal/usecase/holiday.Service.ResolveDayStatus`/`ResolveRange`), bukan digenerate sebagai baris ke tabel absensi/kehadiran -- kebijakan (weekend diubah, cuti bersama dadakan) langsung berlaku ke seluruh data lama tanpa perlu dibersihkan.
- Tiga sumber, precedence saat bertumpuk pada tanggal yang sama: **national > company > weekend** (nasional paling otoritatif, company dipakai terutama untuk MENAMBAH cuti bersama yang belum ada di sync nasional atau libur khusus perusahaan, weekend adalah fallback paling tidak spesifik):
  1. **Weekend** -- dikonfigurasi per perusahaan (`company_settings.working_weekdays`, array ISO weekday 1=Senin..7=Minggu, default `{1,2,3,4,5}`), TIDAK di-hardcode Sabtu-Minggu, mendukung skema 6-hari-kerja.
  2. **Nasional** -- tabel cache lokal `national_holidays` (unique per tanggal), disinkronkan on-demand (endpoint admin, tidak pernah dipanggil di request path absensi/laporan). Baris `source='manual'` (sudah diedit admin) tidak pernah ditimpa sync berikutnya.
  3. **Manual perusahaan** -- tabel `company_holidays`, tanggal tunggal atau rentang (`start_date`/`end_date`), tipe `libur`/`cuti_bersama`, CRUD admin (superadmin-only untuk mutasi, D-7).
- **Sumber sinkronisasi nasional: Google Calendar ICS publik** (`id.indonesian#holiday@group.v.calendar.google.com`), dipilih setelah tes langsung terhadap 3 kandidat: `api-harilibur.vercel.app` dan `dayoffapi.vercel.app` sama-sama mengembalikan `402 Payment Required` (sudah berbayar/tidak bisa diandalkan lagi), `date.nager.at` reliable tapi TIDAK membedakan cuti bersama sama sekali. Google Calendar ICS satu-satunya yang gratis, tanpa auth, dan secara eksplisit melabeli event `"Cuti Bersama ..."` di `SUMMARY` -- dikonfirmasi oleh pemilik project sebagai pilihan (bukan sumber resmi pemerintah, tapi tidak ada alternatif API resmi; trade-off ini didiskusikan eksplisit sebelum diimplementasikan).
- Integrasi resolver: (a) **recap** (`internal/usecase/recap`) -- hari libur dari salah satu dari tiga sumber ditandai "libur" dan DIKECUALIKAN dari Alpha, dihitung SEKALI per bulan (bukan per karyawan, karena ketiga sumber company-wide bukan per-karyawan) via `ResolveRange`; kehadiran nyata pada hari libur (lembur/kerja sukarela) tetap tampil "Hadir" dengan flag `is_holiday` terpisah, bukan disembunyikan. (b) **check-in** (`internal/usecase/attendance`) -- check-in pada hari libur tetap DIIZINKAN (default), tapi TIDAK PERNAH dihitung telat, dan baris attendance ditandai `is_holiday=true`. (c) **overnight (D-14)** -- resolver selalu dipanggil dengan `work_date` (tanggal mulai shift/civil date saat check-in), bukan `time.Now()`, sehingga shift yang mulai Jumat malam tidak salah dihitung libur karena lewat ke Sabtu.
- Dampak skema: migration `000017_add_working_weekdays_and_holidays` -- kolom baru `company_settings.working_weekdays` (smallint[]), tabel baru `national_holidays` dan `company_holidays`, kolom baru `attendances.is_holiday`.
- Diverifikasi: 10 unit test resolver (weekend default, weekend configurable 6-hari-kerja, nasional, manual single+range, precedence, ResolveRange konsisten dengan ResolveDayStatus per-tanggal, overnight edge Jumat-vs-Sabtu independen, sync idempotent, sync preserve manual override, sync gagal tidak merusak cache) + 1 test integrasi check-in (holiday diizinkan+tidak telat+tertandai) + 2 test parser ICS, semua PASS. Diverifikasi hidup di browser: sinkronisasi nyata menarik 28 hari libur nasional 2026 dari Google Calendar (termasuk cuti bersama berlabel), CRUD libur manual perusahaan (tambah+hapus) end-to-end, dan recap bulan Agustus 2026 menunjukkan weekend DAN 17 Agustus (libur nasional) sama-sama otomatis "L" -- sebelumnya kedua kasus itu jatuh ke Alpha karena tidak ada resolver sama sekali.
