package bootstrap

type MenuSeed struct {
	Title    string
	Icon     string
	URI      string
	Order    int64
	Children []MenuSeed
}

func DefaultMenus() []MenuSeed {
	return []MenuSeed{
		{
			Title: "Dashboard",
			Icon:  "lucide:layout-dashboard",
			URI:   "/dashboard",
			Order: 1,
			Children: []MenuSeed{
				{Title: "分析页", Icon: "lucide:area-chart", URI: "/dashboard/analytics", Order: 1},
				{Title: "工作台", Icon: "carbon:workspace", URI: "/dashboard/workspace", Order: 2},
			},
		},
		{
			Title: "KAdmin 管理",
			Icon:  "lucide:settings-2",
			URI:   "/kadmin",
			Order: 10,
			Children: []MenuSeed{
				{Title: "用户管理", Icon: "lucide:users", URI: "/kadmin/users", Order: 1},
				{Title: "权限管理", Icon: "lucide:shield-check", URI: "/kadmin/rbac", Order: 2},
				{Title: "菜单管理", Icon: "lucide:menu", URI: "/kadmin/menus", Order: 3},
				{Title: "字典管理", Icon: "lucide:book-open", URI: "/kadmin/dictionary", Order: 4},
				{Title: "参数配置", Icon: "lucide:sliders-horizontal", URI: "/kadmin/settings", Order: 5},
				{Title: "资源工作台", Icon: "lucide:folder-kanban", URI: "/kadmin/resources", Order: 6},
				{Title: "日志管理", Icon: "lucide:scroll-text", URI: "/kadmin/logs", Order: 7},
				{Title: "定时任务", Icon: "lucide:clock-3", URI: "/kadmin/jobs", Order: 8},
			},
		},
	}
}
