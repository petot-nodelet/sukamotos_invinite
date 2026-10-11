# Product Requirements Document (PRD)

## SUKAMOTO Wedding Invitation

**Versi:** 1.2\
**Status:** Draft untuk validasi\
**Platform:** Web application\
**Backend utama:** Golang\
**Bahasa produk:** Bahasa Indonesia

------------------------------------------------------------------------

## 1. Ringkasan Produk

SUKAMOTO Wedding Invitation adalah platform undangan pernikahan digital
yang memungkinkan pelanggan memilih paket dan template, mengisi data
pernikahan, membayar pesanan, lalu menerbitkan undangan melalui tautan
personal.

Contoh URL undangan:

`https://muklisin-dan-rina.invite.sukamoto.web.id`

Platform menyediakan tiga tingkat layanan:

-   **Silver --- Simple & Elegant:** undangan berbasis template dengan
    personalisasi terbatas.
-   **Gold --- Personalized & Premium:** template dengan pilihan
    kustomisasi lebih luas.
-   **Platinum VIP --- Fully Custom:** layanan desain khusus berdasarkan
    brief pelanggan.

Produk harus dibangun sebagai aplikasi web multi-pelanggan dengan satu
backend dan database terpusat. Setiap undangan memiliki konfigurasi dan
URL sendiri, tanpa perlu membuat aplikasi terpisah untuk tiap pasangan.

## 2. Latar Belakang dan Masalah

Pasangan yang akan menikah membutuhkan undangan yang mudah dibagikan,
terlihat profesional di perangkat seluler, dan sesuai anggaran.
Pelanggan dengan kebutuhan berbeda juga membutuhkan pilihan antara
template siap pakai, personalisasi, dan desain eksklusif.

Platform perlu mengurangi proses manual untuk pengisian data,
pembayaran, publikasi, serta pengiriman tautan, sambil tetap memberi
admin kendali atas verifikasi pembayaran dan proyek desain khusus.

## 3. Visi Produk

Menyediakan layanan undangan digital yang mudah digunakan, menarik
secara visual, dapat dikembangkan secara teknis, dan memiliki model
bisnis yang jelas dari paket ekonomis hingga layanan desain eksklusif.

## 4. Tujuan dan Indikator Keberhasilan

### Tujuan

1.  Memudahkan pelanggan membuat undangan digital tanpa pengetahuan
    teknis.
2.  Menyediakan paket dan batas fitur yang transparan.
3.  Mempercepat proses pemesanan, verifikasi, dan publikasi.
4.  Memungkinkan admin mengelola pelanggan, template, pembayaran, dan
    masa aktif dari satu dashboard.
5.  Menghasilkan halaman undangan yang cepat, responsif, dan mudah
    dibagikan.

### Indikator keberhasilan awal

Target berikut adalah usulan awal yang perlu divalidasi setelah
peluncuran:

-   Pengguna dapat menyelesaikan pengisian data inti tanpa bantuan
    admin.
-   Tidak ada undangan yang aktif akibat pembayaran yang belum
    diverifikasi.
-   Tidak ada slug undangan ganda.
-   Halaman undangan dapat digunakan pada ukuran layar ponsel umum.
-   Admin dapat menemukan pesanan berdasarkan ID, nama pelanggan, atau
    status.
-   Status masa aktif dan kedaluwarsa sesuai tanggal aktivasi yang
    tercatat.
-   Kesalahan kritis pada alur pemesanan dan publikasi dipantau dan
    diperbaiki.

## 5. Target Pengguna

### 5.1 Pelanggan Silver

Pasangan yang membutuhkan undangan siap pakai, praktis, dan ekonomis.

### 5.2 Pelanggan Gold

Pasangan yang menginginkan kontrol lebih atas tampilan, konten, acara,
dan pengalaman tamu.

### 5.3 Pelanggan Platinum VIP

Pasangan yang memiliki konsep spesifik dan membutuhkan proses konsultasi
serta desain khusus.

### 5.4 Admin

Petugas yang mengelola paket, template, pesanan, pembayaran, publikasi,
masa aktif, dan komunikasi pelanggan.

## 6. Positioning dan Strategi Paket

### 6.1 Silver --- Simple & Elegant

**Positioning:** "Undangan digital siap pakai, tinggal isi data."

-   Masa aktif: 30 hari sejak undangan diaktifkan.
-   Menggunakan template tetap.
-   Pelanggan mengubah data, foto, dan opsi warna yang diizinkan.
-   Pelanggan tidak dapat mengubah struktur layout atau membuat section
    sendiri.

Fitur: - Nama mempelai dan foto. - Nama orang tua. - Quote atau doa. -
Couple story singkat. - Informasi akad dan resepsi. - Nama tempat,
alamat, dan tautan peta. - Countdown. - RSVP dasar. - Ucapan dan doa. -
Tombol WhatsApp dan navigasi lokasi. - Musik latar dengan kontrol
pengguna. - Cover dan galeri 6--10 foto sesuai batas template. - Satu
video atau tautan YouTube. - Informasi rekening atau amplop digital. -
Tautan undangan personal.

Contoh template awal: S01 Floral Elegant, S02 Minimalist White, S03
Classic Romantic, S04 Japanese Garden.

### 6.2 Gold --- Personalized & Premium

**Positioning:** "Template premium yang bisa disesuaikan dengan
pasangan."

-   Masa aktif: 90 hari sejak undangan diaktifkan.
-   Seluruh fitur Silver.
-   Pilihan layout, warna, dan font yang tersedia.
-   Custom cover dan background dari opsi yang didukung.
-   Section dan wording yang dapat dikonfigurasi.
-   Couple story lebih fleksibel.
-   Batas galeri dan video lebih besar, ditentukan pada konfigurasi
    paket.
-   Musik dari koleksi atau unggahan yang diizinkan.
-   Custom slug jika tersedia.
-   Preview peta.
-   Multiple event, misalnya akad, resepsi, dan ngunduh mantu.
-   Live RSVP, daftar tamu, guestbook, ucapan digital, amplop digital,
    QR Code, dan social sharing preview.
