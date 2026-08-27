# AUDIT_FINDINGS.md — Temuan Celah, Bug, dan Keputusan yang Butuh Approval

> Setiap temuan: deskripsi, dampak, rekomendasi + trade-off, dan pertanyaan approval eksplisit.
> **Sesuai aturan CLAUDE.md §1.3: tidak ada perbaikan yang diterapkan diam-diam. Semua menunggu keputusan pemilik project.**
> Severity: 🔴 Kritis · 🟠 Tinggi · 🟡 Sedang · ⚪ Rendah/kosmetik

---

## A. Keamanan

### A1. 🔴 Geofencing tidak ditegakkan sama sekali
**Deskripsi:** Jarak ke kantor dihitung tapi hasilnya tidak pernah dibandingkan/ditolak. Karyawan bisa check-in dari lokasi manapun di dunia. Fitur "radius" di UI murni kosmetik.
**Dampak:** Absensi bisa dipalsukan sepenuhnya dari jarak jauh — merusak validitas seluruh data kehadiran.
**Rekomendasi:** Tegakkan radius di server menggunakan nilai dari `konfigurasi_lokasi` (yang sudah ada tapi tidak dipakai), bukan koordinat hardcoded. Tolak/beri flag submission di luar radius.
**Trade-off:** Karyawan lapangan/dinas luar yang sah akan ikut tertolak kecuali ada mekanisme override (mis. approval manual, atau flag "dinas luar").
**❓ Approval:** Apakah geofencing boleh **mulai ditegakkan** di sistem baru? Kalau ya, radius default berapa meter, dan bagaimana menangani kasus dinas luar/WFH yang sah?

### A2. 🔴 4 route maintenance tanpa autentikasi (`/createSymlink`, `/foo1`, `/create-symlink`, `/create-symlinkk`)
**Deskripsi:** Endpoint publik yang menjalankan `shell_exec`, `symlink()`, dan `Artisan::call()` tanpa auth sama sekali.
**Dampak:** Permukaan serangan terbuka di production; `/foo1` khususnya adalah pola shell_exec yang berbahaya kalau argumennya pernah diubah jadi dinamis.
**Rekomendasi:** Di sistem baru, operasi symlink/storage-link dilakukan lewat deploy script/CI, bukan route HTTP. Tidak perlu endpoint setara di Next.js/Go sama sekali.
**Trade-off:** Tidak ada — ini murni operational tooling yang salah tempat, tidak ada business value yang hilang.
**❓ Approval:** Setuju endpoint-endpoint ini **tidak diporting** ke sistem baru (digantikan proses deploy/CI)?

### A3. 🔴 Password admin/karyawan direset ke `"12345"` di setiap edit
**Deskripsi:** `KaryawanController::update()` dan `UserController::update()` selalu menimpa password dengan hash dari string hardcoded `"12345"`, walau admin hanya mengedit field lain (nama, telepon, dll).
**Dampak:** Semua akun yang pernah diedit admin kemungkinan sekarang punya password `"12345"` tanpa sepengetahuan pemiliknya — risiko keamanan besar pada sistem lama, dan perlu ditangani saat migrasi data.
**Rekomendasi:** Sistem baru: update profil TIDAK PERNAH menyentuh password kecuali field password diisi eksplisit (pola yang sudah benar di `updateprofile()` self-service). Untuk data lama: paksa reset password terverifikasi (bukan `"12345"` lagi) untuk akun yang diketahui pernah terkena bug ini.
**Trade-off:** Perlu komunikasi ke seluruh karyawan/admin bahwa password mereka akan direset saat migrasi (mengganggu tapi perlu).
**❓ Approval:** (1) Setuju perilaku ini **diperbaiki** di sistem baru? (2) Untuk migrasi data, apakah semua user dipaksa reset password saat go-live, atau ada strategi lain untuk tahu siapa yang terkena bug ini?

### A4. 🟠 Tidak ada rate limiting pada login (brute force terbuka)
**Rekomendasi:** Tambahkan rate limiting (mis. per IP + per akun) pada endpoint login di API Go.
**Trade-off:** Minimal — praktik standar, tidak ada downside fungsional.
**❓ Approval:** Setuju diterapkan sebagai perbaikan standar (tidak butuh diskusi mendalam, tapi dicatat karena mengubah perilaku dari "tanpa batas" ke "dibatasi")?

### A5. 🟠 IDOR pada `edit_task`/`update_task`
**Deskripsi:** Karyawan bisa mengedit task karyawan lain dengan mengganti `{id}` di URL.
**Rekomendasi:** Tambahkan scoping kepemilikan (`WHERE nik = current_user.nik`) di endpoint baru.
**Trade-off:** Tidak ada — ini murni bug, memperbaikinya tidak mengubah alur bisnis yang sah.
**❓ Approval:** Setuju diperbaiki (scoping kepemilikan ditambahkan)?

