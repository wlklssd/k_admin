<template>
  <div class="page-stack">
    <section class="page-heading">
      <div>
        <h1>菜单管理</h1>
        <p>维护后台导航菜单的层级、路径与图标</p>
      </div>
      <a-space wrap>
        <a-button :disabled="loading" @click="layoutEditorOpen = true">
          <ApartmentOutlined />
          菜单布局
        </a-button>
        <a-button :loading="loading" @click="loadMenus">
          <ReloadOutlined />
          刷新
        </a-button>
        <a-button type="primary" @click="openDrawer()">
          <PlusOutlined />
          新增菜单
        </a-button>
      </a-space>
    </section>

    <section class="panel">
      <a-form :model="filters" layout="inline" class="search-form">
        <a-form-item label="关键词">
          <a-input
            v-model:value="filters.keyword"
            allow-clear
            class="control-lg"
            placeholder="菜单标题 / 路由 / 图标"
          />
        </a-form-item>
        <a-form-item label="类型">
          <a-select
            v-model:value="filters.type"
            allow-clear
            class="control-md"
            :options="typeOptions"
            placeholder="全部"
          />
        </a-form-item>
        <a-form-item>
          <a-space>
            <a-button type="primary" @click="applySearch">
              <SearchOutlined />
              查询
            </a-button>
            <a-button @click="resetSearch">
              <ClearOutlined />
              重置
            </a-button>
          </a-space>
        </a-form-item>
      </a-form>
    </section>

    <section class="panel">
      <div class="table-toolbar">
        <span class="muted-text"> 当前菜单树按导航展示顺序排列。 </span>
        <a-segmented v-model:value="density" :options="['默认', '紧凑']" />
      </div>

      <a-table
        row-key="id"
        :columns="columns"
        :data-source="filteredMenus"
        :expanded-row-keys="expandedRowKeys"
        :loading="loading"
        :pagination="false"
        :row-class-name="menuRowClassName"
        :scroll="{ x: 1040 }"
        :size="density === '紧凑' ? 'small' : 'middle'"
        @expanded-rows-change="onExpandedRowsChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'title'">
            <div class="menu-title-cell">
              <span class="menu-icon-box" :title="record.icon || '未设置图标'">
                <IconifyIcon v-if="record.icon" :icon="record.icon" />
                <span v-else class="menu-icon-fallback">—</span>
              </span>
              <div class="name-cell">
                <strong>{{ record.title }}</strong>
                <span>#{{ record.id }}</span>
              </div>
            </div>
          </template>

          <template v-else-if="column.key === 'uri'">
            <span v-if="record.uri" class="code-cell">{{ record.uri }}</span>
            <span v-else class="muted-text">—</span>
          </template>

          <template v-else-if="column.key === 'type'">
            <a-tag
              :color="record.type === ADMIN_MENU_TYPE.MENU ? 'blue' : 'default'"
            >
              {{ typeText(record.type) }}
            </a-tag>
          </template>

          <template v-else-if="column.key === 'action'">
            <a-space>
              <a-button
                v-if="record.type === ADMIN_MENU_TYPE.DIRECTORY"
                type="link"
                size="small"
                @click="openDrawer(undefined, record)"
              >
                新增子级
              </a-button>
              <a-button type="link" size="small" @click="openDrawer(record)">
                编辑
              </a-button>
              <a-popconfirm
                title="确认删除该菜单？存在子菜单时将无法删除，请先删除子菜单。"
                @confirm="removeMenu(record)"
              >
                <a-button type="link" size="small" danger>删除</a-button>
              </a-popconfirm>
            </a-space>
          </template>
        </template>
      </a-table>
    </section>

    <a-drawer
      v-model:open="drawerOpen"
      :destroy-on-close="true"
      :title="editingMenu ? '编辑菜单' : '新增菜单'"
      width="560"
    >
      <a-form ref="formRef" :model="formState" :rules="rules" layout="vertical">
        <a-form-item label="上级菜单" name="parentId">
          <a-tree-select
            v-model:value="formState.parentId"
            allow-clear
            tree-default-expand-all
            :field-names="{ children: 'children', label: 'title', value: 'id' }"
            :tree-data="parentOptions"
            placeholder="根菜单"
          />
        </a-form-item>
        <a-form-item label="菜单标题" name="title">
          <a-input
            v-model:value="formState.title"
            placeholder="例如：菜单管理"
          />
        </a-form-item>
        <a-form-item label="访问路径" name="uri">
          <a-input
            v-model:value="formState.uri"
            placeholder="例如：/kadmin/menus"
          />
        </a-form-item>
        <a-form-item label="类型" name="type">
          <a-select v-model:value="formState.type" :options="formTypeOptions" />
        </a-form-item>
        <a-form-item label="图标" name="icon">
          <a-input
            v-model:value="formState.icon"
            placeholder="例如：lucide:menu"
          />
        </a-form-item>
        <a-form-item label="分组标题" name="header">
          <a-input
            v-model:value="formState.header"
            placeholder="可选，对应 header 字段"
          />
        </a-form-item>
        <a-row :gutter="12">
          <a-col :span="12">
            <a-form-item label="插件名" name="pluginName">
              <a-input
                v-model:value="formState.pluginName"
                placeholder="可选"
              />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="UUID" name="uuid">
              <a-input v-model:value="formState.uuid" placeholder="可选" />
            </a-form-item>
          </a-col>
        </a-row>
      </a-form>
      <template #extra>
        <a-space>
          <a-button @click="drawerOpen = false">取消</a-button>
          <a-button type="primary" :loading="saving" @click="submitForm">
            保存
          </a-button>
        </a-space>
      </template>
    </a-drawer>

    <MenuSorter
      v-model:open="layoutEditorOpen"
      :menus="menus"
      :saving="layoutSaving"
      @save="saveMenuLayout"
    />
  </div>
