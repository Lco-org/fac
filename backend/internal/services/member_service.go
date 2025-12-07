package services

import (
	"context"
	"cso-erp/internal/database"
	"cso-erp/internal/models"
	"errors"
	"time"

	"github.com/google/uuid"
)

type MemberService struct {
	db *database.DB
}

func NewMemberService(db *database.DB) *MemberService {
	return &MemberService{db: db}
}

func (s *MemberService) ListMembers(ctx context.Context) ([]models.Member, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, membership_number, full_name, full_name_ar, COALESCE(member_type, ''), phone, email,
		       address, address_ar, join_date, COALESCE(status, 'active'), is_active, created_at
		FROM members
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members = []models.Member{}
	for rows.Next() {
		var m models.Member
		err := rows.Scan(
			&m.ID, &m.MembershipNumber, &m.FullName, &m.FullNameAr, &m.MemberType, &m.Phone, &m.Email,
			&m.Address, &m.AddressAr, &m.JoinDate, &m.Status, &m.IsActive, &m.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		members = append(members, m)
	}

	return members, nil
}

func (s *MemberService) GetMemberByID(ctx context.Context, id uuid.UUID) (*models.Member, error) {
	var m models.Member
	err := s.db.QueryRowContext(ctx, `
		SELECT id, membership_number, full_name, full_name_ar, COALESCE(member_type, ''), phone, email,
		       address, address_ar, join_date, COALESCE(status, 'active'), is_active, created_at
		FROM members
		WHERE id = $1
	`, id).Scan(
		&m.ID, &m.MembershipNumber, &m.FullName, &m.FullNameAr, &m.MemberType, &m.Phone, &m.Email,
		&m.Address, &m.AddressAr, &m.JoinDate, &m.Status, &m.IsActive, &m.CreatedAt,
	)

	if err != nil {
		return nil, ErrMemberNotFound
	}

	return &m, nil
}

func (s *MemberService) CreateMember(ctx context.Context, req *models.CreateMemberRequest, createdBy uuid.UUID) (*models.Member, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM members WHERE membership_number = $1)", req.MembershipNumber).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrMembershipNumberExists
	}

	// Set default join date if not provided
	joinDate := time.Now()
	if req.JoinDate != nil {
		joinDate = *req.JoinDate
	}

	var m models.Member
	err = s.db.QueryRowContext(ctx, `
		INSERT INTO members (membership_number, full_name, full_name_ar, member_type, phone, email,
		                     address, address_ar, join_date, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, membership_number, full_name, full_name_ar, member_type, phone, email,
		          address, address_ar, join_date, status, is_active, created_at
	`, req.MembershipNumber, req.FullName, req.FullNameAr, req.MemberType, req.Phone, req.Email,
		req.Address, req.AddressAr, joinDate, createdBy).Scan(
		&m.ID, &m.MembershipNumber, &m.FullName, &m.FullNameAr, &m.MemberType, &m.Phone, &m.Email,
		&m.Address, &m.AddressAr, &m.JoinDate, &m.Status, &m.IsActive, &m.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (s *MemberService) UpdateMember(ctx context.Context, id uuid.UUID, req *models.UpdateMemberRequest) (*models.Member, error) {
	var m models.Member
	err := s.db.QueryRowContext(ctx, `
		UPDATE members SET
			full_name = COALESCE($1, full_name),
			full_name_ar = COALESCE($2, full_name_ar),
			member_type = COALESCE($3, member_type),
			phone = COALESCE($4, phone),
			email = COALESCE($5, email),
			address = COALESCE($6, address),
			address_ar = COALESCE($7, address_ar),
			join_date = COALESCE($8, join_date),
			status = COALESCE($9, status),
			is_active = COALESCE($10, is_active),
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $11
		RETURNING id, membership_number, full_name, full_name_ar, member_type, phone, email,
		          address, address_ar, join_date, status, is_active, created_at
	`, req.FullName, req.FullNameAr, req.MemberType, req.Phone, req.Email, req.Address, req.AddressAr,
		req.JoinDate, req.Status, req.IsActive, id).Scan(
		&m.ID, &m.MembershipNumber, &m.FullName, &m.FullNameAr, &m.MemberType, &m.Phone, &m.Email,
		&m.Address, &m.AddressAr, &m.JoinDate, &m.Status, &m.IsActive, &m.CreatedAt,
	)

	if err != nil {
		return nil, ErrMemberNotFound
	}

	return &m, nil
}

func (s *MemberService) DeleteMember(ctx context.Context, id uuid.UUID) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM members WHERE id = $1`, id)
	if err != nil {
		return errors.New("cannot delete member: associated records exist")
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrMemberNotFound
	}

	return nil
}
