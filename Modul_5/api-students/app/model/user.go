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
	CreatedAt time.Time `json:"created_at"`
}

// UserProfile adalah subset User yang aman untuk dikirim ke client (tanpa password).
type UserProfile struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}
