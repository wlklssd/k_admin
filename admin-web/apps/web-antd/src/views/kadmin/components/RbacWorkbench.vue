<template>
  <div class="page-stack">
    <section class="page-heading">
      <div>
        <h1>权限管理</h1>
        <p>按部门配置职位，职位的人员分配在用户管理中维护</p>
      </div>
      <a-space wrap>
        <a-button @click="loadOverview">
          <ReloadOutlined />
          刷新
        </a-button>
      </a-space>
    </section>

    <a-row :gutter="[16, 16]">
      <a-col :xs="24" :lg="8">
        <section class="panel">
          <div class="panel-title">
            <h2>部门</h2>
            <a-button size="small" type="primary" @click="openDepartmentDrawer()">
              <PlusOutlined />
              新建部门
            </a-button>
          </div>
          <a-input
            v-model:value="departmentKeyword"
            allow-clear
            class="toolbar-input"
            placeholder="搜索部门"
          >
            <template #prefix>
              <SearchOutlined />
            </template>
          </a-input>
          <a-list
            class="role-list"
            :data-source="filteredDepartments"
            :loading="loading"
            item-layout="horizontal"
          >
            <template #renderItem="{ item }">
              <a-list-item
                :class="['role-list-item', { active: item.id === selectedDepartmentId }]"
                @click="selectDepartment(item.id)"
              >
                <a-list-item-meta>
                  <template #avatar>
                    <a-avatar :style="{ backgroundColor: item.status === 1 ? '#1677ff' : '#8c8c8c' }">
                      {{ item.name.slice(0, 1) }}
                    </a-avatar>
                  </template>
                  <template #title>
                    <a-space :size="6">
                      <span>{{ item.name }}</span>
                      <a-tag v-if="item.status !== 1">停用</a-tag>
                    </a-space>
                  </template>
                  <template #description>
                    {{ item.code || '未设置编码' }} · {{ item.roleIds.length }} 个职位
                  </template>
                </a-list-item-meta>
              </a-list-item>
            </template>
          </a-list>
        </section>
      </a-col>

      <a-col :xs="24" :lg="16">
        <section class="panel">
          <a-empty v-if="!selectedDepartment" description="请选择部门" />
          <template v-else>
            <div class="panel-title">
              <div>
                <h2>{{ selectedDepartment.name }}</h2>
                <p class="muted-text">
                  {{ selectedDepartment.code || '未设置编码' }}
                  <template v-if="selectedDepartment.description">
                    · {{ selectedDepartment.description }}
                  </template>
                </p>
              </div>
              <a-space wrap>
                <a-button @click="openDepartmentDrawer(selectedDepartment)">
                  <EditOutlined />
                  编辑部门
                </a-button>
                <a-popconfirm title="确认删除该部门？" @confirm="removeDepartment">
                  <a-button danger>
                    <DeleteOutlined />
                    删除
                  </a-button>
                </a-popconfirm>
              </a-space>
            </div>

            <a-tabs v-model:active-key="activeTab" class="rbac-tabs">
              <a-tab-pane key="departmentRoles" tab="部门职位">
                <div class="table-toolbar">
                  <span class="muted-text">配置该部门下包含哪些职位，职位即现有权限组</span>
                  <a-space>
                    <a-button type="primary" :loading="savingDepartmentRoles" @click="saveDepartmentRoles">
                      保存部门职位
                    </a-button>
                  </a-space>
                </div>
                <a-transfer
                  v-model:target-keys="selectedDepartmentRoleKeys"
                  class="rbac-transfer"
                  :data-source="transferRoles"
                  :list-style="{ height: '360px' }"
                  :render="renderTransferItem"
                  :titles="['可选职位', '部门职位']"
                  show-search
                />
              </a-tab-pane>

              <a-tab-pane key="roleMenus" tab="职位菜单权限">
                <div class="table-toolbar">
                  <a-select
                    v-model:value="selectedRoleId"
                    class="control-lg"
                    :options="departmentRoleOptions"
                    placeholder="选择部门下的职位"
                  />
                  <a-button
                    type="primary"
                    :disabled="!selectedRole || selectedRole.isAdmin"
                    :loading="savingMenus"
                    @click="saveMenus"
                  >
                    保存菜单权限
                  </a-button>
                </div>
                <a-empty v-if="!selectedRole" description="请先在部门职位中选择或添加职位" />
                <template v-else>
                  <a-alert
                    v-if="selectedRole.isAdmin"
                    class="form-alert"
                    type="info"
                    show-icon
                    message="admin 最高权限职位不受菜单限制，仅展示当前系统权限全集。"
                  />
                  <a-tree
                    v-model:checkedKeys="checkedMenuKeys"
                    checkable
                    default-expand-all
                    :disabled="selectedRole.isAdmin"
                    :tree-data="menuTreeData"
                  />
                </template>
              </a-tab-pane>
            </a-tabs>
          </template>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>职位</h2>
            <a-button size="small" type="primary" @click="openRoleDrawer()">
              <PlusOutlined />
              新建职位
            </a-button>
          </div>
          <a-table
            row-key="id"
            size="small"
            :columns="roleColumns"
            :data-source="roles"
            :pagination="{ pageSize: 6, showTotal: (total: number) => `共 ${total} 个职位` }"
            :scroll="{ x: 720 }"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'role'">
                <a-space>
                  <a-avatar :style="{ backgroundColor: record.isAdmin ? '#d46b08' : '#1677ff' }">
                    {{ record.name.slice(0, 1) }}
                  </a-avatar>
                  <div class="name-cell">
                    <strong>{{ record.name }}</strong>
                    <span>{{ record.slug }}</span>
                  </div>
                </a-space>
              </template>
              <template v-else-if="column.key === 'summary'">
                <a-space :size="4" wrap>
                  <a-tag>{{ record.menuIds.length }} 菜单</a-tag>
                  <a-tag>{{ record.userIds.length }} 人</a-tag>
                  <a-tag v-if="record.isAdmin" color="gold">最高权限</a-tag>
                </a-space>
              </template>
              <template v-else-if="column.key === 'action'">
                <a-space>
                  <a-button size="small" :disabled="record.isAdmin" @click="openRoleDrawer(record)">
                    改名
                  </a-button>
                  <a-popconfirm title="确认删除该职位？" @confirm="removeRole(record)">
                    <a-button size="small" danger :disabled="record.isAdmin">删除</a-button>
                  </a-popconfirm>
                </a-space>
              </template>
            </template>
          </a-table>
        </section>
      </a-col>
    </a-row>

    <a-drawer
      v-model:open="departmentDrawerOpen"
      :title="editingDepartment ? '编辑部门' : '新建部门'"
      width="420"
      :destroy-on-close="true"
    >
      <a-form
        ref="departmentFormRef"
        :model="departmentForm"
        :rules="departmentRules"
        layout="vertical"
      >
        <a-form-item label="部门名称" name="name">
          <a-input v-model:value="departmentForm.name" placeholder="例如：运营部" />
        </a-form-item>
        <a-form-item label="部门编码" name="code">
          <a-input v-model:value="departmentForm.code" placeholder="例如：operation" />
        </a-form-item>
        <a-form-item label="排序" name="sort">
          <a-input-number v-model:value="departmentForm.sort" class="full-width" :min="0" />
        </a-form-item>
        <a-form-item label="状态" name="status">
          <a-select
            v-model:value="departmentForm.status"
            :options="[
              { label: '启用', value: 1 },
              { label: '停用', value: 2 },
            ]"
          />
        </a-form-item>
        <a-form-item label="描述" name="description">
          <a-textarea
            v-model:value="departmentForm.description"
            :rows="3"
            placeholder="可选，描述部门职责"
          />
        </a-form-item>
      </a-form>
      <template #extra>
        <a-space>
          <a-button @click="departmentDrawerOpen = false">取消</a-button>
          <a-button type="primary" :loading="savingDepartment" @click="submitDepartment">
            保存
          </a-button>
        </a-space>
      </template>
    </a-drawer>

    <a-drawer
      v-model:open="roleDrawerOpen"
      :title="editingRole ? '编辑职位' : '新建职位'"
      width="420"
      :destroy-on-close="true"
    >
      <a-form ref="roleFormRef" :model="roleForm" :rules="roleRules" layout="vertical">
        <a-form-item label="职位名称" name="name">
          <a-input v-model:value="roleForm.name" placeholder="例如：运营人员" />
        </a-form-item>
        <a-form-item label="职位标识" name="slug">
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
  createDepartment,
  createRole,
  deleteDepartment,
  deleteRole,
  getRbacOverview,
  updateDepartment,
  updateDepartmentRoles,
  updateRole,
  updateRoleMenus,
  type RbacDepartment,
  type RbacMenu,
  type RbacRole,
} from '#/api/kadmin/rbac';

