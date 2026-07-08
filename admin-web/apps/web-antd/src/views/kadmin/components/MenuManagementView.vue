<template>
  <div class="page-stack">
    <section class="page-heading">
      <div>
        <h1>菜单管理</h1>
        <p>配置 goadmin_menu 菜单结构，为后台导航和后续代码生成提供基础元数据</p>
      </div>
      <a-space wrap>
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
        <span class="muted-text">
          修改后刷新菜单或重新登录即可看到最新导航；排序使用 goadmin_menu.order。
        </span>
        <a-segmented v-model:value="density" :options="['默认', '紧凑']" />
      </div>

      <a-table
        row-key="id"
        :columns="columns"
        :data-source="filteredMenus"
        :default-expand-all-rows="true"
        :loading="loading"
        :pagination="false"
        :scroll="{ x: 1120 }"
        :size="density === '紧凑' ? 'small' : 'middle'"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'title'">
            <a-space>
              <span class="menu-icon">{{ record.icon || '-' }}</span>
              <div class="name-cell">
                <strong>{{ record.title }}</strong>
                <span>ID: {{ record.id }} · Parent: {{ record.parentId }}</span>
              </div>
            </a-space>
          </template>

          <template v-else-if="column.key === 'uri'">
            <span class="code-cell">{{ record.uri || '-' }}</span>
          </template>

          <template v-else-if="column.key === 'type'">
            <a-tag :color="record.type === 1 ? 'blue' : 'default'">
              {{ typeText(record.type) }}
            </a-tag>
          </template>

          <template v-else-if="column.key === 'action'">
            <a-space>
              <a-button type="link" size="small" @click="openDrawer(undefined, record)">
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
          <a-input v-model:value="formState.title" placeholder="例如：菜单管理" />
        </a-form-item>
        <a-form-item label="访问路径" name="uri">
          <a-input v-model:value="formState.uri" placeholder="例如：/kadmin/menus" />
        </a-form-item>
        <a-row :gutter="12">
          <a-col :span="12">
            <a-form-item label="排序" name="order">
              <a-input-number v-model:value="formState.order" class="full-width" :min="0" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="类型" name="type">
              <a-select v-model:value="formState.type" :options="typeOptions" />
            </a-form-item>
          </a-col>
        </a-row>
        <a-form-item label="图标" name="icon">
          <a-input v-model:value="formState.icon" placeholder="例如：lucide:menu" />
        </a-form-item>
        <a-form-item label="分组标题" name="header">
          <a-input v-model:value="formState.header" placeholder="可选，对应 header 字段" />
        </a-form-item>
        <a-row :gutter="12">
          <a-col :span="12">
            <a-form-item label="插件名" name="pluginName">
              <a-input v-model:value="formState.pluginName" placeholder="可选" />
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
          <a-button type="primary" :loading="saving" @click="submitForm">保存</a-button>
        </a-space>
      </template>
    </a-drawer>
  </div>
</template>

<script setup lang="ts">
import {
  ClearOutlined,
  PlusOutlined,
  ReloadOutlined,
  SearchOutlined,
} from '@ant-design/icons-vue';
import { message, type FormInstance } from 'ant-design-vue';
import { computed, onMounted, reactive, ref } from 'vue';

import {
  createAdminMenu,
  deleteAdminMenu,
  getAdminMenuTree,
  updateAdminMenu,
  type AdminMenu,
  type AdminMenuPayload,
} from '#/api/kadmin/menus';

const loading = ref(false);
const saving = ref(false);
const drawerOpen = ref(false);
const density = ref('默认');
const menus = ref<AdminMenu[]>([]);
const editingMenu = ref<AdminMenu | null>(null);
const formRef = ref<FormInstance>();

const filters = reactive<{
  keyword: string;
  type?: number;
}>({
  keyword: '',
  type: undefined,
});

const formState = reactive<AdminMenuPayload>({
  parentId: 0,
  type: 1,
  order: 0,
  title: '',
  icon: '',
  uri: '',
  header: '',
  pluginName: '',
  uuid: '',
});

const rules = {
  title: [{ required: true, message: '请输入菜单标题' }],
};

const typeOptions = [
  { label: '菜单', value: 1 },
  { label: '目录/分组', value: 0 },
];

const columns = [
  { title: '菜单', key: 'title', width: 300, fixed: 'left' },
  { title: '路径', key: 'uri', dataIndex: 'uri', width: 240 },
  { title: '类型', key: 'type', dataIndex: 'type', width: 100 },
  { title: '排序', key: 'order', dataIndex: 'order', width: 90 },
  { title: 'Header', key: 'header', dataIndex: 'header', width: 150 },
  { title: '插件', key: 'pluginName', dataIndex: 'pluginName', width: 130 },
  { title: '更新时间', key: 'updatedAt', dataIndex: 'updatedAt', width: 170 },
  { title: '操作', key: 'action', width: 210, fixed: 'right' },
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
  excludeMenuSubtree(menus.value, editingMenu.value?.id),
);

onMounted(() => {
  void loadMenus();
});

async function loadMenus() {
  loading.value = true;
  try {
    menus.value = await getAdminMenuTree();
  } catch (error) {
    message.error(error instanceof Error ? error.message : '加载菜单失败');
  } finally {
    loading.value = false;
  }
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
    type: record?.type ?? 1,
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
    if (editingMenu.value) {
      await updateAdminMenu(editingMenu.value.id, payload);
    } else {
      await createAdminMenu(payload);
    }
    drawerOpen.value = false;
    await loadMenus();
    message.success('菜单已保存');
  } catch (error) {
    message.error(error instanceof Error ? error.message : '保存菜单失败');
  } finally {
    saving.value = false;
  }
}

async function removeMenu(record: AdminMenu) {
  try {
    await deleteAdminMenu(record.id);
    await loadMenus();
    message.success('菜单已删除');
  } catch (error) {
    message.error(error instanceof Error ? error.message : '删除菜单失败');
  }
}

function normalizePayload(): AdminMenuPayload {
  return {
    parentId: Number(formState.parentId) || 0,
    type: Number(formState.type) || 0,
    order: Number(formState.order) || 0,
    title: formState.title.trim(),
    icon: formState.icon?.trim(),
    uri: formState.uri?.trim(),
    header: formState.header?.trim(),
    pluginName: formState.pluginName?.trim(),
    uuid: formState.uuid?.trim(),
  };
}

function filterMenuTree(
  items: AdminMenu[],
  predicate: (menu: AdminMenu) => boolean,
): AdminMenu[] {
  const result: AdminMenu[] = [];
  for (const item of items) {
    const children = item.children ? filterMenuTree(item.children, predicate) : [];
    if (predicate(item) || children.length > 0) {
      result.push({
        ...item,
        children,
      });
    }
  }
  return result;
}

function excludeMenuSubtree(items: AdminMenu[], excludeId?: number): AdminMenu[] {
  if (!excludeId) {
    return items;
  }
  const result: AdminMenu[] = [];
  for (const item of items) {
    if (item.id === excludeId) {
      continue;
    }
    const children = item.children
      ? excludeMenuSubtree(item.children, excludeId)
      : [];
    result.push({ ...item, children });
  }
  return result;
}

function nextChildOrder(parent?: AdminMenu) {
  const children = parent?.children || menus.value;
  return children.reduce((max, item) => Math.max(max, item.order), 0) + 1;
}

function typeText(type: number) {
  return type === 1 ? '菜单' : '目录/分组';
}
</script>