-   Optimasi pengalaman mobile.

Gold tetap menggunakan sistem template dan komponen yang dikonfigurasi.
Perubahan desain di luar kemampuan template dapat ditawarkan sebagai
layanan tambahan.

### 6.3 Platinum VIP --- Fully Custom

**Positioning:** "Punya konsep sendiri? Kami wujudkan."

-   Masa aktif: enam bulan atau durasi khusus sesuai kontrak.
-   Seluruh fitur Gold, sesuai kesepakatan proyek.
-   Desain dari nol.
-   Custom layout, animasi, ilustrasi, tipografi, sistem warna,
    transisi, loading screen, musik, dan storytelling.
-   Brief desain, persetujuan ruang lingkup, jadwal, dan pengelolaan
    revisi.
-   Estimasi harga dan waktu pengerjaan berdasarkan kompleksitas.

Contoh: konsep kerajaan Jawa modern dibuat khusus, bukan sekadar memilih
template terdekat.

### 6.4 Aturan paket

-   Harga dan batas fitur harus dapat dikonfigurasi admin.
-   Paket yang dibeli harus menyimpan snapshot harga, masa aktif, dan
    hak fitur saat transaksi dibuat.
-   Perubahan harga paket tidak mengubah transaksi historis.
-   Hak fitur harus divalidasi di backend, bukan hanya disembunyikan di
    frontend.
-   Definisi "custom URL", batas galeri/video, kapasitas tamu, jumlah
    revisi, dan masa aktif Platinum harus ditetapkan sebelum peluncuran.

## 7. Ruang Lingkup MVP

Prinsip UX utama: **pelanggan tidak perlu mendaftar atau login**. Pemesanan dilakukan melalui satu alur sederhana. Sistem menyimpan data berdasarkan nomor pesanan dan kontak pelanggan, bukan akun pelanggan. Dashboard hanya disediakan untuk admin.

### Termasuk MVP

-   Landing page.
-   Halaman daftar paket dan perbandingan fitur.
-   Galeri template dan halaman demo.
-   Form pemesanan tanpa registrasi/login.
-   Nomor pesanan unik dan halaman status pesanan sederhana yang dapat dibuka melalui tautan privat/token.
-   Form data undangan; data disimpan saat dikirim, dengan opsi melanjutkan pengisian melalui tautan privat bila diperlukan.
-   Upload foto dengan validasi.
-   Preview undangan.
-   Checkout dan pembuatan pesanan.
-   Pembayaran transfer bank manual dengan verifikasi admin.
-   Status pesanan dan notifikasi admin.
-   Dashboard admin dengan daftar pesanan masuk.
-   Publikasi undangan dengan slug personal.
-   Masa aktif dan halaman kedaluwarsa.
-   Tombol bagikan melalui WhatsApp.
-   Halaman undangan responsif dengan fitur inti.

### Di luar MVP atau tahap lanjutan

-   Akun dan dashboard khusus pelanggan (tidak diperlukan pada MVP).
-   Pembayaran otomatis melalui payment gateway.
-   WhatsApp Business API untuk pengiriman otomatis.
-   Semua fitur Gold tingkat lanjut.
-   Kustomisasi desain mandiri yang kompleks.
-   Manajemen proyek Platinum lengkap.
-   Analitik lanjutan dan kampanye pemasaran otomatis.
-   Aplikasi native Android/iOS.

## 8. Alur Pengguna

### 8.1 Alur pelanggan tanpa akun

1. Pengunjung membuka halaman utama SUKAMOTO Wedding Invitation.
2. Pengunjung memilih paket: Silver, Gold, atau Platinum VIP.
3. Pengunjung melihat dan memilih template/demo yang sesuai paket. Untuk Platinum VIP, pelanggan dapat memilih referensi demo atau mengirim brief konsep.
4. Pengunjung mengisi form data mempelai, acara, kontak, slug yang diinginkan, dan informasi lain sesuai paket.
5. Sistem memvalidasi form, membuat nomor pesanan unik, dan menampilkan ringkasan pesanan serta instruksi pembayaran.
6. Pelanggan membayar melalui QRIS atau transfer bank. Untuk MVP, pembayaran diverifikasi manual oleh admin. Pelanggan dapat mengirim bukti pembayaran jika diperlukan.
7. Status pesanan menjadi `payment_review`. Pelanggan melihat informasi bahwa pesanan sedang menunggu konfirmasi admin.
8. Admin memverifikasi pembayaran dan kelengkapan data. Jika ada masalah, admin menghubungi pelanggan melalui WhatsApp.
9. Setelah pembayaran disetujui, sistem memublikasikan undangan atau menandai pekerjaan Platinum untuk diproses sesuai kesepakatan.
10. Sistem membuat URL undangan, misalnya `https://muklisin-dan-rina.invite.sukamoto.web.id`.
11. Admin mengirimkan URL kepada pelanggan melalui WhatsApp atau menyalin tautan untuk diserahkan langsung. Pelanggan tidak perlu login untuk membuka undangan.

### 8.2 Alur admin

1. Sistem menerima pesanan baru dan mengirim notifikasi ke kanal admin yang sudah dikonfigurasi sejak awal.
2. Kanal notifikasi MVP dapat berupa notifikasi di dashboard admin dan/atau notifikasi browser ketika dashboard admin terbuka. Integrasi WhatsApp otomatis membutuhkan layanan/API resmi dan menjadi opsi lanjutan; untuk MVP, admin dapat membuka chat WhatsApp pelanggan dari tombol dengan pesan siap kirim.
3. Admin membuka pesanan berdasarkan nomor pesanan, memeriksa data, paket, dan bukti pembayaran.
4. Admin menyetujui atau menolak pembayaran. Alasan penolakan atau data yang kurang disampaikan kepada pelanggan melalui WhatsApp.
5. Setelah disetujui, admin menerbitkan/menyetujui publikasi undangan. Untuk Silver/Gold, proses dapat diotomatisasi; untuk Platinum, admin menyelesaikan desain terlebih dahulu.
6. Sistem menampilkan URL undangan dan tombol `Kirim via WhatsApp`, `Salin Link`, serta `Tandai Terkirim`.
7. Admin mengirimkan URL kepada pelanggan dan menandai pesanan selesai.

