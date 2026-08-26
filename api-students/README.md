# Student REST API

REST API untuk mengelola data mahasiswa menggunakan Go dan Fiber.

## Base URL

http://localhost:3000/api/v1

## API Contract

| Method | Endpoint | Parameter | Contoh Body | Status | Contoh Response |
|---|---|---|---|---|---|
| GET | `/students` | `page`, `limit`, `search`, `sort`, `order`, `is_active`, `min_grade`, `max_grade` | - | 200, 400 | Daftar student + meta |
| GET | `/students/:id` | `id` | - | 200, 400, 404 | Satu data student |
| POST | `/students` | - | `nim`, `name`, `grade`, `is_active` | 201, 400, 409, 415, 422 | Student yang dibuat |
| PUT | `/students/:id` | `id` | `nim`, `name`, `grade`, `is_active` | 200, 400, 404, 409, 415, 422 | Student yang diperbarui |
| PATCH | `/students/:id` | `id` | Field yang ingin diubah | 200, 400, 404, 409, 415, 422 | Student yang diperbarui |
| DELETE | `/students/:id` | `id` | - | 204, 400, 404 | Tidak ada response body |

### GET /students

Mengambil daftar student. Mendukung pagination, search, sorting, dan filtering.

**Parameter:**
- `page` - nomor halaman
- `limit` - jumlah data per halaman
- `search` - pencarian berdasarkan NIM atau nama
- `sort` - field pengurutan
- `order` - `asc` atau `desc`
- `is_active` - filter status aktif
- `min_grade` - nilai minimum
- `max_grade` - nilai maksimum

**Contoh request:**

GET /api/v1/students?page=1&limit=2

**Response 200:**

{
  "success": true,
  "message": "student berhasil ditemukan",
  "data": [...],
  "meta": {
    "page": 1,
    "limit": 2,
    "total": 5,
    "total_pages": 3
  }
}

### GET /students/:id

Mengambil satu student berdasarkan ID.

**Parameter:**
- `id` - ID student

**Contoh request:**

GET /api/v1/students/1

**Response 200:**

{
  "success": true,
  "message": "student berhasil ditemukan",
  "data": {
    "id": 1,
    "nim": "20250001",
    "name": "Andi PATCH Final",
    "grade": 99,
    "is_active": true
  }
}

### POST /students

Membuat student baru.

**Request Body:**

{
  "nim": "20250008",
  "name": "Hendra Wijaya",
  "grade": 91,
  "is_active": true
}

**Response 201:**

{
  "success": true,
  "message": "student berhasil dibuat",
  "data": {
    "id": 7,
    "nim": "20250008",
    "name": "Hendra Wijaya",
    "grade": 91,
    "is_active": true
  }
}

### PUT /students/:id

Mengganti seluruh data student berdasarkan ID.

**Request Body:**

{
  "nim": "20250001",
  "name": "Andi PUT",
  "grade": 95,
  "is_active": true
}

**Response 200:**

{
  "success": true,
  "message": "student berhasil diperbarui",
  "data": {
    "id": 1,
    "nim": "20250001",
    "name": "Andi PUT",
    "grade": 95,
    "is_active": true
  }
}

### PATCH /students/:id

Memperbarui sebagian data student.

**Request Body:**

{
  "grade": 99
}

**Response 200:**

{
  "success": true,
  "message": "student berhasil diperbarui",
  "data": {
    "id": 1,
    "nim": "20250001",
    "name": "Andi PATCH Final",
    "grade": 99,
    "is_active": true
  }
}

### DELETE /students/:id

Menghapus student berdasarkan ID.

**Contoh request:**

DELETE /api/v1/students/5

**Response:**

204 No Content