type TreeNode = {
  title: string;
  key: number;
  children?: TreeNode[];
};

type TransferItem = {
  title: string;
};

const renderTransferItem = (item: TransferItem) => item.title;

const loading = ref(false);
const savingDepartment = ref(false);
const savingDepartmentRoles = ref(false);
const savingRole = ref(false);
const savingMenus = ref(false);
const departments = ref<RbacDepartment[]>([]);
const roles = ref<RbacRole[]>([]);
const menus = ref<RbacMenu[]>([]);
const selectedDepartmentId = ref<number>();
const selectedRoleId = ref<number>();
const departmentKeyword = ref('');
const activeTab = ref('departmentRoles');
const checkedMenuKeys = ref<number[]>([]);
const selectedDepartmentRoleKeys = ref<string[]>([]);
const departmentDrawerOpen = ref(false);
const roleDrawerOpen = ref(false);
const editingDepartment = ref<RbacDepartment | null>(null);
const editingRole = ref<RbacRole | null>(null);
const departmentFormRef = ref<FormInstance>();
const roleFormRef = ref<FormInstance>();

const departmentForm = reactive({
  name: '',
  code: '',
  description: '',
  sort: 0,
  status: 1,
});

const roleForm = reactive({
  name: '',
  slug: '',
});

const departmentRules = {
  name: [{ required: true, message: '请输入部门名称' }],
};

