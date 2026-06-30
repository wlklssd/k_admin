<template>
  <a-config-provider :theme="theme">
    <div v-if="!session.token" class="login-view">
      <section class="login-panel" aria-label="登录">
        <div class="login-brand">
          <span class="brand-mark">K</span>
          <div>
            <strong>KAdmin</strong>
            <span>Ant Design Vue</span>
          </div>
        </div>

        <a-form
          :model="loginForm"
          layout="vertical"
          autocomplete="off"
          @finish="handleLogin"
        >
          <a-form-item
            label="用户名"
            name="username"
            :rules="[{ required: true, message: '请输入用户名' }]"
          >
            <a-input v-model:value="loginForm.username" size="large" placeholder="admin">
              <template #prefix>
                <UserOutlined />
              </template>
            </a-input>
          </a-form-item>

          <a-form-item
            label="密码"
            name="password"
            :rules="[{ required: true, message: '请输入密码' }]"
          >
            <a-input-password v-model:value="loginForm.password" size="large" placeholder="admin">
              <template #prefix>
                <LockOutlined />
              </template>
            </a-input-password>
          </a-form-item>

          <div class="login-options">
            <a-checkbox v-model:checked="loginForm.remember">记住登录</a-checkbox>
            <a-tag color="blue">/api</a-tag>
          </div>

          <a-space direction="vertical" class="full-width" :size="10">
            <a-button type="primary" size="large" block html-type="submit" :loading="loginLoading">
              登录
            </a-button>
            <a-button size="large" block @click="enterDemo">
              演示登录
            </a-button>
          </a-space>
        </a-form>
      </section>
    </div>

    <a-layout v-else class="app-shell">
      <a-layout-sider
        v-model:collapsed="collapsed"
        class="app-sider"
        theme="light"
        collapsible
        breakpoint="lg"
      >
        <div class="side-logo">
          <span class="brand-mark">K</span>
          <strong v-if="!collapsed">KAdmin</strong>
        </div>
        <a-menu
          v-model:selectedKeys="selectedKeys"
          mode="inline"
          :items="menuItems"
          @click="handleMenuClick"
        />
      </a-layout-sider>

      <a-layout>
        <a-layout-header class="app-header">
          <div class="header-left">
            <a-button type="text" class="icon-btn" @click="collapsed = !collapsed">
              <MenuUnfoldOutlined v-if="collapsed" />
              <MenuFoldOutlined v-else />
            </a-button>
            <a-breadcrumb>
              <a-breadcrumb-item>后台</a-breadcrumb-item>
              <a-breadcrumb-item>{{ currentTitle }}</a-breadcrumb-item>
            </a-breadcrumb>
          </div>

          <div class="header-actions">
            <a-input-search
              v-model:value="quickKeyword"
              class="quick-search"
              placeholder="搜索"
              allow-clear
              @search="handleQuickSearch"
            />
            <a-badge :count="noticeCount" size="small">
              <a-button shape="circle" class="icon-btn" @click="noticeCount = 0">
                <BellOutlined />
              </a-button>
            </a-badge>
            <a-dropdown>
              <button class="user-menu" type="button">
                <a-avatar size="small">{{ userInitial }}</a-avatar>
                <span>{{ displayName }}</span>
              </button>
              <template #overlay>
                <a-menu @click="handleUserMenu">
                  <a-menu-item key="profile">
                    <UserOutlined />
                    个人信息
                  </a-menu-item>
                  <a-menu-divider />
                  <a-menu-item key="logout">
                    <LogoutOutlined />
                    退出登录
                  </a-menu-item>
                </a-menu>
              </template>
            </a-dropdown>
          </div>
        </a-layout-header>

        <a-layout-content class="app-content">
          <DashboardView v-if="activePage === 'dashboard'" />
          <ResourceWorkbench v-else-if="activePage === 'components'" />
          <SettingsView v-else-if="activePage === 'settings'" />
          <NotFoundView v-else @go-home="navigateTo('dashboard')" />
        </a-layout-content>
      </a-layout>
    </a-layout>
  </a-config-provider>
</template>

<script setup lang="ts">
import {
  AppstoreOutlined,
  BellOutlined,
  DashboardOutlined,
  LockOutlined,
  LogoutOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  SettingOutlined,
  UserOutlined,
} from '@ant-design/icons-vue';
import { message } from 'ant-design-vue';
import { computed, h, onBeforeUnmount, onMounted, reactive, ref } from 'vue';

