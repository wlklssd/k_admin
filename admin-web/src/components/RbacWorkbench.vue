<template>
  <div class="page-stack">
    <section class="page-heading">
      <div>
        <h1>权限管理</h1>
        <p>检查权限组、菜单授权和组内员工分配</p>
      </div>
      <a-space wrap>
        <a-button @click="loadOverview">
          <ReloadOutlined />
          刷新
        </a-button>
      </a-space>
    </section>

    <a-row :gutter="[16, 16]">
      <a-col :xs="24" :lg="9">
        <section class="panel">
          <div class="panel-title">
            <h2>权限组</h2>
            <a-button size="small" type="primary" @click="openRoleDrawer()">
              <PlusOutlined />
              新建权限组
            </a-button>
          </div>
          <a-input
            v-model:value="roleKeyword"
            allow-clear
            class="toolbar-input"
            placeholder="搜索权限组"
          >
            <template #prefix>
              <SearchOutlined />
            </template>
          </a-input>
          <a-list
            class="role-list"
            :data-source="filteredRoles"
            :loading="loading"
            item-layout="horizontal"
          >
            <template #renderItem="{ item }">
              <a-list-item
                :class="['role-list-item', { active: item.id === selectedRoleId }]"
                @click="selectRole(item.id)"
              >
                <a-list-item-meta>
                  <template #avatar>
                    <a-avatar :style="{ backgroundColor: item.isAdmin ? '#d46b08' : '#1677ff' }">
                      {{ item.name.slice(0, 1) }}
                    </a-avatar>
                  </template>
                  <template #title>
                    <a-space :size="6">
                      <span>{{ item.name }}</span>
                      <a-tag v-if="item.isAdmin" color="gold">最高权限</a-tag>
                    </a-space>
                  </template>
                  <template #description>
                    {{ item.slug }} · {{ item.userIds.length }} 人 · {{ item.menuIds.length }} 菜单
                  </template>
                </a-list-item-meta>
              </a-list-item>
            </template>
          </a-list>
        </section>
      </a-col>

      <a-col :xs="24" :lg="15">
        <section class="panel">
          <a-empty v-if="!selectedRole" description="请选择权限组" />
          <template v-else>
            <div class="panel-title">
              <div>
                <h2>{{ selectedRole.name }}</h2>
                <p class="muted-text">{{ selectedRole.slug }}</p>
              </div>
              <a-space wrap>
                <a-button :disabled="selectedRole.isAdmin" @click="openRoleDrawer(selectedRole)">
                  <EditOutlined />
                  改名
                </a-button>
                <a-popconfirm title="确认删除该权限组？" @confirm="removeRole">
                  <a-button danger :disabled="selectedRole.isAdmin">
                    <DeleteOutlined />
                    删除
                  </a-button>
                </a-popconfirm>
              </a-space>
            </div>

            <a-alert
              v-if="selectedRole.isAdmin"
              class="form-alert"
              type="info"
              show-icon
              message="admin 最高权限用户不受菜单和权限限制，管理员组仅展示当前系统权限全集。"
            />

            <a-tabs v-model:active-key="activeTab" class="rbac-tabs">
              <a-tab-pane key="menus" tab="菜单权限">
                <div class="table-toolbar">
                  <span class="muted-text">配置该权限组可见和可访问的后台菜单</span>
                  <a-button
                    type="primary"
                    :disabled="selectedRole.isAdmin"
                    :loading="savingMenus"
                    @click="saveMenus"
                  >
                    保存菜单权限
                  </a-button>
                </div>
                <a-tree
                  v-model:checkedKeys="checkedMenuKeys"
                  checkable
                  default-expand-all
                  :disabled="selectedRole.isAdmin"
                  :tree-data="menuTreeData"
                />
              </a-tab-pane>

              <a-tab-pane key="users" tab="组内员工">
                <div class="table-toolbar">
                  <span class="muted-text">调配该权限组下的人员，admin 用户固定保留最高权限</span>
                  <a-button
                    type="primary"
                    :disabled="selectedRole.isAdmin"
                    :loading="savingUsers"
                    @click="saveUsers"
                  >
                    保存员工分配
                  </a-button>
                </div>
                <a-transfer
                  v-model:target-keys="selectedUserKeys"
                  class="rbac-transfer"
                  :data-source="transferUsers"
                  :disabled="selectedRole.isAdmin"
                  :list-style="{ height: '360px' }"
                  :render="(item) => item.title"
                  :titles="['可选员工', '组内员工']"
                  show-search
                />
              </a-tab-pane>
            </a-tabs>
          </template>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>员工与权限组</h2>
            <a-tag>{{ users.length }} 人</a-tag>
          </div>
          <a-table
            row-key="id"
            size="small"
            :columns="userColumns"
            :data-source="users"
            :pagination="{ pageSize: 6, showTotal: (total: number) => `共 ${total} 人` }"
            :scroll="{ x: 720 }"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'name'">
                <a-space>
                  <a-avatar>{{ (record.name || record.username).slice(0, 1) }}</a-avatar>
                  <div class="name-cell">
                    <strong>{{ record.name || record.username }}</strong>
                    <span>{{ record.username }}</span>
                  </div>
                </a-space>
              </template>
              <template v-else-if="column.key === 'roles'">
                <a-space :size="4" wrap>
                  <a-tag v-for="role in record.roles" :key="role">{{ role }}</a-tag>
                  <a-tag v-if="record.username === 'admin'" color="gold">最高权限</a-tag>
                </a-space>
              </template>
            </template>
          </a-table>
        </section>
      </a-col>
    </a-row>

    <a-drawer
      v-model:open="roleDrawerOpen"
      :title="editingRole ? '编辑权限组' : '新建权限组'"
      width="420"
      :destroy-on-close="true"
    >
      <a-form ref="roleFormRef" :model="roleForm" :rules="roleRules" layout="vertical">
        <a-form-item label="权限组名称" name="name">
          <a-input v-model:value="roleForm.name" placeholder="例如：运营人员" />
        </a-form-item>
        <a-form-item label="权限标识" name="slug">
          <a-input v-model:value="roleForm.slug" placeholder="例如：operator" />
        </a-form-item>
      </a-form>
      <template #extra>
        <a-space>
          <a-button @click="roleDrawerOpen = false">取消</a-button>
          <a-button type="primary" :loading="savingRole" @click="submitRole">保存</a-button>
        </a-space>
      </template>
    </a-drawer>
  </div>