const roleRules = {
  name: [{ required: true, message: '请输入职位名称' }],
  slug: [{ required: true, message: '请输入职位标识' }],
};

const roleColumns = [
  { title: '职位', key: 'role', width: 260, fixed: 'left' },
  { title: '权限概览', key: 'summary', width: 260 },
  { title: '操作', key: 'action', width: 180, fixed: 'right' },
];

const filteredDepartments = computed(() => {
  const keyword = departmentKeyword.value.trim().toLowerCase();
  if (!keyword) return departments.value;
  return departments.value.filter(
    (department) =>
      department.name.toLowerCase().includes(keyword) ||
      department.code.toLowerCase().includes(keyword),
  );
});

const selectedDepartment = computed(() =>
  departments.value.find((department) => department.id === selectedDepartmentId.value),
);

const selectedRole = computed(() =>
  roles.value.find((role) => role.id === selectedRoleId.value),
);

const menuTreeData = computed(() => toTreeData(menus.value));

const transferRoles = computed(() =>
  roles.value.map((role) => ({
    key: String(role.id),
    title: `${role.name}（${role.slug}）`,
    description: `${role.menuIds.length} 个菜单权限`,
  })),
);

const departmentRoleOptions = computed(() => {
  if (!selectedDepartment.value) return [];
  return selectedDepartment.value.roleIds
    .map((roleId) => roles.value.find((role) => role.id === roleId))
    .filter((role): role is RbacRole => Boolean(role))
    .map((role) => ({
      label: `${role.name}（${role.slug}）`,
      value: role.id,
    }));
});

watch(selectedDepartment, (department) => {
  selectedDepartmentRoleKeys.value = department?.roleIds
    ? department.roleIds.map(String)
    : [];
  if (!department || !department.roleIds.includes(selectedRoleId.value || 0)) {
    selectedRoleId.value = department?.roleIds[0];
  }
});

watch(
  selectedRole,
  (role) => {
    checkedMenuKeys.value = role?.menuIds ? [...role.menuIds] : [];
  },
  { immediate: true },
);

onMounted(() => {
  void loadOverview();
});

async function loadOverview() {
  loading.value = true;
  try {
    const data = await getRbacOverview();
    departments.value = data.departments || [];
    roles.value = data.roles || [];
    menus.value = data.menus || [];
    if (
      !selectedDepartmentId.value ||
      !departments.value.some((department) => department.id === selectedDepartmentId.value)
    ) {
      selectedDepartmentId.value = departments.value[0]?.id;
    }
    if (
      selectedDepartment.value &&
      !selectedDepartment.value.roleIds.includes(selectedRoleId.value || 0)
    ) {
      selectedRoleId.value = selectedDepartment.value.roleIds[0];
    }
  } catch (error) {
    message.error(error instanceof Error ? error.message : '加载权限数据失败');
  } finally {
    loading.value = false;
  }
}

