package permissions

import (
	"context"
	"database/sql"
	"fmt"
)

type Store struct {
	db              *sql.DB
	schema          string
	accountsSchema  string
}

func NewStore(db *sql.DB, setup Setup) *Store {
	return &Store{
		db:             db,
		schema:         setup.Schema,
		accountsSchema: setup.AccountsSchema,
	}
}

func (s *Store) table(name string) string {
	return qualifiedName(s.schema, name)
}

func (s *Store) accountsTable() string {
	return qualifiedName(s.accountsSchema, "accounts")
}

func (s *Store) GetProfile(ctx context.Context, idProfile string) (Profile, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id_profile, name, is_superuser, created_at, updated_at
		FROM `+s.table("profiles")+`
		WHERE id_profile = $1
	`, idProfile)

	var item Profile
	if err := row.Scan(&item.IDProfile, &item.Name, &item.IsSuperuser, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return Profile{}, err
	}
	return item, nil
}

func (s *Store) GetAccountProfileID(ctx context.Context, accountID string) (string, error) {
	var idProfile sql.NullString
	err := s.db.QueryRowContext(ctx, `
		SELECT id_profile FROM `+s.accountsTable()+` WHERE id = $1
	`, accountID).Scan(&idProfile)
	if err != nil {
		return "", err
	}
	if !idProfile.Valid || idProfile.String == "" {
		return "", sql.ErrNoRows
	}
	return idProfile.String, nil
}

func (s *Store) SetAccountProfileID(ctx context.Context, accountID, idProfile string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE `+s.accountsTable()+`
		SET id_profile = $2, updated_at = NOW()
		WHERE id = $1
	`, accountID, nullIfEmpty(idProfile))
	return err
}

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}

type permissionGrantRow struct {
	IDPermission   string
	BeAction       int
	Enabled        bool
	GrantedActions sql.NullInt64
	IsSuperuser    bool
	HasAssignment  bool
}

func (s *Store) loadGrant(ctx context.Context, idProfile, idPermission string) (permissionGrantRow, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT p.id_permission, p.be_action, p.enabled, pp.granted_actions, pr.is_superuser,
		       (pp.id_permission IS NOT NULL) AS has_assignment
		FROM `+s.table("profiles")+` pr
		INNER JOIN `+s.table("permissions")+` p ON p.id_permission = $2
		LEFT JOIN `+s.table("profile_permissions")+` pp
			ON pp.id_profile = pr.id_profile AND pp.id_permission = p.id_permission
		WHERE pr.id_profile = $1
	`, idProfile, idPermission)

	var grant permissionGrantRow
	if err := row.Scan(&grant.IDPermission, &grant.BeAction, &grant.Enabled, &grant.GrantedActions, &grant.IsSuperuser, &grant.HasAssignment); err != nil {
		return permissionGrantRow{}, err
	}
	return grant, nil
}

func (s *Store) ListGroups(ctx context.Context) ([]Group, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id_permission_group, name, sort_order, route_prefix, created_at, updated_at
		FROM `+s.table("permission_groups")+`
		ORDER BY sort_order, id_permission_group
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Group
	for rows.Next() {
		var item Group
		if err := rows.Scan(&item.IDPermissionGroup, &item.Name, &item.SortOrder, &item.RoutePrefix, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) ListCategories(ctx context.Context) ([]Category, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id_permission_category, name, id_permission_group, sort_order, created_at, updated_at
		FROM `+s.table("permission_categories")+`
		ORDER BY sort_order, id_permission_category
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Category
	for rows.Next() {
		var item Category
		if err := rows.Scan(&item.IDPermissionCategory, &item.Name, &item.IDPermissionGroup, &item.SortOrder, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) ListPermissions(ctx context.Context) ([]Permission, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id_permission, name, id_permission_category, id_parent, be_action, route_name, icon, enabled, sort_order, created_at, updated_at
		FROM `+s.table("permissions")+`
		ORDER BY sort_order, id_permission
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanPermissions(rows)
}

func (s *Store) ListFlatForProfile(ctx context.Context, idProfile string) ([]FlatPermission, error) {
	profile, err := s.GetProfile(ctx, idProfile)
	if err != nil {
		return nil, err
	}

	if profile.IsSuperuser {
		items, err := s.ListPermissions(ctx)
		if err != nil {
			return nil, err
		}
		out := make([]FlatPermission, 0, len(items))
		for _, item := range items {
			if !item.Enabled {
				continue
			}
			out = append(out, FlatPermission{
				Permission:     item,
				GrantedActions: item.BeAction,
			})
		}
		return out, nil
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT p.id_permission, p.name, p.id_permission_category, p.id_parent, p.be_action, p.route_name, p.icon, p.enabled, p.sort_order, p.created_at, p.updated_at,
		       pp.granted_actions
		FROM `+s.table("profile_permissions")+` pp
		INNER JOIN `+s.table("permissions")+` p ON p.id_permission = pp.id_permission
		WHERE pp.id_profile = $1 AND p.enabled = true
		ORDER BY p.sort_order, p.id_permission
	`, idProfile)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []FlatPermission
	for rows.Next() {
		var item Permission
		var granted sql.NullInt64
		if err := rows.Scan(
			&item.IDPermission, &item.Name, &item.IDPermissionCategory, &item.IDParent,
			&item.BeAction, &item.RouteName, &item.Icon, &item.Enabled, &item.SortOrder,
			&item.CreatedAt, &item.UpdatedAt, &granted,
		); err != nil {
			return nil, err
		}
		effective := effectiveGrantedActions(item.BeAction, granted)
		out = append(out, FlatPermission{Permission: item, GrantedActions: effective})
	}
	return out, rows.Err()
}

func effectiveGrantedActions(beAction int, granted sql.NullInt64) int {
	if !granted.Valid {
		return beAction
	}
	return beAction & int(granted.Int64)
}

func scanPermissions(rows *sql.Rows) ([]Permission, error) {
	var out []Permission
	for rows.Next() {
		var item Permission
		if err := rows.Scan(
			&item.IDPermission, &item.Name, &item.IDPermissionCategory, &item.IDParent,
			&item.BeAction, &item.RouteName, &item.Icon, &item.Enabled, &item.SortOrder,
			&item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) HasProfilePermission(ctx context.Context, idProfile, idPermission string, actionBit int) (bool, error) {
	grant, err := s.loadGrant(ctx, idProfile, idPermission)
	if err != nil {
		return false, err
	}
	if grant.IsSuperuser {
		return grant.Enabled && MaskAllows(grant.BeAction, actionBit), nil
	}
	if !grant.HasAssignment || !grant.Enabled {
		return false, nil
	}
	if !grant.GrantedActions.Valid {
		return MaskAllows(grant.BeAction, actionBit), nil
	}
	allowed := int(grant.GrantedActions.Int64)
	return MaskAllows(allowed, actionBit), nil
}

func (s *Store) EnsureAccountDefaultProfile(ctx context.Context, accountID, defaultProfile string) error {
	_, err := s.db.ExecContext(ctx, fmt.Sprintf(`
		UPDATE %s
		SET id_profile = $2, updated_at = NOW()
		WHERE id = $1 AND (id_profile IS NULL OR id_profile = '')
	`, s.accountsTable()), accountID, defaultProfile)
	return err
}