### Status pesanan

`submitted → payment_review → payment_approved → publishing → active → link_sent`

Status alternatif: `payment_rejected`, `needs_customer_update`, `cancelled`, `expired`, `refunded` (jika pengembalian dana didukung). Status pengerjaan Platinum dipisahkan dari status pembayaran.

**Prinsip penting:** notifikasi tidak boleh menjadi satu-satunya catatan pesanan. Semua pesanan dan perubahan status wajib tersimpan di database, sehingga pesanan tetap dapat diproses jika notifikasi gagal.

## 9. Kebutuhan Fungsional

### FR-01: Landing Page

Sistem menampilkan penjelasan produk, paket, template, demo, cara kerja,
FAQ, kontak, dan CTA yang jelas.

**Kriteria penerimaan:** - Pengguna dapat membuka detail paket dan demo
template. - CTA utama menuju alur pemesanan. - Tampilan responsif pada
desktop dan ponsel.

### FR-02: Paket dan Hak Fitur

Sistem menampilkan harga, durasi, fitur, dan batasan setiap paket.

**Kriteria penerimaan:** - Harga berasal dari konfigurasi backend. - Hak
fitur diperiksa server-side. - Paket dapat diubah admin tanpa mengedit
banyak file kode.

### FR-03: Template dan Demo

Sistem menampilkan galeri, kategori, preview, dan template yang tersedia
untuk paket tertentu.

**Kriteria penerimaan:** - Setiap template memiliki ID unik, nama, tema,
thumbnail, demo, dan status aktif. - Template yang dinonaktifkan tidak
dapat dipilih untuk pesanan baru. - Template yang sudah digunakan tetap
dapat dirender untuk undangan aktif yang sah.

### FR-04: Pemesanan Tanpa Akun

Pelanggan dapat memesan tanpa registrasi atau login. Sistem membuat nomor pesanan unik dan token privat yang cukup sulit ditebak untuk melihat status pesanan atau melanjutkan form jika fitur tersebut diperlukan. Dashboard admin tetap wajib menggunakan autentikasi dan role admin.

**Kriteria penerimaan:** - Tidak ada kewajiban membuat akun pelanggan. - Nomor pesanan bukan satu-satunya rahasia untuk mengakses data; tautan status menggunakan token acak yang aman. - Halaman status hanya menampilkan informasi minimum yang diperlukan. - Endpoint admin memerlukan autentikasi dan role admin.

### FR-05: Form Undangan

Pelanggan dapat mengisi informasi mempelai, orang tua, cerita, acara,
lokasi, media, dan pengaturan yang didukung paket.

**Kriteria penerimaan:** - Form memiliki validasi server-side. - Draft
dapat disimpan dan dilanjutkan. - Field premium dikunci dengan
penjelasan yang jelas jika tidak termasuk paket. - Pengguna dapat
melihat indikator kelengkapan data.

### FR-06: Media

Pelanggan dapat mengunggah media sesuai batas paket.

**Kriteria penerimaan:** - Validasi ukuran, MIME type, dan format file
dilakukan di server. - Nama file disimpan dengan nama aman yang dibuat
sistem. - Upload tidak dapat mengeksekusi file sebagai kode. - Batas
kapasitas dan jumlah file dapat dikonfigurasi.

### FR-07: Preview

Pelanggan dapat melihat pratinjau undangan sebelum checkout dan setelah
perubahan.

**Kriteria penerimaan:** - Preview menggunakan data aktual draft. -
Tampilan preview mengikuti template dan hak fitur paket. - Preview
privat tidak dapat diakses tanpa token atau otorisasi yang sesuai.

### FR-08: Pesanan dan Pembayaran

Sistem membuat pesanan dengan ringkasan paket, harga, dan status
pembayaran.

**Kriteria penerimaan:** - Harga dihitung dan diverifikasi backend. -
Bukti transfer disimpan secara aman dan hanya dapat dilihat oleh pihak
berwenang. - Admin dapat menyetujui atau menolak pembayaran dengan
catatan. - Aktivasi tidak dapat terjadi hanya karena pelanggan mengubah
request. - Webhook pembayaran, jika diintegrasikan, harus diverifikasi
dan idempotent.

### FR-09: Publikasi Undangan

Sistem menerbitkan undangan pada subdomain berdasarkan slug unik.

**Kriteria penerimaan:** - Slug divalidasi dan unik. - Hostname
dipetakan ke undangan yang benar. - Undangan aktif dapat diakses
publik. - Undangan yang kedaluwarsa menampilkan halaman kedaluwarsa. -
Undangan yang tidak ditemukan menampilkan 404.

### FR-10: Masa Aktif dan Perpanjangan

Sistem menghitung masa aktif berdasarkan waktu aktivasi yang sah.

**Kriteria penerimaan:** - Silver aktif 30 hari, Gold 90 hari, Platinum
enam bulan atau sesuai kontrak. - Zona waktu dan aturan perhitungan
tanggal didokumentasikan. - Perpanjangan tercatat sebagai transaksi atau
perubahan kontrak yang dapat diaudit. - Pekerjaan terjadwal menandai
atau mengingatkan undangan yang mendekati kedaluwarsa.

### FR-11: RSVP dan Guestbook

Sistem menerima RSVP dan ucapan sesuai fitur paket.

**Kriteria penerimaan:** - Pemilik undangan dapat melihat data yang
terkait dengan undangannya. - Validasi dan rate limiting mengurangi
spam. - Admin atau pemilik undangan dapat menyembunyikan pesan yang
tidak pantas. - Data tamu tidak terekspos ke undangan lain.