### A6. 🟠 Kredensial hardcoded di source (Telegram bot token, chat ID) + `.env_local` ter-commit
**Rekomendasi:** Pindahkan ke env vars/secrets manager di sistem baru; **rotasi token Telegram** karena sudah terekspos di riwayat git lama (rotasi ini menyentuh sistem lama secara operasional, bukan mengubah kode-nya — hanya invalidasi token di sisi Telegram/BotFather).
**❓ Approval:** (1) Boleh rotasi token Telegram bot sekarang (di luar kode, via BotFather) untuk menutup eksposur? (2) Siapa yang pegang akses BotFather untuk itu?

### A7. 🟡 Query dengan pola SQL-injection-shaped (`whereRaw` string concatenation)
**Deskripsi:** `gethistori`, dashboard, `task()` filter tanggal via string concatenation mentah dari request.
**Rekomendasi:** Sistem baru otomatis aman (parameterized query via GORM/sqlc) — tidak perlu keputusan khusus, hanya dicatat sebagai bukti kenapa pola ini tidak boleh direplikasi 1:1.
**❓ Approval:** Tidak perlu — informasional.

### A8. 🟡 Tidak ada otorisasi berbasis role pada panel admin
**Deskripsi:** Semua akun admin (`auth:user`), apapun `role_id`-nya, bisa akses semua fitur admin.
**Rekomendasi:** Jika `role_id` dimaksudkan untuk membatasi akses (mis. admin biasa vs superadmin), sistem baru bisa menegakkan itu.
**❓ Approval:** Apakah role-based access control ini memang **diinginkan** untuk sistem baru, atau semua admin memang sengaja punya akses penuh yang sama (dan `role_id` hanya label)?

---

## B. Logika Bisnis / Bug Fungsional

### B1. 🔴 Dua alur check-in paralel (`PresensiController` vs `AbsensiController`) tanpa koordinasi
**Dampak:** Data presensi bisa jadi tidak konsisten/duplikat kalau karyawan memakai kedua alur di hari yang sama; dua controller punya aturan cooldown berbeda (30 menit vs 5 menit) dan hardcode koordinat kantor yang terpisah.
**❓ Approval:** Sistem baru sebaiknya **hanya punya satu alur check-in** (unifikasi). Apakah alur target adalah yang berbasis shift (`AbsensiController`/EOS, karena lebih baru dan punya konsep jadwal kerja), atau ada requirement lain yang belum saya lihat? Ini keputusan besar yang menentukan desain skema/API Phase 1.

### B2. 🔴 Race condition pada check-in/check-out (tidak ada transaction/lock/unique constraint)
**Rekomendasi:** Sistem baru pakai unique constraint DB (`nik + tgl_presensi [+ shift]`) + transaction dengan row lock, menggantikan pola check-then-act.
**Trade-off:** Tidak ada downside — ini murni perbaikan keandalan, tidak mengubah aturan bisnis yang terlihat user.
**❓ Approval:** Setuju diperbaiki sebagai bagian standar migrasi (tidak mengubah perilaku yang terlihat, hanya menghilangkan race condition)?

### B3. 🟠 Ambang keterlambatan tidak konsisten (09:00 vs 09:15) dan tidak berbasis shift aktual karyawan
**Dampak:** Karyawan melihat status "terlambat" berbeda di halaman riwayat pribadi vs laporan admin. Karyawan shift non-standar dinilai terlambat berdasarkan jam kantor umum, bukan jam shift mereka.
**Rekomendasi:** Satu sumber kebenaran untuk ambang keterlambatan, idealnya dari `jam_kerja`/`jadwal_kerja` per karyawan, bukan hardcode.
**❓ Approval:** (1) Ambang mana yang benar — 09:00 atau 09:15 — untuk karyawan **tanpa** shift khusus? (2) Setuju keterlambatan dihitung berbasis jam masuk shift yang ditugaskan (bila ada), bukan angka tunggal untuk semua orang?

### B4. 🟠 Tidak ada overtime/pulang-cepat sama sekali
**❓ Approval:** Apakah sistem baru **perlu** menambahkan perhitungan lembur/pulang-cepat (fitur baru), atau tetap seperti sistem lama (tidak dihitung sama sekali)? Ini di luar "port 1:1" murni — mohon konfirmasi eksplisit karena ini penambahan fitur, bukan perbaikan bug.

### B5. 🟠 Shift overnight (`SH03`) hanya ditangani di alur `AbsensiController`, tidak di alur lama/WFH
**Terkait B1** — kalau alur disatukan, logika overnight (menutup baris kemarin, bukan membuat baris baru) harus jadi bagian universal dari alur check-in, bukan spesifik departemen/controller.
**❓ Approval:** Konfirmasi bahwa logika overnight `SH03` (cek baris kemarin dengan `jam_out IS NULL`, tutup itu alih-alih insert baru) adalah **perilaku yang benar dan harus dipertahankan** di sistem baru?

### B6. 🟠 Izin/sakit tidak punya approval workflow dan tidak memengaruhi laporan kehadiran
**❓ Approval:** (1) Apakah approval workflow untuk izin/sakit **perlu ditambahkan** di sistem baru (fitur baru), atau cukup port apa adanya (auto-approved / hanya tercatat)? (2) Apakah hari izin/sakit **perlu** dikecualikan dari status "tidak hadir" di rekap laporan (perbaikan), atau biarkan seperti sekarang?

