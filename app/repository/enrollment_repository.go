package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"siakad-mini/app/model"
)

type EnrollmentRepository interface {
	Create(ctx context.Context, student model.Student, course model.Course, tahunAkademik string) (model.Enrollment, error)
	FindByID(ctx context.Context, id int) (model.Enrollment, error)
	Delete(ctx context.Context, id int, studentID int) error
	GetTotalSKS(ctx context.Context, studentID int, tahunAkademik string) (int, error)
	GetEnrolledCourses(ctx context.Context, studentID int, tahunAkademik string) ([]model.Course, error)
	GetAllEnrolledCourses(ctx context.Context, studentID int) ([]model.Course, error)
}

type enrollmentPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewEnrollmentRepository(pool *pgxpool.Pool) EnrollmentRepository {
	return &enrollmentPostgresRepository{pool: pool}
}

func (r *enrollmentPostgresRepository) Create(ctx context.Context, student model.Student, reqCourse model.Course, tahunAkademik string) (model.Enrollment, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.Enrollment{}, fmt.Errorf("memulai transaksi: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Lock the course to safely evaluate quota
	var course model.Course
	err = tx.QueryRow(ctx,
		`SELECT c.id, c.sks, c.kuota, 
		        (SELECT COUNT(*) FROM enrollments e WHERE e.course_id = c.id) as terisi
		 FROM courses c WHERE c.id = $1 FOR UPDATE`, reqCourse.ID,
	).Scan(&course.ID, &course.SKS, &course.Kuota, &course.Terisi)
	
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Enrollment{}, ErrNotFound
		}
		return model.Enrollment{}, fmt.Errorf("lock course: %w", err)
	}

	course.SisaKuota = course.Kuota - course.Terisi
	if course.SisaKuota <= 0 {
		return model.Enrollment{}, fmt.Errorf("kuota penuh: %w", ErrConflict)
	}

	// 2. Evaluate Max SKS Limits based on IPK
	var maxSKS int
	if student.IPKTerakhir >= 3.00 {
		maxSKS = 24
	} else if student.IPKTerakhir >= 2.50 {
		maxSKS = 21
	} else {
		maxSKS = 18
	}

	// Calculate current enrolled SKS in the same academic year
	var currentSKS int
	err = tx.QueryRow(ctx,
		`SELECT COALESCE(SUM(c.sks), 0) FROM enrollments e 
		 JOIN courses c ON c.id = e.course_id 
		 WHERE e.student_id = $1 AND e.tahun_akademik = $2`, student.ID, tahunAkademik,
	).Scan(&currentSKS)
	if err != nil {
		return model.Enrollment{}, fmt.Errorf("hitung SKS: %w", err)
	}

	if currentSKS + course.SKS > maxSKS {
		return model.Enrollment{}, fmt.Errorf("batas SKS terlampaui (max: %d, diambil: %d, matkul: %d): %w", maxSKS, currentSKS, course.SKS, ErrConflict)
	}

	// 3. Insert Enrollment
	var en model.Enrollment
	err = tx.QueryRow(ctx,
		`INSERT INTO enrollments (student_id, course_id, tahun_akademik) VALUES ($1, $2, $3) RETURNING id, created_at`,
		student.ID, course.ID, tahunAkademik,
	).Scan(&en.ID, &en.CreatedAt)

	if err != nil {
		if isUniqueViolation(err) {
			return model.Enrollment{}, fmt.Errorf("sudah diambil: %w", ErrDuplicate)
		}
		return model.Enrollment{}, fmt.Errorf("insert enrollment: %w", err)
	}
	en.StudentID = student.ID
	en.CourseID = course.ID
	en.TahunAkademik = tahunAkademik

	if err := tx.Commit(ctx); err != nil {
		return model.Enrollment{}, fmt.Errorf("commit transaksi: %w", err)
	}
	return en, nil
}

func (r *enrollmentPostgresRepository) FindByID(ctx context.Context, id int) (model.Enrollment, error) {
	var en model.Enrollment
	err := r.pool.QueryRow(ctx,
		"SELECT id, student_id, course_id, tahun_akademik, created_at FROM enrollments WHERE id = $1", id,
	).Scan(&en.ID, &en.StudentID, &en.CourseID, &en.TahunAkademik, &en.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Enrollment{}, ErrNotFound
		}
		return model.Enrollment{}, fmt.Errorf("find enrollment by id: %w", err)
	}
	return en, nil
}

func (r *enrollmentPostgresRepository) Delete(ctx context.Context, id int, studentID int) error {
	tag, err := r.pool.Exec(ctx, "DELETE FROM enrollments WHERE id = $1 AND student_id = $2", id, studentID)
	if err != nil {
		return fmt.Errorf("menghapus enrollment: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *enrollmentPostgresRepository) GetTotalSKS(ctx context.Context, studentID int, tahunAkademik string) (int, error) {
	var totalSKS int
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(c.sks), 0) FROM enrollments e 
		 JOIN courses c ON c.id = e.course_id 
		 WHERE e.student_id = $1 AND e.tahun_akademik = $2`, studentID, tahunAkademik,
	).Scan(&totalSKS)
	if err != nil {
		return 0, fmt.Errorf("hitung total sks: %w", err)
	}
	return totalSKS, nil
}

func (r *enrollmentPostgresRepository) GetEnrolledCourses(ctx context.Context, studentID int, tahunAkademik string) ([]model.Course, error) {
	sqlText := `
		SELECT c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota, c.created_at 
		FROM courses c
		JOIN enrollments e ON e.course_id = c.id
		WHERE e.student_id = $1 AND e.tahun_akademik = $2
	`
	rows, err := r.pool.Query(ctx, sqlText, studentID, tahunAkademik)
	if err != nil {
		return nil, fmt.Errorf("ambil course: %w", err)
	}
	defer rows.Close()

	hasil := []model.Course{}
	for rows.Next() {
		var c model.Course
		if err := rows.Scan(&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan course: %w", err)
		}
		hasil = append(hasil, c)
	}

	return hasil, nil
}

func (r *enrollmentPostgresRepository) GetAllEnrolledCourses(ctx context.Context, studentID int) ([]model.Course, error) {
	sqlText := `
		SELECT c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota, c.created_at 
		FROM courses c
		JOIN enrollments e ON e.course_id = c.id
		WHERE e.student_id = $1
		ORDER BY e.created_at DESC
	`
	rows, err := r.pool.Query(ctx, sqlText, studentID)
	if err != nil {
		return nil, fmt.Errorf("ambil semua course: %w", err)
	}
	defer rows.Close()

	hasil := []model.Course{}
	for rows.Next() {
		var c model.Course
		if err := rows.Scan(&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan course: %w", err)
		}
		hasil = append(hasil, c)
	}

	return hasil, nil
}