### FR-12: Status Pesanan Pelanggan

Pelanggan tidak memiliki dashboard akun. Pelanggan dapat melihat status terbatas melalui halaman status yang menggunakan token privat dari konfirmasi pemesanan. Sistem juga menampilkan instruksi pembayaran dan pesan bahwa pesanan sedang menunggu konfirmasi admin. URL undangan final dikirim admin melalui WhatsApp atau diserahkan langsung.

### FR-13: Dashboard Admin

Admin dapat mengelola paket, template, pesanan, pembayaran, undangan, masa aktif, notifikasi, dan proyek Platinum. Daftar pesanan harus tetap tersimpan meskipun notifikasi tidak terkirim.

### FR-14: Notifikasi Admin dan WhatsApp

MVP menyediakan notifikasi dalam dashboard admin dan dukungan notifikasi browser ketika dashboard terbuka serta izin browser diberikan. Setiap pesanan juga ditampilkan dalam antrean pesanan baru. Tombol WhatsApp membuka chat pelanggan dengan teks siap kirim, termasuk nomor pesanan dan link undangan setelah tersedia. Pengiriman pesan WhatsApp otomatis memerlukan integrasi resmi dan menjadi fitur lanjutan.

### FR-15: SEO dan Social Sharing

Halaman publik memiliki metadata yang sesuai, Open Graph, dan gambar
preview jika tersedia.

Dashboard, admin, data privat, dan halaman yang memuat data sensitif
tidak boleh diindeks.

## 10. Kebutuhan Nonfungsional

### Performa

-   Gunakan gambar yang dioptimalkan dan lazy loading untuk media di
    bawah lipatan.
-   Hindari animasi yang menghambat interaksi utama.
-   Target awal Core Web Vitals dan waktu respons ditetapkan serta diuji
    di staging sebelum rilis.
-   Musik tidak boleh memblokir akses konten dan harus memiliki kontrol
    pengguna.

### Responsivitas dan aksesibilitas

-   Mobile-first untuk ponsel Android dan iPhone.
-   Navigasi dan tombol mudah disentuh.
-   Kontras teks memadai.
-   Form memiliki label dan pesan validasi.
-   Interaksi penting dapat digunakan dengan keyboard.

### Keamanan dan privasi

-   Proteksi CSRF, XSS, SQL injection, brute force, dan rate abuse
    sesuai konteks.
-   Validasi backend untuk semua input penting.
-   Role-based access control dan pemeriksaan kepemilikan data.
-   Secret disimpan di environment atau secret manager.
-   Audit log untuk tindakan admin penting.
-   Backup database dan prosedur pemulihan.
-   Kebijakan retensi serta penghapusan data yang jelas.
-   Data pribadi dan nomor rekening tidak ditampilkan di tempat yang
    tidak diperlukan.

### Keandalan

-   Error ditangani tanpa membocorkan stack trace kepada pengguna.
-   Operasi pembayaran dan aktivasi aman terhadap request berulang.
-   Migration database dan prosedur rollback didokumentasikan.
-   Log operasional tidak menyimpan password, token, atau data
    pembayaran sensitif.

## 11. Arsitektur Teknis

### Stack yang disarankan

-   Go dengan router Gin atau Chi.
-   **SQLite sebagai database MVP**, dengan driver Go yang dipilih dan diuji untuk arsitektur Armbian target.
-   Lapisan akses data yang jelas; gunakan `database/sql` atau ORM ringan yang kompatibel dengan SQLite. Jangan mengikat logika bisnis langsung ke SQL vendor tertentu.
-   Go `html/template` untuk server-side rendering.
-   HTMX dan JavaScript ringan bila diperlukan.
-   **Cloudflare Tunnel (`cloudflared`) untuk akses publik ke aplikasi di STB**, tanpa membuka port aplikasi ke internet.
-   Nginx hanya jika dibutuhkan untuk routing atau fungsi lokal tambahan; jangan menambahkannya tanpa kebutuhan.
-   Cloudflare R2 atau object storage lain sebagai opsi terpisah untuk media dan backup off-device, bukan pengganti SQLite.
-   Docker opsional; service native dengan `systemd` diprioritaskan pada STB.

Keputusan MVP: Go + SQLite + Cloudflare Tunnel. Pertahankan repository/lapisan akses data dan migration yang rapi agar perpindahan ke PostgreSQL di masa depan dapat dilakukan bila kebutuhan konkurensi atau operasional meningkat. Hindari menambahkan banyak framework tanpa kebutuhan nyata.

### Struktur awal proyek

``` text
sukamoto-wedding/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── config/
│   ├── database/
│   ├── middleware/
│   ├── auth/
│   ├── users/
│   ├── packages/
│   ├── templates/
│   ├── orders/
│   ├── payments/
│   ├── invitations/
│   ├── guests/
│   ├── notifications/
│   └── admin/
├── migrations/
├── web/
│   ├── templates/
│   └── static/
├── tests/
├── scripts/
├── Dockerfile
├── compose.yaml
├── .env.example
├── .gitignore
├── go.mod
└── README.md
```

### Subdomain

Pola: `{slug}.invite.sukamoto.web.id`

Prasyarat: - DNS wildcard untuk `*.invite.sukamoto.web.id`. - Sertifikat
TLS yang mencakup subdomain tersebut. - Reverse proxy yang meneruskan
hostname dengan benar. - Middleware atau resolver yang memetakan
hostname ke undangan. - Penanganan slug yang tidak ditemukan, tidak
aktif, atau kedaluwarsa.

Jangan membuat aplikasi atau server terpisah untuk setiap pelanggan.

## 12. Model Data Awal

Tabel yang disarankan: `admin_users` (akun admin saja), `packages`, `package_features`, `templates`, `orders`, `order_access_tokens` (jika halaman status/lanjut form memakai token privat), `payments`, `invitations`, `invitation_events`, `invitation_media`, `invitation_sections`, `guests`, `rsvps`, `guestbook_messages`, `digital_gifts`, `platinum_projects`, `notifications`, `renewal_orders`, dan `audit_logs`.

