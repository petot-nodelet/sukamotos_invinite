# SUKAMOTO WEDDING INVITATION — API CONTRACT v1.0

Status: Draft implementasi
Frontend: HTML, CSS, JavaScript
Backend: Golang
Database: SQLite
Protocol: HTTP/HTTPS + JSON
API prefix: `/api/v1`

## 1. Aturan integrasi

1. Dokumen ini menjadi sumber kebenaran bagi frontend dan backend.
2. Frontend tidak boleh mengarang endpoint atau field baru tanpa memperbarui kontrak.
3. Backend tidak boleh mengubah bentuk respons tanpa memperbarui kontrak.
4. Template asli tetap berupa HTML, CSS, JavaScript dan aset statis.
5. Tidak ada login pelanggan. Autentikasi hanya untuk admin.
6. Harga dan status pembayaran selalu ditentukan backend.
7. Data undangan yang belum dipublikasikan tidak boleh diakses publik.
8. Seluruh waktu API menggunakan ISO 8601 UTC, misalnya `2026-10-09T10:00:00Z`.
9. Semua respons API menggunakan JSON kecuali endpoint HTML, file, atau unduhan yang dinyatakan secara eksplisit.
10. Versi API tetap `/api/v1` selama perubahan kompatibel.

## 2. Format respons umum

Respons berhasil:

```
{
  "success": true,
  "data": {},
  "meta": {}
}
```

`meta` bersifat opsional dan digunakan untuk pagination atau informasi tambahan.

Respons gagal:

```
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Periksa kembali data yang dikirim.",
    "fields": {
      "bride_name": "Nama pasangan wajib diisi."
    }
  }
}
```

`fields` bersifat opsional. Pesan error harus aman ditampilkan ke pengguna dan tidak boleh memuat stack trace, SQL, secret, atau detail internal server.

Status HTTP:

- `200 OK`: permintaan berhasil.
- `201 Created`: resource berhasil dibuat.
- `204 No Content`: berhasil tanpa body.
- `400 Bad Request`: JSON atau parameter tidak valid.
- `401 Unauthorized`: belum terautentikasi.
- `403 Forbidden`: tidak memiliki izin.
- `404 Not Found`: resource tidak ditemukan.
- `409 Conflict`: konflik status atau duplikasi.
- `413 Content Too Large`: unggahan terlalu besar.
- `415 Unsupported Media Type`: format tidak didukung.
- `422 Unprocessable Content`: validasi bisnis gagal.
- `429 Too Many Requests`: batas permintaan tercapai.
- `500 Internal Server Error`: kegagalan internal yang tidak terduga.

## 3. Endpoint publik

### GET /health

Memeriksa apakah proses server aktif.

Contoh:

```
{
  "success": true,
  "data": {
    "status": "ok"
  }
}
```

Jangan mengekspos secret, kredensial, atau informasi infrastruktur sensitif melalui health check publik.

### GET /api/v1/packages

Mengambil daftar paket yang aktif.

Contoh data:

```
{
  "success": true,
  "data": [
    {
      "slug": "silver",
      "name": "Silver",
      "price": 0,
      "currency": "IDR",
      "features": [],
      "active": true
    }
  ]
}
```

Angka harga di atas adalah contoh format, bukan harga bisnis. Backend wajib mengisi harga yang sudah dikonfigurasi. Paket yang tidak aktif tidak boleh dapat dipesan.

### GET /api/v1/templates?package=silver

Mengambil katalog template yang tersedia.

Bentuk data:

```
{
  "success": true,
  "data": [
    {
      "slug": "contoh-template",
      "name": "Nama Template",
      "package_slug": "silver",
      "thumbnail_url": "/assets/...",
      "demo_url": "/demo/...",
      "available": true
    }
  ]
}
```

`thumbnail_url` dan `demo_url` hanya boleh berisi path yang valid dan berasal dari katalog template yang dikelola server.

Jangan mengarang nama template, folder, atau URL sebelum repository template diaudit.

Catalog registration currently covers all six original files under `frontend/templates/`:

