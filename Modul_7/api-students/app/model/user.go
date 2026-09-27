package model

import "time"

// User merepresentasikan akun pemakai di database.
// Field Password sengaja diberi tag json:"-" agar TIDAK PERNAH muncul di JSON response.
type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`          // hash bcrypt — tidak pernah dikirim ke client
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// UserProfile adalah subset User yang aman untuk dikirim ke client (tanpa password).
type UserProfile struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// AssignRoleRequest dipakai endpoint PATCH /users/:id/role.
type AssignRoleRequest struct {
	Role string `json:"role" validate:"required"`
}

// AuthUser merepresentasikan identitas pengguna yang terautentikasi dari JWT token.
type AuthUser struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// ErrorResponse adalah format response kegagalan terpusat (Langkah 1).
type ErrorResponse struct {
	Success   bool              `json:"success"`
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	Fields    map[string]string `json:"fields,omitempty"`
	RequestID string            `json:"request_id,omitempty"`
}

// CreateUserRequest untuk validasi deklaratif pembuatan user.
type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=30,alphanum"`
	Email    string `json:"email" validate:"required,email,max=120"`
	Password string `json:"password" validate:"required,min=8,max=72,nospace"`
}

// ReplaceUserRequest untuk validasi deklaratif PUT /users/:id.
type ReplaceUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=30,alphanum"`
	Email    string `json:"email" validate:"required,email,max=120"`
	IsActive bool   `json:"is_active"`
}

// PatchUserRequest untuk validasi deklaratif PATCH /users/:id.
// Catatan: Username menggunakan *string (pointer) agar mendukung omitnil dan perbandingan nil (Perbaikan Bug Compiler #2).
type PatchUserRequest struct {
	Username *string `json:"username,omitempty" validate:"omitnil,min=3,max=30,alphanum"`
	Email    *string `json:"email,omitempty" validate:"omitnil,email,max=120"`
	IsActive *bool   `json:"is_active,omitempty"`
}

// Cursor menyimpan penanda posisi baris terakhir pada keyset pagination.
type Cursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        int       `json:"id"`
}

// CursorQuery menampung parameter query cursor pagination untuk users.
type CursorQuery struct {
	Search   string
	IsActive *bool
	After    *Cursor
	Limit    int
}

// CursorMeta menggantikan Meta pada endpoint yang memakai cursor (Langkah 7).
type CursorMeta struct {
	Limit      int    `json:"limit"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}