</template>

<script setup lang="ts">
import type { FormInstance } from 'ant-design-vue';

import type {
  AdminMenu,
  AdminMenuPayload,
  AdminMenuPosition,
  AdminMenuType,
} from '#/api/kadmin/menus';

import { computed, onMounted, reactive, ref } from 'vue';

import { IconifyIcon } from '@vben/icons';
import { useAccessStore, useUserStore } from '@vben/stores';

import {
  ApartmentOutlined,
  ClearOutlined,
  PlusOutlined,
  ReloadOutlined,
  SearchOutlined,
} from '@ant-design/icons-vue';
import { message } from 'ant-design-vue';

import {
  ADMIN_MENU_TYPE,
  createAdminMenu,
  deleteAdminMenu,
  getAdminMenuTree,
  updateAdminMenu,
  updateAdminMenuLayout,
} from '#/api/kadmin/menus';
import { router } from '#/router';
import { generateAccess } from '#/router/access';
import { accessRoutes } from '#/router/routes';

import {
  canSetMenuAsItem,
  filterMenuParentOptions,
} from './menu-sort';
import MenuSorter from './MenuSorter.vue';

const loading = ref(false);
const saving = ref(false);
const drawerOpen = ref(false);
const layoutEditorOpen = ref(false);
const layoutSaving = ref(false);
const density = ref('默认');
const menus = ref<AdminMenu[]>([]);
const editingMenu = ref<AdminMenu | null>(null);
const formRef = ref<FormInstance>();
const expandedRowKeys = ref<number[]>([]);
const accessStore = useAccessStore();
const userStore = useUserStore();

const filters = reactive<{
  keyword: string;
  type?: AdminMenuType;
}>({
  keyword: '',
  type: undefined,
});

const formState = reactive<AdminMenuPayload>({
  parentId: 0,
  type: ADMIN_MENU_TYPE.MENU,
  order: 0,
  title: '',
  icon: '',
  uri: '',
  header: '',
  pluginName: '',
  uuid: '',
});

const rules = {
  parentId: [{ validator: validateParentMenu }],
  title: [{ required: true, message: '请输入菜单标题' }],
  type: [{ validator: validateMenuTypeChange }],
};

const typeOptions = [
  { label: '菜单', value: ADMIN_MENU_TYPE.MENU },
  { label: '目录/分组', value: ADMIN_MENU_TYPE.DIRECTORY },
];

const formTypeOptions = computed(() =>
  typeOptions.map((option) => ({
    ...option,
    disabled:
      option.value === ADMIN_MENU_TYPE.MENU &&
      !canSetMenuAsItem(editingMenu.value),
  })),
);

const columns = [
  { title: '菜单', key: 'title', width: 260, fixed: 'left' },
  { title: '路径', key: 'uri', dataIndex: 'uri', width: 220 },
  { title: '类型', key: 'type', dataIndex: 'type', width: 96 },
  { title: '更新时间', key: 'updatedAt', dataIndex: 'updatedAt', width: 168 },
  { title: '操作', key: 'action', width: 196, fixed: 'right' },
];