Data kontak pelanggan disimpan pada pesanan atau tabel kontak terkait; tabel pelanggan tidak boleh diasumsikan sebagai akun login. Jangan membuat tabel akun pelanggan untuk MVP.

Gunakan foreign key, indeks yang relevan, timestamp, status eksplisit,
dan migration. Simpan snapshot paket/harga pada pesanan. Hindari
duplikasi data yang tidak perlu dan terapkan pembatasan akses
berdasarkan pemilik undangan.

## 13. Alur Admin

1. Sistem menyimpan pesanan baru ke database dan memasukkannya ke antrean admin.
2. Sistem memunculkan notifikasi dashboard dan, bila didukung browser serta diizinkan, notifikasi browser.
3. Admin membuka detail pesanan, memeriksa paket, data undangan, kontak pelanggan, dan bukti pembayaran.
4. Admin menyetujui pembayaran atau menandai pesanan perlu diperbaiki/ditolak.
5. Setelah pembayaran sah, sistem mengaktifkan publikasi otomatis untuk paket yang siap atau menunggu admin menyelesaikan desain Platinum.
6. Sistem membuat/menetapkan slug unik dan menampilkan URL final.
7. Admin menggunakan tombol `Kirim via WhatsApp` atau `Salin Link`; nomor pelanggan dan pesan siap kirim diisi bila tersedia.
8. Admin menandai tautan sebagai terkirim. Status ini merupakan catatan internal, bukan bukti bahwa pesan WhatsApp telah dibaca.
9. Semua perubahan status dan tindakan penting dicatat dalam audit log.

Kanal notifikasi khusus admin (misalnya browser pada perangkat admin atau akun/nomor WhatsApp admin yang telah ditetapkan sejak awal) harus dikonfigurasi sejak awal. Untuk notifikasi WhatsApp otomatis, diperlukan nomor/akun bisnis serta integrasi WhatsApp Business Platform atau penyedia resmi; jangan menganggap tombol membuka WhatsApp sebagai pengiriman otomatis.

## 14. Analitik dan Marketing

Metrik yang dapat dipantau: - Pengunjung landing page. - Klik CTA. -
Demo template dibuka. - Paket yang dipilih. - Form dimulai dan
diselesaikan. - Checkout dimulai. - Pembayaran berhasil. - Konversi per
paket. - Permintaan konsultasi Platinum. - Perpanjangan undangan.

Patuhi privasi dan minimalkan pengumpulan data. Jangan menggunakan
testimoni atau angka penjualan fiktif.

## 15. Risiko dan Mitigasi

  -----------------------------------------------------------------------
  Risiko                              Mitigasi
  ----------------------------------- -----------------------------------
  Pembayaran palsu atau aktivasi      Verifikasi admin atau webhook
  tanpa pembayaran                    resmi; aktivasi idempotent

  Slug bertabrakan                    Unique constraint database dan
                                      penanganan konflik

  Media berukuran besar               Batas ukuran, validasi tipe,
                                      optimasi, dan object storage

  Penyalahgunaan RSVP/guestbook       Rate limiting, validasi, moderasi,
                                      dan anti-spam

  Perbedaan hak fitur antar paket     Konfigurasi terpusat dan
                                      pemeriksaan backend

  Subdomain tidak berfungsi           Konfigurasi wildcard DNS, TLS, dan
                                      pengujian hostname

  Undangan kedaluwarsa tidak          Satu aturan perhitungan masa aktif
  konsisten                           dan job terjadwal

  Permintaan Platinum melebar         Brief tertulis, ruang lingkup,
                                      harga, jadwal, dan batas revisi

  Data pribadi terekspos              Pemeriksaan otorisasi, minimisasi
                                      data, dan pengujian akses
  -----------------------------------------------------------------------

## 16. Tahapan Pengembangan

### Tahap 1 --- Fondasi dan UI

-   Setup Go dan database.
-   Konfigurasi environment.
-   Landing page.
-   Daftar paket.
-   Galeri dan demo template.
-   Layout responsif.

### Tahap 2 --- Form Pemesanan Tanpa Akun

-   Form data undangan tanpa registrasi/login.
-   Nomor pesanan unik dan token status privat.
-   Validasi serta penyimpanan data form.
-   Upload media.
-   Preview undangan dan ringkasan pesanan.

### Tahap 3 --- Pembayaran dan Admin

-   Pembuatan pesanan dan instruksi QRIS/transfer bank.
-   Status pembayaran.
-   Verifikasi manual oleh admin.
-   Dashboard antrean admin.
-   Notifikasi dashboard/browser.
-   Tombol WhatsApp siap kirim dan penandaan link terkirim.

### Tahap 4 --- Publikasi

-   Aktivasi.
-   Slug dan routing subdomain.
-   Halaman undangan publik.
-   Masa aktif dan kedaluwarsa.
-   Bagikan melalui WhatsApp.

### Tahap 5 --- Fitur Lanjutan

-   Payment gateway.
-   RSVP dan guestbook tingkat lanjut.
-   QR Code.
-   Multiple event.
-   Kustomisasi Gold.
-   Workflow Platinum.

### Tahap 6 --- Peluncuran

-   Uji keamanan.
-   Uji alur end-to-end.
-   Uji responsivitas dan performa.
-   Backup dan pemulihan.
-   Deployment dan monitoring.
-   Dokumentasi operasional.

## 17. Rencana Pengujian

Minimal lakukan: - Unit test untuk validasi paket, harga, slug, dan masa
aktif. - Integration test untuk database, pesanan, pembayaran, dan
publikasi. - Security test untuk memastikan token status pesanan tidak dapat ditebak dan tidak membocorkan data pesanan lain; uji role admin dan akses dashboard admin. - Test pembayaran berulang
untuk memastikan tidak terjadi aktivasi ganda. - Test slug duplikat. -
Test undangan aktif, kedaluwarsa, dan tidak ditemukan. - Test upload
file tidak valid. - End-to-end test tanpa akun dari memilih paket sampai admin mengirim link undangan. - Uji tampilan di ukuran layar ponsel dan desktop.

