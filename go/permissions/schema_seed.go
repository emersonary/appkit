package permissions

import (
	"fmt"
	"strings"
)

func buildSeedSQL(setup Setup, groups, categories, permissions, profiles, profilePerms string) string {
	if len(setup.Groups) == 0 && len(setup.Profiles) == 0 {
		return ""
	}

	var b strings.Builder

	if len(setup.Groups) > 0 {
		b.WriteString(fmt.Sprintf("\nINSERT INTO %s (id_permission_group, name, sort_order, route_prefix) VALUES\n", groups))
		rows := make([]string, 0, len(setup.Groups))
		for _, g := range setup.Groups {
			rows = append(rows, fmt.Sprintf(
				"\t(%s, %s, %d, %s)",
				quoteLiteral(strings.TrimSpace(g.ID)),
				quoteLiteral(strings.TrimSpace(g.Name)),
				g.SortOrder,
				quoteLiteral(strings.TrimSpace(g.RoutePrefix)),
			))
		}
		b.WriteString(strings.Join(rows, ",\n"))
		b.WriteString("\nON CONFLICT (id_permission_group) DO UPDATE SET\n")
		b.WriteString("\tname = EXCLUDED.name,\n")
		b.WriteString("\tsort_order = EXCLUDED.sort_order,\n")
		b.WriteString("\troute_prefix = EXCLUDED.route_prefix;\n")
	}

	if len(setup.Categories) > 0 {
		b.WriteString(fmt.Sprintf("\nINSERT INTO %s (id_permission_category, name, id_permission_group, sort_order) VALUES\n", categories))
		rows := make([]string, 0, len(setup.Categories))
		for _, c := range setup.Categories {
			rows = append(rows, fmt.Sprintf(
				"\t(%s, %s, %s, %d)",
				quoteLiteral(strings.TrimSpace(c.ID)),
				quoteLiteral(strings.TrimSpace(c.Name)),
				quoteLiteral(strings.TrimSpace(c.Group)),
				c.SortOrder,
			))
		}
		b.WriteString(strings.Join(rows, ",\n"))
		b.WriteString("\nON CONFLICT (id_permission_category) DO UPDATE SET\n")
		b.WriteString("\tname = EXCLUDED.name,\n")
		b.WriteString("\tid_permission_group = EXCLUDED.id_permission_group,\n")
		b.WriteString("\tsort_order = EXCLUDED.sort_order;\n")
	}

	if len(setup.Permissions) > 0 {
		b.WriteString(fmt.Sprintf("\nINSERT INTO %s (id_permission, name, id_permission_category, id_parent, be_action, route_name, icon, enabled, sort_order) VALUES\n", permissions))
		rows := make([]string, 0, len(setup.Permissions))
		for _, p := range setup.Permissions {
			enabled := true
			if p.Enabled != nil {
				enabled = *p.Enabled
			}
			parent := strings.TrimSpace(p.Parent)
			parentSQL := "NULL"
			if parent != "" {
				parentSQL = quoteLiteral(parent)
			}
			rows = append(rows, fmt.Sprintf(
				"\t(%s, %s, %s, %s, %d, %s, %s, %t, %d)",
				quoteLiteral(strings.TrimSpace(p.ID)),
				quoteLiteral(strings.TrimSpace(p.Name)),
				quoteLiteral(strings.TrimSpace(p.Category)),
				parentSQL,
				p.BeAction,
				quoteLiteral(strings.TrimSpace(p.RouteName)),
				quoteLiteral(strings.TrimSpace(p.Icon)),
				enabled,
				p.SortOrder,
			))
		}
		b.WriteString(strings.Join(rows, ",\n"))
		b.WriteString("\nON CONFLICT (id_permission) DO UPDATE SET\n")
		b.WriteString("\tname = EXCLUDED.name,\n")
		b.WriteString("\tid_permission_category = EXCLUDED.id_permission_category,\n")
		b.WriteString("\tid_parent = EXCLUDED.id_parent,\n")
		b.WriteString("\tbe_action = EXCLUDED.be_action,\n")
		b.WriteString("\troute_name = EXCLUDED.route_name,\n")
		b.WriteString("\ticon = EXCLUDED.icon,\n")
		b.WriteString("\tenabled = EXCLUDED.enabled,\n")
		b.WriteString("\tsort_order = EXCLUDED.sort_order;\n")
	}

	if len(setup.Profiles) > 0 {
		b.WriteString(fmt.Sprintf("\nINSERT INTO %s (id_profile, name, is_superuser) VALUES\n", profiles))
		rows := make([]string, 0, len(setup.Profiles))
		for _, pr := range setup.Profiles {
			rows = append(rows, fmt.Sprintf(
				"\t(%s, %s, %t)",
				quoteLiteral(strings.TrimSpace(pr.ID)),
				quoteLiteral(strings.TrimSpace(pr.Name)),
				pr.Superuser,
			))
		}
		b.WriteString(strings.Join(rows, ",\n"))
		b.WriteString("\nON CONFLICT (id_profile) DO UPDATE SET\n")
		b.WriteString("\tname = EXCLUDED.name,\n")
		b.WriteString("\tis_superuser = EXCLUDED.is_superuser;\n")
	}

	if len(setup.ProfilePerms) > 0 {
		b.WriteString(fmt.Sprintf("\nINSERT INTO %s (id_profile, id_permission, granted_actions) VALUES\n", profilePerms))
		rows := make([]string, 0, len(setup.ProfilePerms))
		for _, pp := range setup.ProfilePerms {
			grantedSQL := "NULL"
			if pp.GrantedActions != nil {
				grantedSQL = fmt.Sprintf("%d", *pp.GrantedActions)
			}
			rows = append(rows, fmt.Sprintf(
				"\t(%s, %s, %s)",
				quoteLiteral(strings.TrimSpace(pp.Profile)),
				quoteLiteral(strings.TrimSpace(pp.Permission)),
				grantedSQL,
			))
		}
		b.WriteString(strings.Join(rows, ",\n"))
		b.WriteString("\nON CONFLICT (id_profile, id_permission) DO UPDATE SET\n")
		b.WriteString("\tgranted_actions = EXCLUDED.granted_actions;\n")
	}

	return b.String()
}
