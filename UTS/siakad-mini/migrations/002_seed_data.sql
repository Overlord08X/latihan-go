-- Migrasi 002: Seeder Data Awal (Admin, 20 Mahasiswa, 10 Mata Kuliah, Sample KRS)

-- 1. Seed User Admin
INSERT INTO users (id, email, password, role) VALUES (1, 'admin@siakad.ac.id', '$2b$10$H62YsUdG8XtkFxJQU8vS/O2W/WC8YCh0BfLt06jCpyBoi54PfIz/u', 'admin') ON CONFLICT (email) DO NOTHING;

-- 2. Seed 20 Users Mahasiswa & Data Students
INSERT INTO users (id, email, password, role) VALUES (2, 'mhs01@siakad.ac.id', '$2b$10$wZGuAIoC7EnR1sZbXjusYekTm.j..HG3GxyrJeIYrybO9LlAVuQ5e', 'mahasiswa') ON CONFLICT (email) DO NOTHING;
INSERT INTO students (id, user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES (1, 2, '434241096001', 'Raihan Zulfa Kamal', 'D4 Teknik Informatika', 2024, 3.85) ON CONFLICT (nim) DO NOTHING;
INSERT INTO users (id, email, password, role) VALUES (3, 'mhs02@siakad.ac.id', '$2b$10$U4c0YBLbm1p2m0bTxIU3Te/9KaCUIidOqS9dpm8emVkT1prQaVXy2', 'mahasiswa') ON CONFLICT (email) DO NOTHING;
INSERT INTO students (id, user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES (2, 3, '434241096002', 'Ahmad Fauzi', 'D4 Teknik Informatika', 2024, 3.5) ON CONFLICT (nim) DO NOTHING;
INSERT INTO users (id, email, password, role) VALUES (4, 'mhs03@siakad.ac.id', '$2b$10$zSXgDIn9yecDxpCzdd9QiOpMEZUYILXuuUcxQEbeguTNgINldnG/6', 'mahasiswa') ON CONFLICT (email) DO NOTHING;
INSERT INTO students (id, user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES (3, 4, '434241096003', 'Budi Santoso', 'D4 Teknik Informatika', 2023, 2.8) ON CONFLICT (nim) DO NOTHING;
INSERT INTO users (id, email, password, role) VALUES (5, 'mhs04@siakad.ac.id', '$2b$10$VbD2aCAofywmGGin4wW6Q.rEIXd5E8lATlSy0u5/sBJ9gB4c/EOtW', 'mahasiswa') ON CONFLICT (email) DO NOTHING;
INSERT INTO students (id, user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES (4, 5, '434241096004', 'Citra Dewi', 'Sistem Informasi', 2024, 2.3) ON CONFLICT (nim) DO NOTHING;
INSERT INTO users (id, email, password, role) VALUES (6, 'mhs05@siakad.ac.id', '$2b$10$bdIsZvSEJ6wboB.Dvcvyv.glzA8epPMm/hzwc3hs4mQdV2NzgK3XK', 'mahasiswa') ON CONFLICT (email) DO NOTHING;
INSERT INTO students (id, user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES (5, 6, '434241096005', 'Dimas Pratama', 'Sistem Informasi', 2023, 3.75) ON CONFLICT (nim) DO NOTHING;
INSERT INTO users (id, email, password, role) VALUES (7, 'mhs06@siakad.ac.id', '$2b$10$l.cQXEX6nW5Lq4p2C7Tlh.NMWSZISzfNofK2ubwxKC1ErQNHB6F3i', 'mahasiswa') ON CONFLICT (email) DO NOTHING;
INSERT INTO students (id, user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES (6, 7, '434241096006', 'Eka Novitasari', 'Sistem Informasi', 2022, 2.9) ON CONFLICT (nim) DO NOTHING;
INSERT INTO users (id, email, password, role) VALUES (8, 'mhs07@siakad.ac.id', '$2b$10$1gtf1QGQ7ElS/lAmtnHBie1EN05xj5.AafowA4n./HpEDf9IAVS.K', 'mahasiswa') ON CONFLICT (email) DO NOTHING;
INSERT INTO students (id, user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES (7, 8, '434241096007', 'Fajar Hidayat', 'D4 Teknik Informatika', 2024, 2.45) ON CONFLICT (nim) DO NOTHING;
INSERT INTO users (id, email, password, role) VALUES (9, 'mhs08@siakad.ac.id', '$2b$10$mkZk0P7zNywXmW/UR55NTe42ze453b/2DzBoh38qZTFnMIlQtF56a', 'mahasiswa') ON CONFLICT (email) DO NOTHING;
INSERT INTO students (id, user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES (8, 9, '434241096008', 'Gita Permata', 'D4 Teknik Informatika', 2023, 3.9) ON CONFLICT (nim) DO NOTHING;
INSERT INTO users (id, email, password, role) VALUES (10, 'mhs09@siakad.ac.id', '$2b$10$9BIMSFRHJSohZErUebgVneFXET6trl8x2j44Un9brwC.29uKfoVXi', 'mahasiswa') ON CONFLICT (email) DO NOTHING;
INSERT INTO students (id, user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES (9, 10, '434241096009', 'Hadi Kusuma', 'Teknologi Informasi', 2022, 2.65) ON CONFLICT (nim) DO NOTHING;
INSERT INTO users (id, email, password, role) VALUES (11, 'mhs10@siakad.ac.id', '$2b$10$ueonJ.dTIvhUzNNPk2JZj.YJdW77JOOVl7TkHKFts6ixvOtjzsnve', 'mahasiswa') ON CONFLICT (email) DO NOTHING;
INSERT INTO students (id, user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES (10, 11, '434241096010', 'Indah Lestari', 'Teknologi Informasi', 2024, 3.1) ON CONFLICT (nim) DO NOTHING;
INSERT INTO users (id, email, password, role) VALUES (12, 'mhs11@siakad.ac.id', '$2b$10$Eg5l1rZKHM3Nyxl8bInkKOdAsLwcQPfCAZTYg/vklRaW.gZbrzoLq', 'mahasiswa') ON CONFLICT (email) DO NOTHING;
INSERT INTO students (id, user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES (11, 12, '434241096011', 'Joko Widodo', 'D4 Teknik Informatika', 2023, 3.2) ON CONFLICT (nim) DO NOTHING;
INSERT INTO users (id, email, password, role) VALUES (13, 'mhs12@siakad.ac.id', '$2b$10$S3sO96f.XGqfivc1EdD3n.4mmQoCZ.8D77pLVIYgBz4fuHkH21kLy', 'mahasiswa') ON CONFLICT (email) DO NOTHING;
INSERT INTO students (id, user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES (12, 13, '434241096012', 'Kartika Sari', 'Sistem Informasi', 2024, 2.7) ON CONFLICT (nim) DO NOTHING;
INSERT INTO users (id, email, password, role) VALUES (14, 'mhs13@siakad.ac.id', '$2b$10$WZrrDhRbI5.9iQhcqjth..Rhy4OMk9OPdB489YIZ4T0QkhsREGSMO', 'mahasiswa') ON CONFLICT (email) DO NOTHING;
INSERT INTO students (id, user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES (13, 14, '434241096013', 'Lukman Hakim', 'D4 Teknik Informatika', 2022, 3.6) ON CONFLICT (nim) DO NOTHING;
INSERT INTO users (id, email, password, role) VALUES (15, 'mhs14@siakad.ac.id', '$2b$10$jPE6xpPX36msW2fIKxju4OXvmJyx5cXf9Z1n9Bywwh5AzxgPig2yK', 'mahasiswa') ON CONFLICT (email) DO NOTHING;
INSERT INTO students (id, user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES (14, 15, '434241096014', 'Maya Anggraini', 'Teknologi Informasi', 2023, 2.2) ON CONFLICT (nim) DO NOTHING;
INSERT INTO users (id, email, password, role) VALUES (16, 'mhs15@siakad.ac.id', '$2b$10$VjpQ7ebCP2d/Bq1WRtxyMOKxEiMIlJ4LuTrre6IUgrAJmvgspZkeK', 'mahasiswa') ON CONFLICT (email) DO NOTHING;
INSERT INTO students (id, user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES (15, 16, '434241096015', 'Nanda Pratama', 'D4 Teknik Informatika', 2024, 3.4) ON CONFLICT (nim) DO NOTHING;
INSERT INTO users (id, email, password, role) VALUES (17, 'mhs16@siakad.ac.id', '$2b$10$6pzZkGd08NkZGTaeE6yZke.wtfwrVt23n0KboFibUWEgnCXxN372m', 'mahasiswa') ON CONFLICT (email) DO NOTHING;
INSERT INTO students (id, user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES (16, 17, '434241096016', 'Olivia Putri', 'Sistem Informasi', 2023, 3.8) ON CONFLICT (nim) DO NOTHING;
INSERT INTO users (id, email, password, role) VALUES (18, 'mhs17@siakad.ac.id', '$2b$10$OnJFeJiHDg.OUx74a6T/Mer9MlSbARLYqbFq9st9104d.qfcxGama', 'mahasiswa') ON CONFLICT (email) DO NOTHING;
INSERT INTO students (id, user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES (17, 18, '434241096017', 'Panji Asmoro', 'Teknologi Informasi', 2022, 2.55) ON CONFLICT (nim) DO NOTHING;
INSERT INTO users (id, email, password, role) VALUES (19, 'mhs18@siakad.ac.id', '$2b$10$KioprueOi8lV2HaZFEDpduSuHVCWz3dQbvxUVco9ENpAu2YOee3Ra', 'mahasiswa') ON CONFLICT (email) DO NOTHING;
INSERT INTO students (id, user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES (18, 19, '434241096018', 'Qori Amalia', 'D4 Teknik Informatika', 2024, 3.95) ON CONFLICT (nim) DO NOTHING;
INSERT INTO users (id, email, password, role) VALUES (20, 'mhs19@siakad.ac.id', '$2b$10$1zO4e0x2VoEMFlnciqvMbOJU4vDVFueLzTnD9mPEMNtia5ZltICka', 'mahasiswa') ON CONFLICT (email) DO NOTHING;
INSERT INTO students (id, user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES (19, 20, '434241096019', 'Rizky Ramadhan', 'Sistem Informasi', 2023, 2.4) ON CONFLICT (nim) DO NOTHING;
INSERT INTO users (id, email, password, role) VALUES (21, 'mhs20@siakad.ac.id', '$2b$10$6v6W341vVZRD87K0FHswAe7X74bEzmWtvNDXd0kDeRuPICIidwZ7y', 'mahasiswa') ON CONFLICT (email) DO NOTHING;
INSERT INTO students (id, user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES (20, 21, '434241096020', 'Siti Nurhaliza', 'D4 Teknik Informatika', 2022, 3.15) ON CONFLICT (nim) DO NOTHING;

-- 3. Seed 10 Mata Kuliah
INSERT INTO courses (id, kode_mk, nama_mk, sks, semester, kuota) VALUES (1, 'MK001', 'Pemrograman Backend Lanjut', 3, 5, 30) ON CONFLICT (kode_mk) DO NOTHING;
INSERT INTO courses (id, kode_mk, nama_mk, sks, semester, kuota) VALUES (2, 'MK002', 'Basis Data Lanjut', 3, 5, 25) ON CONFLICT (kode_mk) DO NOTHING;
INSERT INTO courses (id, kode_mk, nama_mk, sks, semester, kuota) VALUES (3, 'MK003', 'Arsitektur Perangkat Lunak', 3, 5, 30) ON CONFLICT (kode_mk) DO NOTHING;
INSERT INTO courses (id, kode_mk, nama_mk, sks, semester, kuota) VALUES (4, 'MK004', 'Cloud Computing & DevOps', 3, 5, 20) ON CONFLICT (kode_mk) DO NOTHING;
INSERT INTO courses (id, kode_mk, nama_mk, sks, semester, kuota) VALUES (5, 'MK005', 'Keamanan Siber', 3, 5, 25) ON CONFLICT (kode_mk) DO NOTHING;
INSERT INTO courses (id, kode_mk, nama_mk, sks, semester, kuota) VALUES (6, 'MK006', 'Machine Learning Terapan', 3, 5, 20) ON CONFLICT (kode_mk) DO NOTHING;
INSERT INTO courses (id, kode_mk, nama_mk, sks, semester, kuota) VALUES (7, 'MK007', 'Desain Interaksi & UX', 2, 5, 35) ON CONFLICT (kode_mk) DO NOTHING;
INSERT INTO courses (id, kode_mk, nama_mk, sks, semester, kuota) VALUES (8, 'MK008', 'Manajemen Proyek TI', 2, 5, 40) ON CONFLICT (kode_mk) DO NOTHING;
INSERT INTO courses (id, kode_mk, nama_mk, sks, semester, kuota) VALUES (9, 'MK009', 'Kapita Selekta Informatika', 2, 5, 2) ON CONFLICT (kode_mk) DO NOTHING;
INSERT INTO courses (id, kode_mk, nama_mk, sks, semester, kuota) VALUES (10, 'MK010', 'Etika Profesi & Kewirausahaan', 2, 5, 50) ON CONFLICT (kode_mk) DO NOTHING;

-- 4. Seed Sample Enrollments
INSERT INTO enrollments (student_id, course_id, tahun_akademik) VALUES (11, 9, '2026/2027-Ganjil') ON CONFLICT DO NOTHING;
INSERT INTO enrollments (student_id, course_id, tahun_akademik) VALUES (12, 9, '2026/2027-Ganjil') ON CONFLICT DO NOTHING;
INSERT INTO enrollments (student_id, course_id, tahun_akademik) VALUES (1, 1, '2026/2027-Ganjil') ON CONFLICT DO NOTHING;
INSERT INTO enrollments (student_id, course_id, tahun_akademik) VALUES (1, 2, '2026/2027-Ganjil') ON CONFLICT DO NOTHING;
INSERT INTO enrollments (student_id, course_id, tahun_akademik) VALUES (4, 1, '2026/2027-Ganjil') ON CONFLICT DO NOTHING;
INSERT INTO enrollments (student_id, course_id, tahun_akademik) VALUES (4, 2, '2026/2027-Ganjil') ON CONFLICT DO NOTHING;
INSERT INTO enrollments (student_id, course_id, tahun_akademik) VALUES (4, 3, '2026/2027-Ganjil') ON CONFLICT DO NOTHING;
INSERT INTO enrollments (student_id, course_id, tahun_akademik) VALUES (4, 4, '2026/2027-Ganjil') ON CONFLICT DO NOTHING;
INSERT INTO enrollments (student_id, course_id, tahun_akademik) VALUES (4, 5, '2026/2027-Ganjil') ON CONFLICT DO NOTHING;
INSERT INTO enrollments (student_id, course_id, tahun_akademik) VALUES (4, 7, '2026/2027-Ganjil') ON CONFLICT DO NOTHING;

-- 5. Sync Sequence IDs
SELECT setval('users_id_seq', (SELECT MAX(id) FROM users));
SELECT setval('students_id_seq', (SELECT MAX(id) FROM students));
SELECT setval('courses_id_seq', (SELECT MAX(id) FROM courses));
SELECT setval('enrollments_id_seq', (SELECT COALESCE(MAX(id), 1) FROM enrollments));