const filteredMenus = computed(() => {
  const keyword = filters.keyword.trim().toLowerCase();
  return filterMenuTree(menus.value, (menu) => {
    const matchKeyword =
      !keyword ||
      menu.title.toLowerCase().includes(keyword) ||
      menu.uri.toLowerCase().includes(keyword) ||
      menu.icon.toLowerCase().includes(keyword);
    const matchType = filters.type === undefined || menu.type === filters.type;
    return matchKeyword && matchType;
  });
});

const parentOptions = computed(() =>
  filterMenuParentOptions(menus.value, editingMenu.value?.id),
);

onMounted(() => {
  void loadMenus();
});

async function loadMenus() {
  loading.value = true;
  try {
    menus.value = await getAdminMenuTree();
    // a-table 的 defaultExpandAllRows 仅在挂载时生效，此时数据尚未返回；
    // 改为受控展开：数据到达后自动展开所有含子菜单的节点，保证子菜单可见。
    expandedRowKeys.value = collectExpandableKeys(menus.value);
  } catch (error) {
    message.error(error instanceof Error ? error.message : '加载菜单失败');
  } finally {
    loading.value = false;
  }
}

function onExpandedRowsChange(keys: (number | string)[]) {
  expandedRowKeys.value = keys as number[];
}

function applySearch() {
  // 菜单树为前端实时过滤，输入即生效，此处仅保留入口与其他列表页保持一致。
}

function resetSearch() {
  filters.keyword = '';
  filters.type = undefined;
}

function openDrawer(record?: AdminMenu, parent?: AdminMenu) {
  editingMenu.value = record || null;
  Object.assign(formState, {
    parentId: record?.parentId ?? parent?.id ?? 0,
    type: record?.type ?? ADMIN_MENU_TYPE.MENU,
    order: record?.order ?? nextChildOrder(parent),
    title: record?.title ?? '',
    icon: record?.icon ?? '',
    uri: record?.uri ?? '',
    header: record?.header ?? '',
    pluginName: record?.pluginName ?? '',
    uuid: record?.uuid ?? '',
  });
  drawerOpen.value = true;
}

async function submitForm() {
  await formRef.value?.validate();
  saving.value = true;
  try {
    const payload = normalizePayload();
    await (editingMenu.value
      ? updateAdminMenu(editingMenu.value.id, payload)
      : createAdminMenu(payload));
    drawerOpen.value = false;
    message.success('菜单已保存');
    await loadMenus();
    await refreshNavigationMenusSafely();
  } catch (error) {
    message.error(error instanceof Error ? error.message : '保存菜单失败');
  } finally {
    saving.value = false;
  }
}

async function removeMenu(record: AdminMenu) {
  try {
    await deleteAdminMenu(record.id);
    message.success('菜单已删除');
    await loadMenus();
    await refreshNavigationMenusSafely();
  } catch (error) {
    message.error(error instanceof Error ? error.message : '删除菜单失败');
  }
}

async function saveMenuLayout(items: AdminMenuPosition[]) {
  layoutSaving.value = true;
  try {
    menus.value = await updateAdminMenuLayout(items);
    expandedRowKeys.value = collectExpandableKeys(menus.value);
    layoutEditorOpen.value = false;
    message.success('菜单布局已保存');
    await refreshNavigationMenusSafely();
  } catch (error) {
    message.error(error instanceof Error ? error.message : '保存菜单布局失败');
  } finally {
    layoutSaving.value = false;
  }
}

async function refreshNavigationMenusSafely() {
  try {
    await refreshNavigationMenus();
  } catch {
    message.warning('菜单数据已更新，导航刷新失败，请刷新页面后查看最新布局');
  }
}

async function refreshNavigationMenus() {
  const { accessibleMenus, accessibleRoutes } = await generateAccess({
    roles: userStore.userInfo?.roles ?? [],
    router,
    routes: accessRoutes,
  });
  accessStore.setAccessMenus(accessibleMenus);
  accessStore.setAccessRoutes(accessibleRoutes);
  accessStore.setIsAccessChecked(true);
  const currentRoute = router.currentRoute.value;
  await router.replace({
    force: true,
    hash: currentRoute.hash,
    path: currentRoute.path,
    query: currentRoute.query,
  });
}