## 18. Asumsi dan Keputusan yang Belum Final

**Keputusan UX:** MVP tidak mewajibkan akun/login pelanggan. Akun admin tetap wajib.

Hal berikut harus dikonfirmasi sebelum peluncuran:
1. Harga setiap paket dan biaya perpanjangan.
2. Definisi hari/bulan masa aktif dan zona waktu resmi.
3. Batas pasti foto, video, tamu, RSVP, dan penyimpanan tiap paket.
4. Detail instruksi pembayaran QRIS/transfer dan aturan bukti pembayaran.
5. Kanal notifikasi admin yang akan dikonfigurasi sejak awal.
6. Apakah custom slug termasuk Gold atau dikenakan biaya tambahan.
7. Masa aktif Platinum dan jumlah revisi sesuai kontrak.
8. Kebijakan privasi, retensi data, dan penghapusan media setelah kedaluwarsa.
9. Hosting, object storage, domain, dan pengelolaan sertifikat.
10. Apakah RSVP/guestbook tetap dapat dibaca setelah undangan kedaluwarsa.

## 19. Definisi Selesai (Definition of Done)

Sebuah fitur dianggap selesai jika: - Kebutuhan dan acceptance criteria
terpenuhi. - Validasi frontend dan backend sesuai kebutuhan. - Akses dan
kepemilikan data telah diuji. - Error dan kondisi gagal ditangani. - Tes
relevan lulus. - UI berfungsi di perangkat ponsel dan desktop. -
Dokumentasi konfigurasi diperbarui. - Tidak ada secret yang ditanam
dalam source code. - Integrasi eksternal yang belum dikonfigurasi
dinyatakan secara eksplisit.

## 20. Instruksi Implementasi untuk AI Coding Assistant

Gunakan PRD ini sebagai sumber kebutuhan utama. Kerjakan aplikasi
bertahap, dimulai dari Tahap 1.

Pada setiap tahap: 1. Jelaskan tujuan dan keputusan teknis. 2. Tampilkan
struktur file yang relevan. 3. Berikan kode lengkap untuk file yang
dibuat atau diubah. 4. Berikan perintah setup, migration, run, dan test.
5. Jelaskan konfigurasi eksternal yang diperlukan. 6. Jangan mengklaim
integrasi berfungsi jika belum benar-benar diimplementasikan dan diuji.
7. Pertahankan arsitektur modular dan kompatibilitas dengan tahap
sebelumnya. 8. Jangan mengimplementasikan fitur premium sebagai UI palsu
tanpa business logic. 9. Perbarui dokumentasi saat kontrak API, skema
database, atau konfigurasi berubah.

**Prioritas pertama:** hasilkan fondasi Golang yang dapat dijalankan,
landing page responsif, perbandingan paket, dan galeri demo template.
Setelah itu, lanjutkan ke autentikasi, dashboard, form, pembayaran, dan
publikasi undangan.


---

## 21. Target Deployment Hardware — STB HG680P

### 21.1 Deployment target

Target resmi deployment awal adalah **STB HG680P dengan Armbian Linux**. Media penyimpanan sistem dan data **belum ditentukan** dan wajib diputuskan sebelum deployment produksi.

Perangkat HG680P umumnya menggunakan Amlogic S905X, CPU ARM quad-core, RAM kelas 2 GB, penyimpanan internal yang terbatas, dan Ethernet 100 Mbps. Spesifikasi unit yang benar-benar digunakan wajib diverifikasi sebelum deployment karena konfigurasi dan kapasitas penyimpanan dapat berbeda.

Gunakan instalasi Armbian minimal/server yang benar-benar kompatibel dengan unit. Verifikasi versi OS, kernel, arsitektur userland, driver jaringan, dan cara boot aktual. Jangan mengasumsikan semua image Armbian untuk keluarga perangkat Amlogic memiliki kompatibilitas yang sama.

### 21.2 Prinsip desain untuk hardware terbatas

- Prioritaskan Go server-side rendering dengan `html/template`, CSS, JavaScript ringan, dan HTMX bila diperlukan.
- Hindari frontend berat, proses build frontend di perangkat, serta layanan latar belakang yang tidak diperlukan.
- Bangun binary Go untuk arsitektur OS target. Verifikasi hasil `uname -m` dan `go env GOARCH` pada sistem target sebelum build.
- Mulai dengan satu aplikasi Go dan satu database. Jangan menjalankan banyak container hanya demi pemisahan layanan.
- Docker bersifat opsional; deployment binary dengan `systemd` dapat menjadi pilihan yang lebih ringan.
- Jangan menyimpan banyak foto/video pelanggan pada penyimpanan internal STB. Gunakan penyimpanan eksternal yang andal atau object storage jika memungkinkan.
- Buat backup database dan konfigurasi ke perangkat atau lokasi lain; backup yang hanya berada pada STB yang sama tidak cukup untuk pemulihan bencana.
- Pantau RAM, ruang disk, suhu, uptime, log, dan status layanan.
- Atur rotasi log dan batas ukuran upload agar penyimpanan tidak cepat penuh.

### 21.3 Strategi penyimpanan

Karena media penyimpanan belum diputuskan, deployment produksi tidak boleh mengasumsikan data aplikasi aman disimpan di media boot yang sedang digunakan.

Pilihan yang perlu dievaluasi:
- **microSD:** mudah diganti dan cocok untuk eksperimen, tetapi kartu berkualitas rendah dapat cepat aus saat menerima banyak penulisan.
- **USB storage:** dapat menyediakan ruang tambahan untuk media dan backup, tetapi perlu diuji kestabilan koneksi, kualitas perangkat, dan perilaku setelah reboot.
- **eMMC internal:** dapat menjadi pilihan jika tersedia, didukung image/bootloader, dan kapasitasnya cukup; jangan menulis atau mengubah partisi tanpa backup dan prosedur pemulihan.
- **Penyimpanan eksternal/object storage:** lebih baik dipertimbangkan untuk koleksi foto/video pelanggan yang bertumbuh, jika koneksi dan anggaran mendukung.

