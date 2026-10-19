package helper

import "sort"

// PermissionSet menyimpan pemetaan role ke daftar permission miliknya.
type PermissionSet struct {
	byRole map[string]map[string]struct{}
}

// NewPermissionSet mengubah hasil query menjadi bentuk yang cepat dicari.
// map bersarang dipilih karena pencarian "apakah role X punya permission Y"
// menjadi satu langkah, bukan perulangan sepanjang daftar.
func NewPermissionSet(raw map[string][]string) *PermissionSet {
	byRole := make(map[string]map[string]struct{}, len(raw))
	for role, permissions := range raw {
		set := make(map[string]struct{}, len(permissions))
		for _, permission := range permissions {
			set[permission] = struct{}{}
		}
		byRole[role] = set
	}
	return &PermissionSet{byRole: byRole}
}

// Can menjawab satu pertanyaan: apakah role ini memiliki permission itu.
//
// Perhatikan perilakunya ketika role atau permission tidak dikenal:
// jawabannya SELALU false. Prinsipnya disebut fail closed.
func (p *PermissionSet) Can(role, permission string) bool {
	if p == nil {
		return false
	}
	permissions, ok := p.byRole[role]
	if !ok {
		return false
	}
	_, granted := permissions[permission]
	return granted
}

// PermissionsOf mengembalikan seluruh permission milik sebuah role,
// terurut, untuk ditampilkan pada endpoint profil.
func (p *PermissionSet) PermissionsOf(role string) []string {
	result := []string{}
	if p == nil {
		return result
	}
	for permission := range p.byRole[role] {
		result = append(result, permission)
	}
	sort.Strings(result)
	return result
}

// KnownRoles mengembalikan daftar role yang dikenal sistem, terurut.
func (p *PermissionSet) KnownRoles() []string {
	result := []string{}
	if p == nil {
		return result
	}
	for role := range p.byRole {
		result = append(result, role)
	}
	sort.Strings(result)
	return result
}

// IsKnownRole dipakai saat memvalidasi permintaan pergantian role.
func (p *PermissionSet) IsKnownRole(role string) bool {
	if p == nil {
		return false
	}
	_, ok := p.byRole[role]
	return ok
}