</template>

<script setup lang="ts">
import {
  DeleteOutlined,
  EditOutlined,
  PlusOutlined,
  ReloadOutlined,
  SearchOutlined,
} from '@ant-design/icons-vue';
import { message, type FormInstance } from 'ant-design-vue';
import { computed, onMounted, reactive, ref, watch } from 'vue';

import {
  createRole,
  deleteRole,
  getRbacOverview,
  updateRole,
  updateRoleMenus,
  updateRoleUsers,
  type RbacMenu,
  type RbacRole,
  type RbacUser,
} from '../api/rbac';

type TreeNode = {
  title: string;
  key: number;
  children?: TreeNode[];
};

const loading = ref(false);
const savingRole = ref(false);
const savingMenus = ref(false);
const savingUsers = ref(false);
const roles = ref<RbacRole[]>([]);
const menus = ref<RbacMenu[]>([]);
const users = ref<RbacUser[]>([]);
const selectedRoleId = ref<number>();
const roleKeyword = ref('');
const activeTab = ref('menus');
const checkedMenuKeys = ref<number[]>([]);
const selectedUserKeys = ref<string[]>([]);
const roleDrawerOpen = ref(false);
const editingRole = ref<RbacRole | null>(null);
const roleFormRef = ref<FormInstance>();
const roleForm = reactive({
  name: '',
  slug: '',
});

const roleRules = {
  name: [{ required: true, message: '请输入权限组名称' }],
  slug: [{ required: true, message: '请输入权限标识' }],
};

const userColumns = [
  { title: '员工', key: 'name', width: 240 },
  { title: '权限组', key: 'roles', width: 360 },
];

