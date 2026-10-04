# Dokumentasi API - SIAKAD Mini

RESTful API backend untuk layanan akademik sederhana yang mengelola data mahasiswa, mata kuliah, dan Kartu Rencana Studi (KRS).

- **Base URL:** `http://localhost:3000/api/v1`
- **Format Data:** JSON
- **Autentikasi:** Bearer Token (JWT). Semua endpoint **kecuali Login** mewajibkan header: 
  `Authorization: Bearer <token_anda>`

---

## 1. Autentikasi

### 1.1. Login
- **Endpoint:** `POST /auth/login`
- **Akses:** Publik
- **Fungsi:** Mendapatkan access token.
- **Request Body:**
  ```json
  {
    "email": "admin@siakad.ac.id",
    "password": "rahasia123"
  }
  ```
- **Response Sukses (200 OK):**
  ```json
  {
    "success": true,
    "message": "Login berhasil",
    "data": {
      "access_token": "eyJhbGciOiJIUzI1NiIs...",
      "token_type": "Bearer",
      "expires_in": 900,
      "user": {
        "id": 1,
        "email": "admin@siakad.ac.id",
        "role": "admin"
      }
    }
  }
  ```

### 1.2. Profil (Me)
- **Endpoint:** `GET /auth/me`
- **Akses:** Semua Role (Admin & Mahasiswa)
- **Fungsi:** Mengambil data profil user yang sedang login.
- **Response Sukses (200 OK):**
  *(Jika Mahasiswa, response memuat tambahan data `students`)*
  ```json
  {
    "success": true,
    "message": "Data profil berhasil diambil",
    "data": {
      "id": 2,
      "email": "student1@siakad.ac.id",
      "role": "mahasiswa",
      "student": {
        "nim": "187221000001",
        "nama": "Mahasiswa 1",
        "prodi": "Informatika",
        "angkatan": 2026,
        "ipk_terakhir": 3.45
      }
    }
  }
  ```

---

## 2. Manajemen Mahasiswa

### 2.1. Daftar Mahasiswa
- **Endpoint:** `GET /students`
- **Akses:** Admin
- **Fungsi:** Melihat daftar mahasiswa dengan *pagination* dan *filter*.
- **Query Parameters:**
  - `page` (opsional, default: 1)
  - `per_page` (opsional, default: 10, max: 50)
  - `search` (opsional, mencari NIM atau Nama)
  - `prodi` (opsional, filter program studi)
  - `angkatan` (opsional, filter tahun angkatan)
  - `sort` (opsional, `nama` atau `-ipk_terakhir`)
- **Response Sukses (200 OK):**
  ```json
  {
    "success": true,
    "message": "Daftar mahasiswa berhasil diambil",
    "data": [
      {
        "id": 1,
        "nim": "187221000001",
        "nama": "Mahasiswa 1",
        "prodi": "Informatika",
        "angkatan": 2026,
        "ipk_terakhir": 3.45
      }
    ],
    "meta": {
      "current_page": 1,
      "per_page": 10,
      "total": 20,
      "last_page": 2
    }
  }
  ```

### 2.2. Tambah Mahasiswa Baru
- **Endpoint:** `POST /students`
- **Akses:** Admin
- **Fungsi:** Mendaftarkan mahasiswa (otomatis membuatkan akun login). Password default adalah NIM.
- **Request Body:**
  ```json
  {
    "nim": "187221000099",
    "nama": "Budi Santoso",
    "email": "budi@siakad.ac.id",
    "prodi": "Sistem Informasi",
    "angkatan": 2026,
    "ipk_terakhir": 3.80
  }
  ```
- **Response Sukses (201 Created):** Mengembalikan data mahasiswa yang baru dibuat.

