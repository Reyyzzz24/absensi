# LOGIC_SPEC.md — Spesifikasi Logika Sistem Absensi (Laravel, as-is)

> Hasil audit Phase 0. Sumber: `/Users/eprisi/Documents/ABSENSI/absensi` (working tree, termasuk perubahan belum di-commit per 2026-08-26).
> Dokumen ini menjelaskan **apa yang benar-benar dilakukan sistem sekarang**, termasuk bug/inkonsistensi — bukan bagaimana seharusnya. Rekomendasi perbaikan ada di `AUDIT_FINDINGS.md`.

## 0. Peta arsitektur saat ini

Ada **dua alur check-in paralel yang sama-sama live**, menulis ke tabel `presensi` yang sama:

- **`PresensiController`** — alur lama: check-in kantor (`presensi.create/store`) + WFH (`presensi.wfh/wfhstore`) + task + izin/sakit + laporan + monitoring.
- **`AbsensiController`** ("EOS", `resources/views/eos/`) — alur baru (belum di-commit ke git): sadar jadwal shift (`jadwal_kerja`/`jam_kerja`), reverse-geocoding, kirim foto ke chat Telegram berbeda.

Tidak ada yang mencegah karyawan memakai kedua alur (`/presensi/create` dan `/absensi/create`) di hari yang sama.

**Catatan penting untuk migrasi:** tabel domain inti (`karyawan`, `presensi`, `departemen`, `jam_kerja`, `jadwal_kerja`, `konfigurasi_jamkerja`, `konfigurasi_lokasi`, `pengajuan_izin`, `role`) **tidak punya file migration Laravel** — skema asli harus direkonstruksi dari dump database live, bukan dari repo.

## 1. Timezone

- Semua timestamp server pakai `config/app.php` → `timezone = 'Asia/Jakarta'`, via `date()`/`strtotime()` PHP biasa (tidak eksplisit per-request).
- `jam_in`/`jam_out`/`tgl_presensi` **selalu digenerate server-side** saat insert (`date("Y-m-d")`/`date("H:i:s")`) — klien tidak pernah mengirim timestamp untuk check-in/out (baik untuk anti-tamper).
- Kolom datetime disimpan sebagai string naive (tanpa offset TZ) — implisit Asia/Jakarta. Tidak ada isu DST (Indonesia tidak pakai DST).
- JS `new Date()` di view hanya untuk jam berjalan kosmetik, tidak dikirim ke server.

## 2. Keterlambatan / pulang cepat / lembur

- **Tidak ada overtime dan tidak ada perhitungan pulang-cepat sama sekali.** Hanya "terlambat" (check-in saja) yang dihitung.
- Ambang batas terlambat **hardcoded dan tidak konsisten**:
  - `09:15` — `DashboardController::index()`/`dashboardadmin()`, `cetakrekap.blade.php`, `cetakrekapexcel.blade.php`, `cetaklaporanpdf.blade.php`.
  - `09:00` — `gethistori.blade.php` (riwayat pribadi karyawan pakai ambang berbeda dari dashboard/laporan admin).
- Tabel `jam_kerja` (jam masuk/pulang per shift) ditampilkan di UI EOS tapi **tidak pernah dipakai untuk menghitung keterlambatan** — ambang hardcoded dipakai rata untuk semua karyawan tanpa memandang shift mereka.
- Durasi kerja dihitung via fungsi `selisih($jam_masuk, $jam_keluar)` yang **diduplikasi verbatim di 6 file blade**; tidak menangani shift lintas tengah malam (durasi negatif), rounding via `explode(".", ...)` (rawan locale desimal).
- Tidak ada field/kolom lembur di database.

## 3. Shift

