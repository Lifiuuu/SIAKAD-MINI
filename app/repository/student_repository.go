package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"siakad-mini/app/model"
)

type StudentRepository interface {
	FindAll(ctx context.Context, q model.ListQuery) ([]model.Student, int, error)
	FindByID(ctx context.Context, id int) (model.Student, error)
	FindByUserID(ctx context.Context, userID int) (model.Student, error)
	CreateWithUser(ctx context.Context, s model.Student, u model.User) (model.Student, error)
	Update(ctx context.Context, s model.Student) (model.Student, error)
	SoftDelete(ctx context.Context, id int) error
}

type studentPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewStudentRepository(pool *pgxpool.Pool) StudentRepository {
	return &studentPostgresRepository{pool: pool}
}

var kolomUrutStudent = map[string]string{
	"id":           "id",
	"ipk_terakhir": "ipk_terakhir",
	"nama":         "LOWER(nama)",
}

func buildStudentFilter(q model.ListQuery) (string, []any) {
	where := " WHERE deleted_at IS NULL"
	args := []any{}

	if q.Search != "" {
		where += fmt.Sprintf(" AND (nama ILIKE $%d OR nim ILIKE $%d)", len(args)+1, len(args)+1)
		args = append(args, "%"+q.Search+"%")
	}

	if q.Prodi != "" {
		where += fmt.Sprintf(" AND prodi ILIKE $%d", len(args)+1)
		args = append(args, "%"+q.Prodi+"%")
	}

	if q.Angkatan > 0 {
		where += fmt.Sprintf(" AND angkatan = $%d", len(args)+1)
		args = append(args, q.Angkatan)
	}

	return where, args
}

func (r *studentPostgresRepository) FindAll(ctx context.Context, q model.ListQuery) ([]model.Student, int, error) {
	where, args := buildStudentFilter(q)

	var total int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM students"+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("menghitung students: %w", err)
	}

	arah := "ASC"
	if q.Order == "desc" {
		arah = "DESC"
	}

	sqlText := fmt.Sprintf(
		"SELECT id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, created_at FROM students%s ORDER BY %s %s LIMIT $%d OFFSET $%d",
		where, func() string {
			if col, ok := kolomUrutStudent[q.Sort]; ok {
				return col
			}
			return "id"
		}(), arah, len(args)+1, len(args)+2,
	)
	args = append(args, q.Limit, q.Offset())

	rows, err := r.pool.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("mengambil daftar students: %w", err)
	}
	defer rows.Close()

	hasil := []model.Student{}
	for rows.Next() {
		var s model.Student
		if err := rows.Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("membaca baris student: %w", err)
		}
		hasil = append(hasil, s)
	}

	return hasil, total, nil
}

func (r *studentPostgresRepository) FindByID(ctx context.Context, id int) (model.Student, error) {
	var s model.Student
	err := r.pool.QueryRow(ctx,
		"SELECT id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, created_at FROM students WHERE id = $1 AND deleted_at IS NULL", id,
	).Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		return model.Student{}, fmt.Errorf("mengambil student: %w", err)
	}
	return s, nil
}

func (r *studentPostgresRepository) FindByUserID(ctx context.Context, userID int) (model.Student, error) {
	var s model.Student
	err := r.pool.QueryRow(ctx,
		"SELECT id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, created_at FROM students WHERE user_id = $1 AND deleted_at IS NULL", userID,
	).Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		return model.Student{}, fmt.Errorf("mengambil student by user_id: %w", err)
	}
	return s, nil
}

func (r *studentPostgresRepository) CreateWithUser(ctx context.Context, s model.Student, u model.User) (model.Student, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.Student{}, fmt.Errorf("memulai transaksi: %w", err)
	}
	defer tx.Rollback(ctx)

	var userID int
	err = tx.QueryRow(ctx,
		"INSERT INTO users (email, password, role) VALUES ($1, $2, $3) RETURNING id",
		u.Email, u.Password, u.Role,
	).Scan(&userID)
	if err != nil {
		if isUniqueViolation(err) {
			return model.Student{}, ErrDuplicate
		}
		return model.Student{}, fmt.Errorf("menyimpan user: %w", err)
	}

	err = tx.QueryRow(ctx,
		"INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at",
		userID, s.NIM, s.Nama, s.Prodi, s.Angkatan, s.IPKTerakhir,
	).Scan(&s.ID, &s.CreatedAt)

	if err != nil {
		if isUniqueViolation(err) {
			return model.Student{}, ErrDuplicate
		}
		return model.Student{}, fmt.Errorf("menyimpan student: %w", err)
	}
	
	s.UserID = userID

	if err := tx.Commit(ctx); err != nil {
		return model.Student{}, fmt.Errorf("commit transaksi: %w", err)
	}
	return s, nil
}

func (r *studentPostgresRepository) Update(ctx context.Context, s model.Student) (model.Student, error) {
	err := r.pool.QueryRow(ctx,
		"UPDATE students SET nama = $1, prodi = $2, angkatan = $3, ipk_terakhir = $4 WHERE id = $5 AND deleted_at IS NULL RETURNING id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, created_at",
		s.Nama, s.Prodi, s.Angkatan, s.IPKTerakhir, s.ID,
	).Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		if isUniqueViolation(err) {
			return model.Student{}, ErrDuplicate
		}
		return model.Student{}, fmt.Errorf("memperbarui student: %w", err)
	}
	return s, nil
}

func (r *studentPostgresRepository) SoftDelete(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(ctx, "UPDATE students SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL", id)
	if err != nil {
		return fmt.Errorf("soft delete student: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
