package transport

import (
	"context"
	"strings"

	"github.com/emersonary/appkit/menu"
	menuv1 "github.com/emersonary/appkit/menu/gen/menu/v1"
)

type MenuServer struct {
	menuv1.UnimplementedMenuServiceServer
	menu      *menu.Service
	accountID func(context.Context) (string, error)
}

func NewMenuServer(svc *menu.Service, accountID func(context.Context) (string, error)) *MenuServer {
	return &MenuServer{menu: svc, accountID: accountID}
}

func (s *MenuServer) GetMenu(ctx context.Context, _ *menuv1.GetMenuRequest) (*menuv1.GetMenuResponse, error) {
	accountID, err := s.accountID(ctx)
	if err != nil {
		return nil, err
	}

	layout, err := s.menu.GetMenu(ctx, accountID)
	if err != nil {
		return nil, MapGRPCError(err)
	}

	return layoutToProto(layout), nil
}

func layoutToProto(layout menu.Layout) *menuv1.GetMenuResponse {
	menus := make([]*menuv1.Menu, 0, len(layout.Menus))
	for _, m := range layout.Menus {
		menus = append(menus, &menuv1.Menu{
			Id:    m.ID,
			Name:  m.Name,
			Items: itemsToProto(m.Items),
		})
	}

	return &menuv1.GetMenuResponse{
		Sidebar: &menuv1.SidebarConfig{
			Floating:         layout.Sidebar.Floating,
			HideWhenSelected: layout.Sidebar.HideWhenSelected,
			Locked:           layout.Sidebar.Locked,
			DefaultMenu:      layout.Sidebar.DefaultMenu,
		},
		Menus:                        menus,
		DefaultSelectedPermissionId: layout.DefaultSelectedPermissionID,
	}
}

func itemsToProto(items []menu.ResolvedMenuItem) []*menuv1.MenuItem {
	out := make([]*menuv1.MenuItem, 0, len(items))
	for _, item := range items {
		out = append(out, &menuv1.MenuItem{
			FullId:       item.FullID,
			PermissionId: item.PermissionID,
			Name:         item.Name,
			RouteName:    item.RouteName,
			Icon:         item.Icon,
			Children:     itemsToProto(item.Children),
		})
	}
	return out
}

func AccountIDFromBearer(ctx context.Context, resolve func(context.Context, string) (string, error)) (string, error) {
	token, err := BearerToken(ctx)
	if err != nil {
		return "", err
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return "", MapGRPCError(menu.ErrUnauthenticated)
	}
	return resolve(ctx, token)
}
