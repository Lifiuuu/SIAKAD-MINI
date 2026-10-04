package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"siakad-mini/app/model"
)

type CourseRepository interface {
	FindAll(ctx context.Context, q model.ListQuery) ([]model.Course, int, error)
	FindByID(ctx context.Context, id int) (model.Course, error)
}

type coursePostgresRepository struct {
	pool *pgxpool.Pool
}

func NewCourseRepository(pool *pgxpool.Pool) CourseRepository {
	return &coursePostgresRepository{pool: pool}
}

func buildCourseFilter(q model.ListQuery) (string, []any) {
	where := " WHERE 1=1"
	args := []any{}

	if q.Search != "" {
		where += fmt.Sprintf(" AND (nama_mk ILIKE $%d OR kode_mk ILIKE $%d)", len(args)+1, len(args)+1)
		args = append(args, "%"+q.Search+"%")
	}

	if q.Semester > 0 {
		where += fmt.Sprintf(" AND semester = $%d", len(args)+1)
		args = append(args, q.Semester)
	}

	return where, args
}

func (r *coursePostgresRepository) FindAll(ctx context.Context, q model.ListQuery) ([]model.Course, int, error) {
	where, args := buildCourseFilter(q)

	var total int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM courses"+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("menghitung courses: %w", err)
	}

	arah := "ASC"
	if q.Order == "desc" {
		arah = "DESC"
	}

	// For courses, we need to return `terisi` and `sisa_kuota`
	sqlText := fmt.Sprintf(
		`SELECT c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota, 
		        (SELECT COUNT(*) FROM enrollments e WHERE e.course_id = c.id) as terisi,
		        c.created_at 
		 FROM courses c %s ORDER BY c.id %s LIMIT $%d OFFSET $%d`,
		where, arah, len(args)+1, len(args)+2,
	)
	args = append(args, q.Limit, q.Offset())

	rows, err := r.pool.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("mengambil daftar courses: %w", err)
	}
	defer rows.Close()

	hasil := []model.Course{}
	for rows.Next() {
		var c model.Course
		if err := rows.Scan(&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota, &c.Terisi, &c.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("membaca baris course: %w", err)
		}
		c.SisaKuota = c.Kuota - c.Terisi
		hasil = append(hasil, c)
	}

	return hasil, total, nil
}

func (r *coursePostgresRepository) FindByID(ctx context.Context, id int) (model.Course, error) {
	var c model.Course
	err := r.pool.QueryRow(ctx,
		`SELECT c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota, 
		        (SELECT COUNT(*) FROM enrollments e WHERE e.course_id = c.id) as terisi,
		        c.created_at 
		 FROM courses c WHERE c.id = $1`, id,
	).Scan(&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota, &c.Terisi, &c.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Course{}, ErrNotFound
		}
		return model.Course{}, fmt.Errorf("mengambil course: %w", err)
	}
	c.SisaKuota = c.Kuota - c.Terisi
	return c, nil
}