function normalizePayload(): AdminMenuPayload {
  const parentId = Number(formState.parentId) || 0;
  const parentChanged =
    editingMenu.value !== null && editingMenu.value.parentId !== parentId;
  return {
    parentId,
    type: formState.type ?? ADMIN_MENU_TYPE.MENU,
    order: parentChanged
      ? nextChildOrder(findMenuById(menus.value, parentId))
      : Number(formState.order) || 0,
    title: formState.title.trim(),
    icon: formState.icon?.trim(),
    uri: formState.uri?.trim(),
    header: formState.header?.trim(),
    pluginName: formState.pluginName?.trim(),
    uuid: formState.uuid?.trim(),
  };
}

function validateParentMenu(_rule: unknown, value: unknown): Promise<void> {
  const parentId = Number(value) || 0;
  if (parentId === 0) {
    return Promise.resolve();
  }
  const parent = findMenuById(menus.value, parentId);
  if (!parent || parent.type !== ADMIN_MENU_TYPE.DIRECTORY) {
    return Promise.reject(new Error('上级菜单必须是目录/分组'));
  }
  return Promise.resolve();
}

function validateMenuTypeChange(
  _rule: unknown,
  value: unknown,
): Promise<void> {
  if (
    Number(value) === ADMIN_MENU_TYPE.MENU &&
    !canSetMenuAsItem(editingMenu.value)
  ) {
    return Promise.reject(new Error('存在子菜单的节点必须保持为目录/分组'));
  }
  return Promise.resolve();
}

function findMenuById(items: AdminMenu[], id: number): AdminMenu | undefined {
  if (id === 0) {
    return undefined;
  }
  for (const item of items) {
    if (item.id === id) {
      return item;
    }
    const child = findMenuById(item.children ?? [], id);
    if (child) {
      return child;
    }
  }
}

function filterMenuTree(
  items: AdminMenu[],
  predicate: (menu: AdminMenu) => boolean,
): AdminMenu[] {
  const result: AdminMenu[] = [];
  for (const item of items) {
    const children = item.children
      ? filterMenuTree(item.children, predicate)
      : [];
    if (predicate(item) || children.length > 0) {
      result.push({
        ...item,
        children,
      });
    }
  }
  return result;
}

function collectExpandableKeys(items: AdminMenu[]): number[] {
  const keys: number[] = [];
  for (const item of items) {
    if (item.children && item.children.length > 0) {
      keys.push(item.id, ...collectExpandableKeys(item.children));
    }
  }
  return keys;
}

function nextChildOrder(parent?: AdminMenu) {
  const children = parent?.children || menus.value;
  let maxOrder = 0;
  for (const item of children) {
    maxOrder = Math.max(maxOrder, item.order);
  }
  return maxOrder + 1;
}

function menuRowClassName(record: AdminMenu) {
  if (!record.parentId) {
    return 'menu-row menu-row-root';
  }

  const depth = getMenuDepth(record.id, filteredMenus.value) ?? 1;
  return `menu-row menu-row-child menu-row-level-${Math.min(depth, 3)}`;
}

function typeText(type: AdminMenuType) {
  return type === ADMIN_MENU_TYPE.MENU ? '菜单' : '目录/分组';
}

function getMenuDepth(
  id: number,
  items: AdminMenu[],
  depth = 0,
): number | undefined {
  for (const item of items) {
    if (item.id === id) {
      return depth;
    }

    if (item.children?.length) {
      const childDepth = getMenuDepth(id, item.children, depth + 1);
      if (childDepth !== undefined) {
        return childDepth;
      }
    }
  }
}
</script>

<style scoped>
:deep(.menu-row > td) {
  transition:
    background-color 0.24s ease,
    box-shadow 0.24s ease;
}

:deep(.menu-row-child) {
  animation: menu-child-slide-in 0.24s ease both;
}

:deep(.menu-row-child > td) {
  background-color: #f1f5f9;
}

:deep(.menu-row-level-2 > td) {
  background-color: #e8eef7;
}

:deep(.menu-row-level-3 > td) {
  background-color: #dee8f3;
}

:deep(.menu-row-child > td:first-child) {
  box-shadow: inset 3px 0 0 #8aa4c8;
}

:deep(.menu-row-child:hover > td) {
  background-color: #dfeaf7 !important;
}

@keyframes menu-child-slide-in {
  from {
    opacity: 0;
    transform: translateY(-8px);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