### 2.3. Detail Mahasiswa
- **Endpoint:** `GET /students/{id}`
- **Akses:** Admin & Mahasiswa (Hanya akunnya sendiri)
- **Fungsi:** Mengambil data diri beserta rincian mata kuliah yang diambil (SKS).
- **Response Sukses (200 OK):**
  ```json
  {
    "success": true,
    "message": "Detail mahasiswa berhasil diambil",
    "data": {
      "id": 1,
      "nim": "187221000001",
      "nama": "Mahasiswa 1",
      "prodi": "Informatika",
      "angkatan": 2026,
      "ipk_terakhir": 3.45,
      "total_sks": 6,
      "batas_sks": 24,
      "enrollments": [
        {
          "enrollment_id": 1,
          "kode_mk": "IF101",
          "nama_mk": "Advanced Backend Programming",
          "sks": 3,
          "tahun_akademik": "2026/2027-Ganjil"
        }
      ]
    }
  }
  ```

### 2.4. Update Mahasiswa
- **Endpoint:** `PUT /students/{id}`
- **Akses:** Admin
- **Fungsi:** Memperbarui data mahasiswa. NIM tidak bisa diubah.
- **Request Body:**
  ```json
  {
    "nama": "Budi Santoso",
    "prodi": "Sistem Informasi",
    "angkatan": 2026,
    "ipk_terakhir": 3.90
  }
  ```
- **Response Sukses (200 OK):** Mengembalikan data terbaru.

### 2.5. Hapus Mahasiswa (Soft Delete)
- **Endpoint:** `DELETE /students/{id}`
- **Akses:** Admin
- **Fungsi:** Menghapus data mahasiswa secara *Soft Delete* (mengisi kolom `deleted_at`).
- **Response Sukses (204 No Content):** (Kosong).

---

## 3. Manajemen Mata Kuliah

### 3.1. Daftar Mata Kuliah
- **Endpoint:** `GET /courses`
- **Akses:** Semua Role (Admin & Mahasiswa)
- **Fungsi:** Melihat ketersediaan mata kuliah.
- **Query Parameters:**
  - `semester` (opsional)
  - `search` (opsional, kode atau nama mata kuliah)
  - `available` (opsional, `true` jika hanya ingin melihat yang kuotanya belum penuh)
- **Response Sukses (200 OK):**
  ```json
  {
    "success": true,
    "message": "Daftar mata kuliah berhasil diambil",
    "data": [
      {
        "id": 1,
        "kode_mk": "IF101",
        "nama_mk": "Advanced Backend Programming",
        "sks": 3,
        "semester": 1,
        "kuota": 40,
        "terisi": 15,
        "sisa_kuota": 25
      }
    ]
  }
  ```

---

## 4. Kartu Rencana Studi (KRS)

### 4.1. Ambil Mata Kuliah
- **Endpoint:** `POST /enrollments`
- **Akses:** Mahasiswa
- **Fungsi:** Mendaftarkan kelas. Akan memvalidasi IPK (Batas SKS) dan sisa kuota mata kuliah secara ketat.
- **Request Body:**
  ```json
  {
    "course_id": 1,
    "tahun_akademik": "2026/2027-Ganjil"
  }
  ```
- **Response Sukses (201 Created):**
  ```json
  {
    "success": true,
    "message": "Berhasil mengambil mata kuliah"
  }
  ```
- **Response Error (422 Unprocessable Entity):**
  Jika sisa SKS tidak mencukupi, sistem akan mengembalikan error berisi sisa SKS yang diizinkan.

### 4.2. Batalkan Mata Kuliah
- **Endpoint:** `DELETE /enrollments/{id}`
- **Akses:** Mahasiswa (Hanya KRS miliknya sendiri)
- **Fungsi:** Membatalkan pengisian KRS. Kuota mata kuliah akan kembali bertambah.
- **Response Sukses (204 No Content):** (Kosong).

---

## 🚫 Standar Error Response
Jika terjadi kesalahan validasi atau bisnis (*Conflict/Unprocessable Entity*), sistem mengembalikan kode HTTP `422` atau `409` dengan struktur:
```json
{
  "success": false,
  "message": "Validasi gagal",
  "errors": {
    "nim": ["NIM harus 12 digit numerik"],
    "course_id": ["Mata kuliah ini sudah Anda ambil"]
  }
}
```
