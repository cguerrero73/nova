package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/nova/backend/internal/domain/events"
)

type pgEventRepository struct {
	pool *pgxpool.Pool
}

func NewPgEventRepository(pool *pgxpool.Pool) *pgEventRepository {
	return &pgEventRepository{pool: pool}
}

func (r *pgEventRepository) FindByID(ctx context.Context, id string) (*events.Event, error) {
	query := `
		SELECT evt_id, evt_code, evt_org, evt_desc, evt_type, evt_rtype, evt_status,
		       evt_rstatus, evt_object, evt_object_org, evt_notused, evt_tenant_id,
		       evt_created_at, evt_updated_at, evt_created_by, evt_updated_by
		FROM eamevents WHERE evt_id = $1`

	var e events.Event
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&e.ID, &e.Code, &e.Org, &e.Desc, &e.Type, &e.RType, &e.Status,
		&e.RStatus, &e.Object, &e.ObjectOrg, &e.NotUsed, &e.TenantID,
		&e.CreatedAt, &e.UpdatedAt, &e.CreatedBy, &e.UpdatedBy,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &e, err
}

func (r *pgEventRepository) FindByCode(ctx context.Context, code string) (*events.Event, error) {
	query := `
		SELECT evt_id, evt_code, evt_org, evt_desc, evt_type, evt_rtype, evt_status,
		       evt_rstatus, evt_object, evt_object_org, evt_notused, evt_tenant_id,
		       evt_created_at, evt_updated_at, evt_created_by, evt_updated_by
		FROM eamevents WHERE evt_code = $1`

	var e events.Event
	err := r.pool.QueryRow(ctx, query, code).Scan(
		&e.ID, &e.Code, &e.Org, &e.Desc, &e.Type, &e.RType, &e.Status,
		&e.RStatus, &e.Object, &e.ObjectOrg, &e.NotUsed, &e.TenantID,
		&e.CreatedAt, &e.UpdatedAt, &e.CreatedBy, &e.UpdatedBy,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &e, err
}

func (r *pgEventRepository) FindAll(ctx context.Context, tenantID string, org string, limit, offset int) ([]*events.Event, int, error) {
	countQuery := `SELECT COUNT(*) FROM eamevents WHERE evt_tenant_id = $1`
	args := []interface{}{tenantID}

	if org != "" {
		countQuery += ` AND evt_org = $2`
		args = append(args, org)
	}

	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT evt_id, evt_code, evt_org, evt_desc, evt_type, evt_rtype, evt_status,
		       evt_rstatus, evt_object, evt_object_org, evt_notused, evt_tenant_id,
		       evt_created_at, evt_updated_at, evt_created_by, evt_updated_by
		FROM eamevents WHERE evt_tenant_id = $1`

	if org != "" {
		query += ` AND evt_org = $2`
	}
	query += ` ORDER BY evt_created_at DESC LIMIT $3 OFFSET $4`
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var evts []*events.Event
	for rows.Next() {
		var e events.Event
		err := rows.Scan(
			&e.ID, &e.Code, &e.Org, &e.Desc, &e.Type, &e.RType, &e.Status,
			&e.RStatus, &e.Object, &e.ObjectOrg, &e.NotUsed, &e.TenantID,
			&e.CreatedAt, &e.UpdatedAt, &e.CreatedBy, &e.UpdatedBy,
		)
		if err != nil {
			return nil, 0, err
		}
		evts = append(evts, &e)
	}

	return evts, total, nil
}

func (r *pgEventRepository) FindByOrg(ctx context.Context, org string) ([]*events.Event, error) {
	query := `
		SELECT evt_id, evt_code, evt_org, evt_desc, evt_type, evt_rtype, evt_status,
		       evt_rstatus, evt_object, evt_object_org, evt_notused, evt_tenant_id,
		       evt_created_at, evt_updated_at, evt_created_by, evt_updated_by
		FROM eamevents WHERE evt_org = $1 AND evt_notused IS NULL
		ORDER BY evt_code ASC`

	rows, err := r.pool.Query(ctx, query, org)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var evts []*events.Event
	for rows.Next() {
		var e events.Event
		err := rows.Scan(
			&e.ID, &e.Code, &e.Org, &e.Desc, &e.Type, &e.RType, &e.Status,
			&e.RStatus, &e.Object, &e.ObjectOrg, &e.NotUsed, &e.TenantID,
			&e.CreatedAt, &e.UpdatedAt, &e.CreatedBy, &e.UpdatedBy,
		)
		if err != nil {
			return nil, err
		}
		evts = append(evts, &e)
	}

	return evts, nil
}

func (r *pgEventRepository) FindByObject(ctx context.Context, objectCode, objectOrg string) ([]*events.Event, error) {
	query := `
		SELECT evt_id, evt_code, evt_org, evt_desc, evt_type, evt_rtype, evt_status,
		       evt_rstatus, evt_object, evt_object_org, evt_notused, evt_tenant_id,
		       evt_created_at, evt_updated_at, evt_created_by, evt_updated_by
		FROM eamevents WHERE evt_object = $1 AND evt_object_org = $2`

	rows, err := r.pool.Query(ctx, query, objectCode, objectOrg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var evts []*events.Event
	for rows.Next() {
		var e events.Event
		err := rows.Scan(
			&e.ID, &e.Code, &e.Org, &e.Desc, &e.Type, &e.RType, &e.Status,
			&e.RStatus, &e.Object, &e.ObjectOrg, &e.NotUsed, &e.TenantID,
			&e.CreatedAt, &e.UpdatedAt, &e.CreatedBy, &e.UpdatedBy,
		)
		if err != nil {
			return nil, err
		}
		evts = append(evts, &e)
	}

	return evts, nil
}

func (r *pgEventRepository) FindByType(ctx context.Context, typeCode string) ([]*events.Event, error) {
	query := `
		SELECT evt_id, evt_code, evt_org, evt_desc, evt_type, evt_rtype, evt_status,
		       evt_rstatus, evt_object, evt_object_org, evt_notused, evt_tenant_id,
		       evt_created_at, evt_updated_at, evt_created_by, evt_updated_by
		FROM eamevents WHERE evt_type = $1`

	rows, err := r.pool.Query(ctx, query, typeCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var evts []*events.Event
	for rows.Next() {
		var e events.Event
		err := rows.Scan(
			&e.ID, &e.Code, &e.Org, &e.Desc, &e.Type, &e.RType, &e.Status,
			&e.RStatus, &e.Object, &e.ObjectOrg, &e.NotUsed, &e.TenantID,
			&e.CreatedAt, &e.UpdatedAt, &e.CreatedBy, &e.UpdatedBy,
		)
		if err != nil {
			return nil, err
		}
		evts = append(evts, &e)
	}

	return evts, nil
}

func (r *pgEventRepository) FindByStatus(ctx context.Context, status string) ([]*events.Event, error) {
	query := `
		SELECT evt_id, evt_code, evt_org, evt_desc, evt_type, evt_rtype, evt_status,
		       evt_rstatus, evt_object, evt_object_org, evt_notused, evt_tenant_id,
		       evt_created_at, evt_updated_at, evt_created_by, evt_updated_by
		FROM eamevents WHERE evt_status = $1`

	rows, err := r.pool.Query(ctx, query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var evts []*events.Event
	for rows.Next() {
		var e events.Event
		err := rows.Scan(
			&e.ID, &e.Code, &e.Org, &e.Desc, &e.Type, &e.RType, &e.Status,
			&e.RStatus, &e.Object, &e.ObjectOrg, &e.NotUsed, &e.TenantID,
			&e.CreatedAt, &e.UpdatedAt, &e.CreatedBy, &e.UpdatedBy,
		)
		if err != nil {
			return nil, err
		}
		evts = append(evts, &e)
	}

	return evts, nil
}

func (r *pgEventRepository) Create(ctx context.Context, e *events.Event) error {
	query := `
		INSERT INTO eamevents (evt_id, evt_code, evt_org, evt_desc, evt_type, evt_rtype,
		                       evt_status, evt_rstatus, evt_object, evt_object_org,
		                       evt_notused, evt_tenant_id, evt_created_at, evt_updated_at,
		                       evt_created_by, evt_updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`

	_, err := r.pool.Exec(ctx, query,
		e.ID, e.Code, e.Org, e.Desc, e.Type, e.RType, e.Status,
		e.RStatus, e.Object, e.ObjectOrg, e.NotUsed, e.TenantID,
		e.CreatedAt, e.UpdatedAt, e.CreatedBy, e.UpdatedBy,
	)
	return err
}

func (r *pgEventRepository) Update(ctx context.Context, e *events.Event) error {
	query := `
		UPDATE eamevents 
		SET evt_org = $2, evt_desc = $3, evt_type = $4, evt_rtype = $5,
		    evt_status = $6, evt_rstatus = $7, evt_object = $8, evt_object_org = $9,
		    evt_notused = $10, evt_updated_at = $11, evt_updated_by = $12
		WHERE evt_id = $1`

	_, err := r.pool.Exec(ctx, query,
		e.ID, e.Org, e.Desc, e.Type, e.RType, e.Status,
		e.RStatus, e.Object, e.ObjectOrg, e.NotUsed, e.UpdatedAt, e.UpdatedBy,
	)
	return err
}

func (r *pgEventRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE eamevents SET evt_notused = '+' WHERE evt_id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}