| Package | File |
| --- | --- |
| SILVER | `SILVER/template1_classic_matcha.html` |
| SILVER | `SILVER/template4-burgundy-minimalist.html` |
| SILVER | `SILVER/template8_islamic_traditional.html` |
| GOLD | `GOLD/dusty_rose_wedding.html` |
| GOLD | `GOLD/black-champagne-gold-ivory.html` |
| PLATINUM VIP | `PLATINUM/3_Imperial-Modern.html` |

`demo_url` points to the registered source file served by the backend under `/demo/`. This is a static preview route, not a versioned API endpoint. A client-side filled preview may load that page in an iframe and send invitation fields using origin-checked `postMessage`; those fields must not be placed in the URL or persistent browser storage.

### GET /api/v1/templates/{slug}

Mengambil detail template, fitur yang didukung, dan field formulir yang diperlukan.

Contoh:

```
{
  "success": true,
  "data": {
    "slug": "contoh-template",
    "name": "Nama Template",
    "package_slug": "silver",
    "supported_fields": [
      "bride_name",
      "groom_name",
      "event_date"
    ]
  }
}
```

Field di atas hanya ilustrasi. Daftar sebenarnya harus berasal dari hasil audit HTML dan kebutuhan template.

### POST /api/v1/orders

Membuat pesanan tanpa akun pelanggan.

Request:

```
{
  "package_slug": "silver",
  "template_slug": "contoh-template",
  "customer": {
    "name": "Nama Pemesan",
    "whatsapp": "6281234567890"
  },
  "invitation": {
    "bride_name": "Nama Mempelai Wanita",
    "groom_name": "Nama Mempelai Pria",
    "events": []
  },
  "idempotency_key": "random-unique-request-key"
}
```

Semua nilai di atas adalah contoh format.

Aturan:

- Backend memvalidasi paket dan template.
- Harga dihitung ulang dari database, bukan dari harga kiriman frontend.
- Data pasangan divalidasi sesuai field yang didukung template.
- Permintaan duplikat dengan idempotency key yang sama tidak boleh membuat pesanan ganda.
- Backend menghasilkan ID internal dan referensi pesanan acak yang aman.
- Jangan menggunakan ID berurutan sebagai rahasia akses.
- Data pribadi tidak boleh dikembalikan secara berlebihan.
- Status awal: `pending_payment`.

Respons `201 Created`:

```
{
  "success": true,
  "data": {
    "order_reference": "random-public-reference",
    "status": "pending_payment",
    "total_amount": 0,
    "currency": "IDR",
    "payment_instructions": {
      "method": "bank_transfer",
      "description": "Instruksi pembayaran dari konfigurasi server."
    }
  }
}
```

Nilai `total_amount` di atas adalah placeholder format. Implementasi wajib memakai harga yang sah dari database.

### GET /api/v1/orders/{reference}

Melihat status pesanan tanpa akun.

Respons hanya berisi informasi minimum yang diperlukan:

```
{
  "success": true,
  "data": {
    "order_reference": "random-public-reference",
    "status": "pending_payment",
    "payment_status": "unpaid",
    "published": false,
    "invitation_url": null
  }
}
```

Referensi harus memiliki entropi tinggi dan tidak dapat ditebak. Tambahkan rate limiting. Jika memungkinkan, gunakan token akses terpisah atau token akses acak yang disimpan dalam bentuk hash. Jangan mengembalikan alamat, foto, detail acara, bukti pembayaran, atau data pribadi melalui endpoint status ini.

### POST /api/v1/orders/{reference}/payment

Mengirim bukti pembayaran.

Gunakan `multipart/form-data` dengan field:

- `method`: metode pembayaran yang dikonfigurasi server.
- `proof`: file bukti pembayaran.
- `idempotency_key`: opsional untuk mencegah pengiriman duplikat.

Aturan:

- Batasi ukuran file dan tipe file yang diperbolehkan.
- Periksa isi file, bukan hanya ekstensi atau MIME dari browser.
- Buat nama file internal acak.
- Simpan file di lokasi privat.
- Jangan menyimpan bukti pembayaran pada folder publik atau membundelnya dalam binary Go.
- Admin saja yang dapat mengakses bukti pembayaran.
- Pengiriman bukti tidak sama dengan pembayaran terverifikasi.

Status berubah menjadi `payment_submitted` setelah validasi dan penyimpanan berhasil.

### GET /api/v1/invitations/{slug}