const filteredRoles = computed(() => {
  const keyword = roleKeyword.value.trim().toLowerCase();
  if (!keyword) return roles.value;
  return roles.value.filter(
    (role) =>
      role.name.toLowerCase().includes(keyword) ||
      role.slug.toLowerCase().includes(keyword),
  );
});

const selectedRole = computed(() =>
  roles.value.find((role) => role.id === selectedRoleId.value),
);

const menuTreeData = computed(() => toTreeData(menus.value));

const transferUsers = computed(() =>
  users.value
    .filter((user) => user.username !== 'admin')
    .map((user) => ({
      key: String(user.id),
      title: `${user.name || user.username}（${user.username}）`,
      description: user.roles.join('、'),
    })),
);

watch(selectedRole, (role) => {
  checkedMenuKeys.value = role?.menuIds ? [...role.menuIds] : [];
  selectedUserKeys.value = role?.userIds
    ? role.userIds.filter((id) => id !== 1).map(String)
    : [];
});

onMounted(() => {
  void loadOverview();
});

async function loadOverview() {
  loading.value = true;
  try {
    const data = await getRbacOverview();
    roles.value = data.roles || [];
    menus.value = data.menus || [];
    users.value = data.users || [];
    if (!selectedRoleId.value || !roles.value.some((role) => role.id === selectedRoleId.value)) {
      selectedRoleId.value = roles.value[0]?.id;
    }
  } catch (error) {
    message.error(error instanceof Error ? error.message : '加载权限数据失败');
  } finally {
    loading.value = false;
  }
}

function selectRole(roleId: number) {
  selectedRoleId.value = roleId;
}

function openRoleDrawer(role?: RbacRole) {
  editingRole.value = role || null;
  roleForm.name = role?.name || '';
  roleForm.slug = role?.slug || '';
  roleDrawerOpen.value = true;
}

async function submitRole() {
  await roleFormRef.value?.validate();
  savingRole.value = true;
  try {
    const payload = {
      name: roleForm.name.trim(),
      slug: roleForm.slug.trim(),
    };
    const role = editingRole.value
      ? await updateRole(editingRole.value.id, payload)
      : await createRole(payload);
    await loadOverview();
    selectedRoleId.value = role.id;
    roleDrawerOpen.value = false;
    message.success('权限组已保存');
  } catch (error) {
    message.error(error instanceof Error ? error.message : '保存权限组失败');
  } finally {
    savingRole.value = false;
  }
}

async function removeRole() {
  if (!selectedRole.value || selectedRole.value.isAdmin) return;
  try {
    await deleteRole(selectedRole.value.id);
    selectedRoleId.value = undefined;
    await loadOverview();
    message.success('权限组已删除');
  } catch (error) {
    message.error(error instanceof Error ? error.message : '删除权限组失败');
  }
}

async function saveMenus() {
  if (!selectedRole.value || selectedRole.value.isAdmin) return;
  savingMenus.value = true;
  try {
    await updateRoleMenus(selectedRole.value.id, checkedMenuKeys.value.map(Number));
    await loadOverview();
    message.success('菜单权限已保存');
  } catch (error) {
    message.error(error instanceof Error ? error.message : '保存菜单权限失败');
  } finally {
    savingMenus.value = false;
  }
}

async function saveUsers() {
  if (!selectedRole.value || selectedRole.value.isAdmin) return;
  savingUsers.value = true;
  try {
    await updateRoleUsers(selectedRole.value.id, selectedUserKeys.value.map(Number));
    await loadOverview();
    message.success('员工分配已保存');
  } catch (error) {
    message.error(error instanceof Error ? error.message : '保存员工分配失败');
  } finally {
    savingUsers.value = false;
  }
}

function toTreeData(items: RbacMenu[]): TreeNode[] {
  return items.map((item) => ({
    title: item.uri ? `${item.title} (${item.uri})` : item.title,
    key: item.id,
    children: item.children ? toTreeData(item.children) : undefined,
  }));
}
</script>