- Hanya alur `AbsensiController`/EOS yang punya konsep shift, lewat 3 tabel:
  - `jam_kerja` — katalog shift (`kode_jam_kerja`, `nama_jam_kerja`, `jam_masuk`, `jam_pulang`).
  - `jadwal_kerja` — penugasan shift per karyawan per tanggal (`nik`, `tanggal`, `kode_jam_kerja`).
  - `konfigurasi_jamkerja` — default shift mingguan per karyawan per hari (`nik`, `hari`, `kode_jam_kerja`) — tumpang tindih dengan `jadwal_kerja`, tidak direkonsiliasi di kode manapun.
- Kode shift dikenal: `SH01`, `SH02`, `SH03`, `LBR` (libur) — magic string, bukan enum/lookup terpusat.
- **Shift lintas tengah malam** hanya ditangani untuk `SH03`, hanya di `AbsensiController`: cek baris `presensi` kemarin dengan `kode_jam_kerja='SH03' AND jam_out IS NULL` → jika ada, aksi hari ini justru menutup (`jam_out`) baris kemarin, bukan membuat baris baru.
  - `PresensiController` (alur lama/WFH) **tidak punya logika ini sama sekali** — karyawan shift overnight yang pakai alur lama tidak bisa clock-out dengan benar.
- Tidak ada dukungan multi-shift per hari, kecuali departemen `GA` yang punya branch khusus mengizinkan pasangan in/out berulang di tanggal sama (dengan cooldown).

## 4. Geofencing / GPS

- Formula jarak (haversine/law-of-cosines) diduplikasi identik di `PresensiController` dan `AbsensiController`.
- **Koordinat kantor hardcoded** (`-6.216185261706369, 106.76481990185137`) di 3 lokasi kode — tabel `konfigurasi_lokasi` (lokasi + radius, bisa diedit admin) **tidak pernah dibaca** oleh kode check-in manapun. Fitur edit lokasi kantor di UI admin **tidak berefek apapun**.
- **Radius yang dihitung tidak pernah ditegakkan.** `$radius` dihitung lalu tidak pernah dicek dengan `if ($radius > X) reject`. Karyawan bisa check-in dari lokasi manapun di dunia. Lingkaran radius di peta UI murni visual klien.
- Tidak ada deteksi mock-location/fake-GPS sama sekali.
- WFH flow identik dengan flow kantor (geofencing sama, yang toh tidak ditegakkan) — bedanya cuma prefix teks `"-WFH-"` di notifikasi Telegram.

## 5. Anti-duplikat check-in / idempotency

- **Tidak ada unique constraint DB.** Pencegahan murni level aplikasi via check-then-act (`SELECT count()`/`first()` sebelum `insert()`), race-condition-prone — tidak ada `DB::transaction()`, `lockForUpdate()`, atau unique index.
- Cooldown antara aksi in/out: **30 menit** di `PresensiController`, tapi **hanya 5 menit** di `AbsensiController` (komentar kode di sana bahkan salah, masih menulis "1800 seconds = 30 minutes" padahal nilainya 300). Inkonsisten antara dua controller.
- Departemen `GA` punya state machine idempotency sendiri (selalu insert baris baru, cek `jam_out` dari baris terakhir global, bukan discope ke hari ini) — beda logika dari departemen lain.
- Karena dua controller menulis ke tabel yang sama tanpa saling tahu, karyawan bisa memicu baris duplikat/inkonsisten dengan memakai `/presensi/store` lalu `/absensi/store` (atau sebaliknya) di hari yang sama.

## 6. Libur / akhir pekan / izin / sakit

- `pengajuan_izin` menyimpan permintaan izin/sakit (`status`: `'i'`=izin, `'s'`=sakit). **Tidak ada alur approval** — insert langsung diterima apa adanya; halaman admin (`izinsakit.blade.php`) hanya filter, tidak ada aksi accept/reject.
- `buatsakit()` (form) **tidak punya handler store** — jalan buntu/fitur tidak lengkap.
- Izin/sakit **tidak pernah dipakai** dalam perhitungan keterlambatan atau laporan rekap — karyawan cuti sebulan penuh akan tampak "tidak hadir" polos tanpa penanda di grid rekap bulanan.
- Tidak ada tabel kalender libur. Kode shift `LBR` (libur) murni field tampilan, tidak pernah dipakai untuk exclude hari dari perhitungan manapun.
- Ada query ringkasan task/izin yang **dikomentari (dead code)** di `DashboardController` — kemungkinan fitur dashboard yang dinonaktifkan karena bermasalah.