Mengembalikan data undangan publik yang sudah diterbitkan.

Contoh:

```
{
  "success": true,
  "data": {
    "slug": "pasangan-abc",
    "template_slug": "contoh-template",
    "published": true,
    "invitation": {}
  }
}
```

Hanya data yang memang dimaksudkan untuk dipublikasikan boleh dikembalikan.

### GET /i/{slug}

Merender HTML undangan berdasarkan template asli dan data undangan.

Ketentuan:

- Tidak ada undangan draft yang dapat diakses.
- Slug tidak boleh menjadi akses ke data pesanan atau pembayaran.
- Data pengguna harus di-escape dengan benar.
- Jangan memasukkan HTML mentah yang berasal dari input pelanggan.
- Aset CSS, JavaScript, gambar, font, dan audio harus dimuat menggunakan path yang benar.
- Halaman harus menggunakan template yang terdaftar dan valid.
- Render the registered original HTML rather than a generic placeholder. Supply published invitation data as safely encoded JSON and let the template's shared runtime update text, images, date, venue, and countdown without evaluating user-provided markup.

## 4. Endpoint admin

Semua endpoint di bawah memerlukan sesi admin yang valid dan pemeriksaan otorisasi di server.

### POST /api/v1/admin/auth/login

Request:

```
{
  "username": "admin",
  "password": "..."
}
```

Keamanan:

- Gunakan password hashing yang sesuai.
- Gunakan cookie sesi `HttpOnly`, `Secure` di HTTPS, dan `SameSite` yang sesuai.
- Regenerasi ID sesi setelah login.
- Terapkan rate limiting.
- Jangan mengembalikan password atau hash password.
- Jangan menyimpan token sesi dalam localStorage.
- Jangan membuat endpoint pendaftaran admin publik.

### POST /api/v1/admin/auth/logout

Mengakhiri sesi dan menghapus cookie sesi melalui mekanisme yang benar.

### GET /api/v1/admin/orders?status=pending_payment&page=1&limit=20

Mengambil daftar pesanan untuk admin.

Mendukung pagination dan filter status. Batasi nilai `limit` maksimum.

### GET /api/v1/admin/orders/{id}

Mengambil detail pesanan dan informasi pembayaran untuk admin yang berwenang.

### POST /api/v1/admin/orders/{id}/verify-payment

Request:

```
{
  "decision": "approve",
  "note": "Pembayaran telah diperiksa."
}
```

Nilai `decision`:

- `approve`
- `reject`

Ketentuan:

- Hanya admin yang berwenang dapat memverifikasi.
- Verifikasi tidak boleh dilakukan hanya berdasarkan upload.
- Perubahan status harus mengikuti transisi yang sah.
- Catat identitas admin, waktu, keputusan, dan alasan dalam audit log.
- Operasi harus transaksional.

### POST /api/v1/admin/orders/{id}/publish

Menerbitkan undangan setelah semua syarat bisnis terpenuhi.

Request:

```
{
  "slug": "pasangan-abc"
}
```

Ketentuan:

- Pembayaran harus terverifikasi.
- Data undangan harus valid.
- Template harus tersedia.
- Slug harus divalidasi dan unik.
- Operasi harus aman dari publikasi ganda.
- Undangan yang belum diterbitkan tetap privat.

### POST /api/v1/admin/orders/{id}/reject

Menolak pesanan dengan alasan yang dicatat. Penolakan pesanan dan penolakan pembayaran harus dibedakan di model status jika memiliki konsekuensi berbeda.

### GET /api/v1/admin/templates

Mengambil daftar template beserta metadata internal untuk admin.

### PATCH /api/v1/admin/templates/{id}

Memperbarui metadata atau status ketersediaan template. Jangan mengizinkan perubahan arbitrary filesystem path dari input pengguna.

## 5. Status pesanan dan pembayaran

Status pesanan:

- `pending_payment`
- `payment_submitted`
- `payment_verified`
- `in_progress`
- `published`
- `rejected`
- `cancelled`

Status pembayaran:

- `unpaid`
- `submitted`
- `verified`
- `rejected`

Aturan:

