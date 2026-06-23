export { Sidebar, type SidebarProps } from "./Sidebar";
export { MenuProvider, ConnectedSidebar, useMenu } from "./MenuProvider";
export { MenuService } from "./gen/v1/menu_connect";
export type {
  GetMenuResponse,
  Menu,
  MenuItem,
  SidebarConfig,
} from "./gen/v1/menu_pb";