## 7. Otorisasi

- Dua guard terpisah: `karyawan` (self-service) dan `user` (admin), tanpa role/permission granular (tidak ada `app/Policies/*`).
- Endpoint self-service selalu derive `$nik` dari user yang login — aman dari IDOR untuk endpoint tsb.
- **IDOR nyata**: `edit_task`/`update_task` mengambil/mengubah task berdasarkan `$id` mentah **tanpa cek kepemilikan** — karyawan manapun bisa lihat/edit task karyawan lain via ubah `{id}` di URL.
- Grup admin (`auth:user`) tidak punya pengecekan role sama sekali — semua akun admin, terlepas `role_id`, bisa akses semua route admin.
- `tampilkanpeta`/`tampilkanpeta2` (admin) menerima `presensi.id` sembarang tanpa scoping — ini disengaja (fitur monitoring), tapi tanpa audit log siapa yang melihat apa.
- `cetaklaporan`/`cetakrekap`/`cetakrekaptask` menerima `nik` sembarang dari body tanpa validasi keberadaan → bisa memicu error 500 (null dereference) dengan `nik` tidak valid.

## 8. Audit trail / soft delete

- Tabel `presensi` **tidak punya timestamps maupun soft delete**. Semua tulis pakai `DB::table()` mentah (bypass Eloquent sepenuhnya — model `Presensi` adalah dead code, bahkan `$primaryKey` yang dideklarasikan salah/usang).
- Semua delete (karyawan, user, departemen, jam kerja) adalah **hard delete**, tidak ada log aplikasi (`Log::` tidak pernah dipanggil di controller manapun).
- Satu-satunya "jejak" adalah notifikasi Telegram saat check-in/out — bukan audit log yang bisa diquery.

## 9. Password hashing

- Bcrypt standar Laravel (`config/hashing.php`), `Hash::make()` dipakai konsisten untuk create/update.
- **Bug serius**: `KaryawanController::update()` dan `UserController::update()` **selalu** reset password ke literal hardcoded `"12345"` setiap kali admin mengedit data karyawan/user apapun (termasuk sekadar ubah nomor telepon) — password lama tertimpa diam-diam.
- Password awal akun baru juga `"12345"` hardcoded (untuk create — ini pola "default password" yang wajar, bedanya dengan bug di atas adalah reset-on-every-edit).
- Hanya `PresensiController::updateprofile()` (self-service) yang benar: password hanya di-hash ulang jika field diisi.

## 10. Rate limiting login

- **Tidak ada sama sekali.** Middleware `throttle` hanya dipasang di grup `api`, bukan `web` (tempat semua route login berada). Tidak ada `RateLimiter::` dipanggil dimanapun. Login karyawan maupun admin rentan brute-force tanpa batas.

## 11. Validasi input

- **Tidak ada FormRequest class sama sekali** di seluruh codebase. Validasi inline `$request->validate()` dipakai sangat terbatas.
- Submit check-in hanya validasi 3 field (`lokasi`, `status1`, `image` — semua `required`, tanpa validasi format/isi):
  ```php
  $request->validate(['lokasi'=>'required','status1'=>'required','image'=>'required']);
  ```
  - `lokasi` tidak divalidasi format `"lat,lng"` — string cacat lolos dan menghasilkan NaN/0 di kalkulasi jarak.
  - `image` tidak divalidasi sebagai base64 PNG valid — string cacat menghasilkan file 0-byte.
  - Tidak ada batas ukuran/MIME.