function selectDepartment(departmentId: number) {
  selectedDepartmentId.value = departmentId;
}

function openDepartmentDrawer(department?: RbacDepartment) {
  editingDepartment.value = department || null;
  departmentForm.name = department?.name || '';
  departmentForm.code = department?.code || '';
  departmentForm.description = department?.description || '';
  departmentForm.sort = department?.sort || 0;
  departmentForm.status = department?.status || 1;
  departmentDrawerOpen.value = true;
}

async function submitDepartment() {
  await departmentFormRef.value?.validate();
  savingDepartment.value = true;
  try {
    const payload = {
      name: departmentForm.name.trim(),
      code: departmentForm.code.trim(),
      description: departmentForm.description.trim(),
      sort: Number(departmentForm.sort) || 0,
      status: Number(departmentForm.status) || 1,
    };
    const department = editingDepartment.value
      ? await updateDepartment(editingDepartment.value.id, payload)
      : await createDepartment(payload);
    await loadOverview();
    selectedDepartmentId.value = department.id;
    departmentDrawerOpen.value = false;
    message.success('部门已保存');
  } catch (error) {
    message.error(error instanceof Error ? error.message : '保存部门失败');
  } finally {
    savingDepartment.value = false;
  }
}

async function removeDepartment() {
  if (!selectedDepartment.value) return;
  try {
    await deleteDepartment(selectedDepartment.value.id);
    selectedDepartmentId.value = undefined;
    selectedRoleId.value = undefined;
    await loadOverview();
    message.success('部门已删除');
  } catch (error) {
    message.error(error instanceof Error ? error.message : '删除部门失败');
  }
}

async function saveDepartmentRoles() {
  if (!selectedDepartment.value) return;
  savingDepartmentRoles.value = true;
  try {
    const roleIds = selectedDepartmentRoleKeys.value.map(Number);
    await updateDepartmentRoles(selectedDepartment.value.id, roleIds);
    await loadOverview();
    selectedRoleId.value = roleIds.includes(selectedRoleId.value || 0)
      ? selectedRoleId.value
      : roleIds[0];
    message.success('部门职位已保存');
  } catch (error) {
    message.error(error instanceof Error ? error.message : '保存部门职位失败');
  } finally {
    savingDepartmentRoles.value = false;
  }
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
    const currentDepartmentId = editingRole.value ? undefined : selectedDepartment.value?.id;
    const currentDepartmentRoleIds = editingRole.value
      ? []
      : [...(selectedDepartment.value?.roleIds || [])];
    const payload = {
      name: roleForm.name.trim(),
      slug: roleForm.slug.trim(),
    };
    const role = editingRole.value
      ? await updateRole(editingRole.value.id, payload)
      : await createRole(payload);
    if (currentDepartmentId) {
      await updateDepartmentRoles(currentDepartmentId, [
        ...new Set([...currentDepartmentRoleIds, role.id]),
      ]);
    }
    await loadOverview();
    if (currentDepartmentId) {
      selectedDepartmentId.value = currentDepartmentId;
    }
    selectedRoleId.value = role.id;
    roleDrawerOpen.value = false;
    message.success('职位已保存');
  } catch (error) {
    message.error(error instanceof Error ? error.message : '保存职位失败');
  } finally {
    savingRole.value = false;
  }
}

async function removeRole(role: RbacRole) {
  if (role.isAdmin) return;
  try {
    await deleteRole(role.id);
    if (selectedRoleId.value === role.id) {
      selectedRoleId.value = undefined;
    }
    await loadOverview();
    message.success('职位已删除');
  } catch (error) {
    message.error(error instanceof Error ? error.message : '删除职位失败');
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

function toTreeData(items: RbacMenu[]): TreeNode[] {
  return items.map((item) => ({
    title: item.uri ? `${item.title} (${item.uri})` : item.title,
    key: item.id,
    children: item.children ? toTreeData(item.children) : undefined,
  }));
}
</script>