- `pending_payment` berarti pesanan tercatat dan menunggu pembayaran.
- `payment_submitted` berarti bukti pembayaran telah dikirim.
- `payment_verified` berarti pembayaran telah diverifikasi admin.
- `in_progress` berarti pesanan sedang diproses.
- `published` berarti undangan sudah dapat diakses publik.
- `rejected` berarti pesanan atau pembayaran ditolak sesuai catatan keputusan.
- `cancelled` berarti pesanan dibatalkan menurut aturan bisnis.

Jangan menganggap `payment_verified` otomatis sama dengan `published`. Publikasi adalah tindakan terpisah.

## 6. Model data minimum

Package:

- id
- slug
- name
- price
- currency
- features_json
- active
- created_at
- updated_at

Template:

- id
- slug
- name
- package_slug atau relasi paket yang sesuai
- source_path internal
- thumbnail_path
- demo_path
- metadata_json
- available
- created_at
- updated_at

Order:

- id
- reference_hash atau mekanisme referensi aman
- idempotency_key
- package_id
- template_id
- customer_name
- customer_whatsapp
- invitation_data_json
- total_amount
- currency
- status
- created_at
- updated_at

Payment:

- id
- order_id
- method
- proof_storage_key
- status
- verified_by
- verified_at
- rejection_reason
- created_at
- updated_at

Invitation:

- id
- order_id
- slug
- template_id
- published
- published_at
- created_at
- updated_at

Admin:

- id
- username
- password_hash
- active
- created_at
- updated_at

AuditLog:

- id
- admin_id
- action
- entity_type
- entity_id
- details_json
- created_at

Sesuaikan model setelah audit kebutuhan nyata. Terapkan foreign key, indeks, constraint unik, migrasi database, dan transaksi untuk operasi yang membutuhkan konsistensi.

## 7. Kontrak error

Gunakan kode error yang konsisten, misalnya:

- `VALIDATION_ERROR`
- `PACKAGE_NOT_AVAILABLE`
- `TEMPLATE_NOT_AVAILABLE`
- `ORDER_NOT_FOUND`
- `INVALID_ORDER_REFERENCE`
- `INVALID_STATUS_TRANSITION`
- `PAYMENT_PROOF_INVALID`
- `PAYMENT_ALREADY_VERIFIED`
- `SLUG_ALREADY_EXISTS`
- `UNAUTHORIZED`
- `FORBIDDEN`
- `RATE_LIMITED`
- `INTERNAL_ERROR`

Pesan error tidak boleh mengungkap informasi internal.

## 8. CORS dan CSRF

Jika frontend dan backend disajikan dari origin yang sama, gunakan same-origin dan hindari CORS yang tidak diperlukan.

Jika origin berbeda saat pengembangan, izinkan hanya origin pengembangan yang dikonfigurasi secara eksplisit.

Untuk endpoint admin berbasis cookie, terapkan perlindungan CSRF. Jangan mengandalkan CORS sebagai satu-satunya perlindungan CSRF.

Endpoint publik juga harus memiliki batas ukuran request, validasi, dan rate limiting yang sesuai.

## 9. Template dan file statis

Template SILVER, GOLD, dan PLATINUM dibundel ke aplikasi Go setelah audit repository.

Gunakan `go:embed` untuk aset yang cocok dibundel. Gunakan katalog internal yang memetakan slug template ke template dan aset valid.

Jangan membiarkan request menentukan path file arbitrer.

Gunakan strategi rendering yang kompatibel dengan HTML dan JavaScript yang sudah ada. Hindari penggantian string global dan pastikan data pengguna di-escape.

File unggahan pelanggan tidak otomatis menjadi bagian dari embedded filesystem. File tersebut harus disimpan di storage runtime yang privat dan persisten.

## 10. Definisi selesai

Kontrak dianggap terimplementasi ketika:

1. Endpoint sesuai method dan path.
2. Bentuk request dan respons sesuai dokumen.
3. Validasi backend berjalan.
4. Status pesanan dan pembayaran konsisten.
5. Endpoint admin terlindungi.
6. Data pesanan tidak bocor ke pengguna lain.
7. Undangan draft tidak dapat diakses publik.
8. Template asli dapat dirender tanpa merusak aset.
9. Migrasi SQLite dan pengujian berhasil.
10. Perubahan kontrak didokumentasikan dan diuji.

END OF API CONTRACT v1.0