Sebelum memilih, ukur kapasitas aktual dengan `lsblk -o NAME,SIZE,FSTYPE,MOUNTPOINTS` dan `df -h`. Tentukan lokasi terpisah untuk konfigurasi, data database, media unggahan, dan backup. Gunakan filesystem Linux yang sesuai, mount berdasarkan UUID atau label yang stabil, serta pastikan layanan tidak mulai sebelum volume data yang wajib tersedia berhasil di-mount.

Jangan menyimpan satu-satunya salinan database atau media pelanggan di microSD/USB yang sama dengan sistem operasi. Backup harus disalin ke perangkat atau lokasi lain dan proses restore harus diuji.

### 21.4 Database — SQLite untuk MVP

Database MVP yang dipilih adalah **SQLite**, sesuai rekomendasi client dan target deployment satu STB. SQLite menyimpan data dalam satu file dan tidak memerlukan proses server database terpisah, sehingga mengurangi beban operasional pada perangkat terbatas.

Data yang disimpan meliputi akun dan role, paket serta snapshot harga, pesanan, catatan status pembayaran, data undangan, slug, masa aktif, RSVP/ucapan sesuai ruang lingkup MVP, dan audit log. File foto, musik, dan video tidak disimpan sebagai BLOB di database; simpan sebagai file di lokasi media yang ditetapkan atau gunakan object storage bila diperlukan.

Persyaratan implementasi SQLite:
- Gunakan constraint dan index untuk integritas data, termasuk unique constraint pada slug dan transaksi yang harus unik.
- Gunakan transaksi untuk perubahan status pesanan, pencatatan pembayaran, publikasi, dan aktivasi. Aktivasi harus idempotent agar request berulang tidak memperpanjang masa aktif atau menggandakan transaksi.
- Atur foreign key enforcement, busy timeout, dan kebijakan koneksi yang diuji. Evaluasi WAL mode untuk pola akses baca/tulis aplikasi; jangan menganggap WAL menghilangkan batas satu penulis pada satu waktu.
- Hindari transaksi tulis panjang. Jangan mengandalkan SQLite untuk beban tulis serentak tinggi tanpa load test.
- Jangan menaruh file database aktif pada filesystem jaringan yang tidak mendukung locking SQLite secara andal.
- Jalankan migration versi-terkontrol dan buat backup yang konsisten sebelum perubahan skema.
- Backup file database aktif menggunakan mekanisme backup SQLite yang konsisten atau hentikan layanan saat menyalin; jangan mengandalkan penyalinan file utama saja saat aplikasi sedang menulis, terutama jika WAL aktif.
- Jadwalkan backup ke perangkat/lokasi terpisah, pantau hasil backup, tetapkan retensi, dan lakukan uji restore berkala.
- Simpan lokasi file database melalui konfigurasi, misalnya `/var/lib/sukamoto/data/app.db`; pastikan direktori dimiliki user service dan tidak berada di direktori publik web.

SQLite sesuai untuk MVP dengan trafik awal rendah dan pola penulisan moderat. Sebelum produksi, uji konkurensi pembelian, RSVP, ucapan, dan aktivitas admin. Pertimbangkan migrasi ke PostgreSQL atau database terkelola jika muncul lock contention, banyak penulisan bersamaan, kebutuhan replikasi/high availability, atau operasi multi-server. Migrasi bukan asumsi otomatis; lakukan berdasarkan metrik dan kebutuhan nyata.

### 21.5 Arsitektur dan build Golang

Gunakan aplikasi Go native yang berjalan sebagai service Linux. Untuk MVP, utamakan server-side rendering dengan `html/template`, CSS, JavaScript ringan, dan satu proses aplikasi yang menangani routing, autentikasi, pesanan, dan undangan.

Sebelum build, verifikasi arsitektur userland pada STB:
- `uname -m` untuk arsitektur kernel.
- `dpkg --print-architecture` pada sistem berbasis Debian/Armbian.
- `go version` dan `go env GOOS GOARCH` jika Go tersedia di perangkat.

Jangan menyimpulkan target binary hanya dari nama chipset. Jika perangkat melaporkan ARM 64-bit dan userland mendukung ARM64, build untuk `GOOS=linux GOARCH=arm64`. Jika userland 32-bit, pastikan ABI dan dukungan Go sebelum memilih `GOARCH=arm` dan `GOARM` yang sesuai. Bangun binary di komputer pengembangan atau CI bila lebih praktis, lalu uji binary pada perangkat target.