- Endpoint task/izin/CRUD admin **tanpa validasi sama sekali** — insert apapun langsung ke DB.
- Error DB pada beberapa CRUD admin ditangkap generic try/catch dan ditampilkan sebagai "Data Gagal Disimpan" (menelan detail error validasi).

## 12. Performa query

- Setiap request check-in memicu 5-7 query hampir identik yang diduplikasi lintas method (`create`, `wfh`, `store`, `wfhstore`), bukan disentralisasi.
- `AbsensiController::store()` memanggil Nominatim (reverse geocoding) **secara sinkron** setiap check-in/out, tanpa cache/timeout/retry — memblokir response request pada layanan pihak ketiga.
- Telegram `Http::post()` juga sinkron di setiap check-in — memperlambat response, dan jika gagal setelah `Storage::put` sukses, tidak ada rollback.
- `cetakrekap()` pakai pivot SQL 31-kolom (`MAX(IF(DAY(tgl_presensi)=N,...))` hardcoded 1..31) — rapuh untuk bulan pendek, sulit di-maintain.
- **Pola SQL injection-shaped**: `gethistori()`, `DashboardController::index()`, `task()` pakai `whereRaw()` dengan string concatenation langsung dari `$request` tanpa binding parameter — risiko injeksi nyata tergantung karakter yang bisa lolos.

## 13. Penyimpanan file/foto

- Foto check-in disimpan via `Storage::put()` ke disk `local` default, path `storage/app/public/uploads/absensi/{nik}-{tgl}-{in|out}.png` (alur lama) atau dengan suffix random 8-karakter (alur EOS). Perlu symlink `public/storage` agar bisa diakses web.
- **Nama file di alur lama collision-prone**: `nik-tanggal-in.png`/`nik-tanggal-out.png` — check-in kedua di hari yang sama (setelah cooldown, atau branch `GA` berulang) **menimpa foto sebelumnya secara diam-diam**.
- `SymlinkController::createSymlink` **tanpa middleware auth** di route `/createSymlink`.
- **4 route symlink/maintenance tanpa auth** tersedia sekaligus (`/createSymlink`, `/foo1` [shell_exec], `/create-symlink`, `/create-symlinkk` [Artisan::call via HTTP]) — semua bisa dipanggil siapapun tanpa login.

## 14. Scheduled jobs / cron

- **Kosong total.** `app/Console/Kernel.php` tidak punya schedule aktif, tidak ada custom Artisan command.
- Tidak ada auto-close shift overnight yang tidak checkout, tidak ada auto-mark absen, tidak ada notifikasi harian/bulanan otomatis. Shift `SH03` yang tidak pernah checkout tetap `jam_out = NULL` selamanya kecuali karyawan check-in lagi.

## 15. Rute terkait absensi (ringkasan)

Lihat laporan lengkap di riwayat percakapan/PR description untuk tabel rute detail (method, URI, controller, middleware). Poin kunci:
- `auth:karyawan` untuk semua self-service (presensi, absensi/EOS, wfh, profile, histori, izin, task).
- `auth:user` untuk semua admin (karyawan, departemen, users, monitoring, laporan, rekap, konfigurasi).
- 4 route maintenance tanpa auth (lihat §13).
- Route `edit_task`/`update_task` tanpa scoping kepemilikan (lihat §7).

## 16. Kredensial & rahasia yang ditemukan di source (perlu tindakan sebelum publish/migrasi)

- Token bot Telegram hardcoded di 13+ lokasi across `PresensiController.php`/`AbsensiController.php` (chat ID berbeda antara dua controller).
- `.env_local` ter-track di git (berisi APP_KEY, default koneksi DB dsb).
- `.env` di working tree berisi nilai yang tampak produksi (`APP_URL=https://apismagnusinformatika.com/dev/erp/absensi`, `DB_DATABASE=absensi`).

> **Tindakan disarankan (perlu approval pemilik project):** rotasi token Telegram bot sebelum/selama migrasi, karena sudah ter-expose di riwayat git.
