-- Migration 012: permission moderasi ulasan (Langkah 8 Modul 10)
INSERT INTO permissions (name, description) VALUES
('review:moderate', 'Menghapus ulasan milik siapa pun')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_name, permission_name) VALUES
('admin', 'review:moderate')
ON CONFLICT DO NOTHING;
