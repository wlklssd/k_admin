import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    name: 'KAdmin',
    path: '/kadmin',
    redirect: '/kadmin/users',
    meta: {
      icon: 'lucide:settings-2',
      order: 10,
      title: 'KAdmin 管理',
    },
    children: [
      {
        name: 'KAdminUsers',
        path: 'users',
        component: () =>
          import('#/views/kadmin/components/UserManagementView.vue'),
        meta: {
          icon: 'lucide:users',
          title: '用户管理',
        },
      },
      {
        name: 'KAdminRbac',
        path: 'rbac',
        component: () => import('#/views/kadmin/components/RbacWorkbench.vue'),
        meta: {
          icon: 'lucide:shield-check',
          title: '权限管理',
        },
      },
      {
        name: 'KAdminDictionary',
        path: 'dictionary',
        component: () =>
          import('#/views/kadmin/components/DictionaryManagementView.vue'),
        meta: {
          icon: 'lucide:book-open',
          title: '字典管理',
        },
      },
      {
        name: 'KAdminSettings',
        path: 'settings',
        component: () => import('#/views/kadmin/components/SettingsView.vue'),
        meta: {
          icon: 'lucide:sliders-horizontal',
          title: '参数配置',
        },
      },
      {
        name: 'KAdminResources',
        path: 'resources',
        component: () =>
          import('#/views/kadmin/components/ResourceWorkbench.vue'),
        meta: {
          icon: 'lucide:folder-kanban',
          title: '资源工作台',
        },
      },
      {
        name: 'KAdminLegacyDashboard',
        path: 'legacy-dashboard',
        component: () => import('#/views/kadmin/components/DashboardView.vue'),
        meta: {
          icon: 'lucide:gauge',
          title: '旧版工作台',
        },
      },
    ],
  },
];

export default routes;