import { getUserInfo, getUserMenu, login, logout, type UserInfo } from './api/auth';
import DashboardView from './components/DashboardView.vue';
import NotFoundView from './components/NotFoundView.vue';
import ResourceWorkbench from './components/ResourceWorkbench.vue';
import SettingsView from './components/SettingsView.vue';
import { demoUser } from './mock';
import { getStoredToken, removeStoredToken, setStoredToken } from './utils/storage';

const theme = {
  token: {
    borderRadius: 6,
    colorPrimary: '#1677ff',
  },
};

const loginForm = reactive({
  username: 'admin',
  password: 'admin',
  remember: true,
});
const loginLoading = ref(false);
const collapsed = ref(false);
const selectedKeys = ref(['dashboard']);
const activePage = ref('dashboard');
const quickKeyword = ref('');
const noticeCount = ref(3);
const session = reactive<{
  token: string;
  user: UserInfo | typeof demoUser | null;
}>({
  token: getStoredToken(),
  user: null,
});

const menuItems = [
  {
    key: 'dashboard',
    icon: () => h(DashboardOutlined),
    label: '工作台',
  },
  {
    key: 'components',
    icon: () => h(AppstoreOutlined),
    label: '组件工作台',
  },
  {
    key: 'settings',
    icon: () => h(SettingOutlined),
    label: '系统设置',
  },
];

const titleMap: Record<string, string> = {
  dashboard: '工作台',
  components: '组件工作台',
  settings: '系统设置',
  'not-found': '页面不存在',
};
const pagePaths: Record<string, string> = {
  dashboard: '/dashboard',
  components: '/components',
  settings: '/settings',
};
const pathPages: Record<string, string> = Object.fromEntries(
  Object.entries(pagePaths).map(([key, value]) => [value, key]),
);

const currentTitle = computed(() => titleMap[activePage.value]);
const displayName = computed(() => session.user?.realName || session.user?.username || '管理员');
const userInitial = computed(() => displayName.value.slice(0, 1).toUpperCase());

async function handleLogin() {
  loginLoading.value = true;
  try {
    const data = await login(loginForm.username, loginForm.password);
    const token = data.accessToken || data.token;
    if (!token) {
      throw new Error('登录接口未返回 token');
    }
    persistToken(token);
    await hydrateUser();
    message.success('登录成功');
  } catch (error) {
    message.error(error instanceof Error ? error.message : '登录失败');
  } finally {
    loginLoading.value = false;
  }
}

function enterDemo() {
  persistToken('demo-token');
  session.user = demoUser;
  syncRouteFromHash();
  message.success('已进入演示模式');
}

function persistToken(token: string) {
  session.token = token;
  if (loginForm.remember || token === 'demo-token') {
    setStoredToken(token);
  }
}

async function hydrateUser() {
  if (session.token === 'demo-token') {
    session.user = demoUser;
    return;
  }
  session.user = await getUserInfo();
  void getUserMenu().catch(() => undefined);
}

function handleMenuClick(event: { key: string }) {
  navigateTo(event.key);
}

function handleQuickSearch(value: string) {
  if (!value) {
    return;
  }
  navigateTo('components');
  message.info(`已定位：${value}`);
}

async function handleUserMenu(event: { key: string }) {
  if (event.key === 'logout') {
    if (session.token && session.token !== 'demo-token') {
      await logout().catch(() => undefined);
    }
    removeStoredToken();
    session.token = '';
    session.user = null;
    selectedKeys.value = ['dashboard'];
    activePage.value = 'dashboard';
    window.location.hash = pagePaths.dashboard;
    message.success('已退出');
    return;
  }
  message.info('个人信息');
}

function navigateTo(page: string) {
  const path = pagePaths[page] || pagePaths.dashboard;
  if (window.location.hash === `#${path}`) {
    setActivePage(page);
    return;
  }
  window.location.hash = path;
}

function setActivePage(page: string) {
  activePage.value = page;
  selectedKeys.value = pagePaths[page] ? [page] : [];
}

function syncRouteFromHash() {
  const path = window.location.hash.replace(/^#/, '') || pagePaths.dashboard;
  const page = pathPages[path];
  if (page) {
    setActivePage(page);
    return;
  }
  setActivePage('not-found');
}

onMounted(async () => {
  window.addEventListener('hashchange', syncRouteFromHash);
  syncRouteFromHash();

  if (!session.token) {
    return;
  }
  try {
    await hydrateUser();
  } catch {
    removeStoredToken();
    session.token = '';
  }
});

onBeforeUnmount(() => {
  window.removeEventListener('hashchange', syncRouteFromHash);
});
</script>