### B7. 🟡 `buatsakit()` (form sakit) tidak punya handler store — fitur buntu
**❓ Approval:** Apakah form "buat sakit" terpisah ini memang seharusnya ada (dan perlu dilengkapi di sistem baru), atau cukup digabung ke alur `storeizin` yang sudah ada (dengan `status='s'`)?

### B8. 🟡 Foto check-in bisa tertimpa (filename collision) pada check-in kedua di hari yang sama
**Rekomendasi:** Sistem baru pakai nama file unik per event (timestamp/UUID), bukan `nik-tanggal-status`.
**Trade-off:** Tidak ada downside fungsional.
**❓ Approval:** Setuju diperbaiki sebagai bagian standar (foto historis tidak boleh saling menimpa)?

### B9. 🟡 Tidak ada auto-close untuk shift overnight yang tidak pernah checkout (tidak ada cron)
**❓ Approval:** Apakah sistem baru perlu job terjadwal untuk auto-close/flag shift yang "menggantung" (jam_out tetap NULL lebih dari X jam)? Ini penambahan fitur operasional, bukan sekadar port.

### B10. ⚪ Dua tabel shift-config yang tumpang tindih (`jadwal_kerja` per-tanggal vs `konfigurasi_jamkerja` mingguan) tanpa rekonsiliasi
**❓ Approval:** Yang mana yang jadi "sumber kebenaran" ketika keduanya punya entri untuk karyawan+tanggal yang sama? (Perlu aturan resolusi konflik yang jelas untuk desain skema Phase 1.)

---

## C. Keputusan Migrasi Data (§8 CLAUDE.md)

### C1. 🔴 Fresh start vs port data historis
**Konteks:** Tabel inti tidak punya migration Laravel — skema harus direkonstruksi dari live DB dump. Data presensi lama tidak punya `created_at`/`updated_at`/soft-delete, jadi historinya "flat".
**❓ Approval:**
1. Apakah data presensi/karyawan/departemen historis **perlu dipindahkan** ke PostgreSQL baru, atau sistem baru mulai dari data kosong (fresh start) dan data lama tetap diakses read-only di sistem lama untuk kebutuhan historis?
2. Jika port data: siapa yang bisa memberi akses ke dump database MySQL/MariaDB live saat ini (skema tidak ada di repo)?
3. Bagaimana menangani baris data yang sudah "cacat" secara bisnis (mis. foto 0-byte dari validasi gagal, `presensi` duplikat dari race condition, `jam_out` shift overnight yang menggantung)? Dibersihkan sebelum migrasi, atau diporting apa adanya dengan flag "data historis, tidak terjamin akurat"?

### C2. 🟡 Rotasi kredensial (lihat A6) sebagai bagian dari migrasi
Sudah dicakup di A6 — digabung di sini sebagai pengingat checklist migrasi.

---

## D. Ringkasan Pertanyaan yang Menunggu Jawaban (checklist cepat)

- [ ] A1 — Geofencing: tegakkan? radius berapa? kasus dinas luar?
- [ ] A2 — Setuju 4 route maintenance tanpa auth tidak diporting?
- [ ] A3 — Setuju bug reset password "12345" diperbaiki? Strategi reset password saat migrasi?
- [ ] A4 — Setuju rate limiting login ditambahkan (standar)?
- [ ] A5 — Setuju IDOR task diperbaiki (standar)?
- [ ] A6 — Boleh rotasi token Telegram bot sekarang? Siapa pemegang akses?
- [ ] A8 — Role-based access control diinginkan, atau semua admin memang setara?
- [ ] B1 — **Unifikasi alur check-in** — pakai alur berbasis shift (EOS) sebagai basis?
- [ ] B2 — Setuju race condition diperbaiki (standar, transaction+unique constraint)?
- [ ] B3 — Ambang telat yang benar: 09:00 atau 09:15? Berbasis shift per karyawan?
- [ ] B4 — Overtime/pulang-cepat: fitur baru atau tetap tidak dihitung?
- [ ] B5 — Konfirmasi logika overnight `SH03` dipertahankan sebagai perilaku benar?
- [ ] B6 — Approval workflow izin/sakit: fitur baru atau tetap auto-accept? Exclude dari "tidak hadir" di rekap?
- [ ] B7 — Form "buat sakit" terpisah: lengkapi atau gabung ke `storeizin`?
- [ ] B8 — Setuju filename foto dibuat unik (standar)?
- [ ] B9 — Cron auto-close shift menggantung: fitur baru atau skip?
- [ ] B10 — Sumber kebenaran shift: `jadwal_kerja` atau `konfigurasi_jamkerja` kalau konflik?
- [ ] C1 — Fresh start vs port data historis? Akses dump DB?
- [ ] C2 — Timing rotasi kredensial Telegram.

> Setelah pemilik project menjawab checklist ini, jawaban akan dicatat di `DECISIONS.md` dan dipakai untuk mengunci desain skema/kontrak API di Phase 1.