Contoh build untuk sistem Linux ARM64 (gunakan hanya setelah arsitektur dikonfirmasi):

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o sukamoto-server ./cmd/server
```

Jika driver database atau dependensi memerlukan CGO, jangan gunakan `CGO_ENABLED=0` secara membabi buta; siapkan toolchain silang yang cocok dan uji pada OS target.

Untuk layanan native, buat unit `systemd` dengan user non-root khusus, `WorkingDirectory` yang benar, environment file berizin ketat, `Restart=on-failure`, dan direktori data yang eksplisit. Jangan menjalankan aplikasi sebagai root jika tidak diperlukan.

### 21.5 Deployment dan operasi

Kebutuhan deployment:
- Build binary Go untuk arsitektur target.
- File konfigurasi melalui environment atau file konfigurasi dengan permission yang aman.
- Service `systemd` agar aplikasi berjalan saat boot dan dapat restart setelah kegagalan.
- Service `cloudflared` dengan konfigurasi tunnel yang aman dan otomatis berjalan saat boot.
- Aplikasi Go di loopback; jangan membuka port aplikasi ke jaringan publik secara langsung.
- Nginx hanya jika diperlukan; Cloudflare Tunnel dapat diarahkan langsung ke origin lokal aplikasi Go.
- Firewall yang hanya membuka port administrasi yang diperlukan, bukan port publik aplikasi/database.
- SSH untuk administrasi dengan autentikasi aman.
- Sinkronisasi waktu sistem.
- Backup terjadwal dan uji pemulihan.
- Dokumentasi pembaruan aplikasi dan rollback versi.

Contoh alur rilis:
1. Build dan uji aplikasi pada komputer pengembangan.
2. Jalankan test otomatis.
3. Salin binary dan aset yang dibutuhkan ke HG680P.
4. Pastikan volume data tersedia dan backup SQLite yang konsisten sudah dibuat.
5. Jalankan migration database dengan prosedur backup terlebih dahulu.
6. Restart service aplikasi Go dan `cloudflared`.
7. Periksa health endpoint lokal, log kedua service, akses domain, dan beberapa slug undangan.
8. Pastikan rollback aplikasi dan pemulihan database dapat dilakukan jika rilis gagal.

### 21.6 Akses publik — Cloudflare Tunnel

Pilihan akses publik MVP adalah **Cloudflare Tunnel menggunakan `cloudflared`**. Tunnel membuat koneksi keluar dari STB ke jaringan Cloudflare sehingga pada konfigurasi yang tepat tidak diperlukan port forwarding atau port aplikasi yang terbuka langsung di router. Tunnel tidak meng-host aplikasi dan tidak mencadangkan database; aplikasi Go dan SQLite tetap berjalan di STB.

Persyaratan konfigurasi:
- Domain `sukamoto.web.id` harus dikuasai dan dikelola melalui akun Cloudflare yang sesuai; pastikan nameserver dan status domain benar sebelum cutover.
- Buat tunnel dengan kredensial yang disimpan aman dan jalankan `cloudflared` sebagai service `systemd` non-root jika memungkinkan.
- Aplikasi Go hanya listen pada loopback, misalnya `127.0.0.1:8080`, kecuali ada alasan teknis yang terdokumentasi untuk bind lain.
- Konfigurasikan hostname publik untuk situs utama dan wildcard undangan, misalnya `invite.sukamoto.web.id` serta `*.invite.sukamoto.web.id`, dengan ingress yang mengarah ke origin lokal. Pastikan wildcard DNS/hostname dan aturan tunnel didukung oleh konfigurasi domain aktual.
- Aplikasi harus memvalidasi `Host`/hostname dan memetakan slug ke undangan yang benar; jangan mempercayai header hostname arbitrer dari request yang tidak tepercaya.
- Halaman undangan publik tetap dapat diakses tamu, tetapi dashboard admin dan endpoint sensitif harus memiliki autentikasi aplikasi yang kuat. Cloudflare Access dapat dipertimbangkan untuk rute admin/internal, tetapi jangan memasang kebijakan yang tanpa sengaja memblokir seluruh tamu yang membuka undangan.
- Gunakan HTTPS dari pengunjung ke Cloudflare dan konfigurasi koneksi Cloudflare ke origin dengan benar. Tunnel bukan pengganti autentikasi, otorisasi, validasi input, rate limiting, atau backup.
- Jangan mengekspos port SQLite, SSH, atau endpoint administrasi database ke internet. Batasi SSH untuk administrasi dan gunakan autentikasi aman.
- Pantau status `cloudflared`, log, koneksi internet, penggunaan bandwidth, serta pemulihan setelah reboot. Jika STB atau koneksi internet mati, undangan yang hanya di-host di STB akan ikut tidak tersedia.

Cloudflare R2 bersifat opsional untuk foto, video, musik, atau backup di luar perangkat jika anggaran dan kebutuhan mendukung. R2 bukan syarat agar Tunnel berfungsi. Sebelum produksi, uji akses situs utama dan beberapa slug undangan dari jaringan luar, HTTPS, route admin, serta perilaku saat tunnel terputus.

### 21.7 Batas kapasitas dan kesiapan produksi

HG680P diperlakukan sebagai **target MVP dan server bertrafik rendah sampai pengujian membuktikan kapasitasnya**. Jangan menjanjikan jumlah pelanggan atau pengunjung serentak tertentu sebelum load test.

Sebelum produksi, uji:
- Pemakaian RAM saat aplikasi dan database berjalan.
- Performa halaman undangan di bawah beban.
- Banyaknya request serentak yang dapat ditangani.
- Kapasitas penyimpanan dan pertumbuhan media.
- Pemulihan setelah listrik atau koneksi terputus.
- Restart otomatis setelah reboot.
- Backup dan restore database.
- Masa aktif, slug, pembayaran, dan publikasi.
- Ketersediaan domain dari jaringan luar melalui Cloudflare Tunnel.
- Perilaku aplikasi saat tunnel atau internet putus dan pemulihan otomatis setelah koneksi kembali.
- Waktu backup SQLite, konsistensi backup saat ada aktivitas, serta keberhasilan uji restore.

Untuk layanan komersial dengan kebutuhan uptime tinggi atau trafik yang berkembang, pertimbangkan memindahkan database, media, atau seluruh layanan publik ke server terpisah. HG680P dapat tetap digunakan sebagai server staging, development, atau admin/internal.

### 21.8 Informasi yang harus diverifikasi pada perangkat aktual

Sebelum memilih konfigurasi final, catat:
- OS dan versinya.
- Arsitektur CPU (`arm64` atau `armhf`).
- RAM yang terdeteksi.
- Ruang penyimpanan yang tersedia.
- Apakah sistem boot dari microSD, USB, atau eMMC.
- Jenis koneksi jaringan dan kestabilannya.
- Apakah koneksi internet memiliki IP publik atau berada di balik CGNAT.

Perintah diagnostik yang dapat dijalankan di Linux:

```bash
uname -m
cat /etc/os-release
free -h
df -h
nproc
ip addr
```

Jangan mengubah firmware atau menulis ke eMMC hanya untuk deployment aplikasi tanpa memastikan proses pemulihan dan backup perangkat tersedia.
