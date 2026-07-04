<template>
  <div class="page-stack">
    <section class="page-heading">
      <div>
        <h1>用户管理</h1>
        <p>维护管理员账号、昵称、头像、密码与职位分配</p>
      </div>
    </section>

    <section class="panel">
      <a-form :model="filters" layout="inline" class="search-form">
        <a-form-item label="关键词">
          <a-input
            v-model:value="filters.keyword"
            allow-clear
            class="control-md"
            placeholder="账号 / 昵称"
            @press-enter="searchUsers"
          >
            <template #prefix>
              <SearchOutlined />
            </template>
          </a-input>
        </a-form-item>
        <a-form-item label="部门">
          <a-input
            v-model:value="filters.department"
            allow-clear
            class="control-md"
            placeholder="部门名称"
            @press-enter="searchUsers"
          >
            <template #prefix>
              <SearchOutlined />
            </template>
          </a-input>
        </a-form-item>
        <a-form-item label="职位">
          <a-input
            v-model:value="filters.role"
            allow-clear
            class="control-md"
            placeholder="职位名称"
            @press-enter="searchUsers"
          >
            <template #prefix>
              <SearchOutlined />
            </template>
          </a-input>
        </a-form-item>
        <a-form-item>
          <a-space wrap>
            <a-button type="primary" @click="searchUsers">
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
        <span class="muted-text">按条件筛选用户，并在行内维护单个用户</span>
        <a-space wrap>
          <a-button type="primary" @click="openUserDrawer()">
            <PlusOutlined />
            新增用户
          </a-button>
          <a-button @click="loadUsers">
            <ReloadOutlined />
            刷新
          </a-button>
          <a-button @click="openImportModal">
            <UploadOutlined />
            导入
          </a-button>
          <a-dropdown>
            <a-button>
              <DownloadOutlined />
              导出
            </a-button>
            <template #overlay>
              <a-menu @click="handleExport">
                <a-menu-item key="xlsx">导出 Excel</a-menu-item>
                <a-menu-item key="csv">导出 CSV</a-menu-item>
                <a-menu-item key="sql">导出 SQL</a-menu-item>
              </a-menu>
            </template>
          </a-dropdown>
        </a-space>
      </div>



      <a-table
        row-key="id"
        class="compact-user-table"
        size="small"
        :columns="columns"
        :data-source="users"
        :loading="loading"
        :pagination="pagination"
        :scroll="{ x: 980 }"
        @change="handleTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'user'">
            <a-space>
              <a-avatar :src="record.avatar">
                {{ (record.name || record.username).slice(0, 1).toUpperCase() }}
              </a-avatar>
              <div class="name-cell">
                <strong>{{ record.name || record.username }}</strong>
                <span>{{ record.username }}</span>
              </div>
            </a-space>
          </template>

          <template v-else-if="column.key === 'departments'">
            <a-space :size="4" wrap>
              <a-tag v-for="department in record.departments" :key="department">
                {{ department }}
              </a-tag>
              <a-tag v-if="record.departments.length === 0">未分配</a-tag>
            </a-space>
          </template>

          <template v-else-if="column.key === 'roles'">
            <a-space :size="4" wrap>
              <a-tag v-for="role in record.roles" :key="role">{{ role }}</a-tag>
              <a-tag v-if="record.id === 1" color="gold">最高权限</a-tag>
              <a-tag v-if="record.roles.length === 0 && record.id !== 1">未分配</a-tag>
            </a-space>
          </template>

          <template v-else-if="column.key === 'action'">
            <a-space>
              <a-button type="link" size="small" @click="openUserDrawer(record)">编辑</a-button>
              <a-button type="link" size="small" @click="openPasswordModal(record)">
                重置密码
              </a-button>
              <a-popconfirm title="确认删除该用户？" @confirm="removeUser(record)">
                <a-button type="link" size="small" danger :disabled="record.id === 1">
                  删除
                </a-button>
              </a-popconfirm>
            </a-space>
          </template>
        </template>
      </a-table>
    </section>

    <a-drawer
      v-model:open="userDrawerOpen"
      :title="editingUser ? '编辑用户' : '新增用户'"
      width="520"
      :destroy-on-close="true"
    >
      <a-form ref="userFormRef" :model="userForm" :rules="userRules" layout="vertical">
        <a-form-item label="账号" name="username">
          <a-input v-model:value="userForm.username" placeholder="请输入登录账号" />
        </a-form-item>
        <a-form-item v-if="!editingUser" label="初始密码" name="password">
          <a-input-password
            v-model:value="userForm.password"
            placeholder="至少 8 位，包含字母和数字"
          />
        </a-form-item>
        <a-form-item label="昵称" name="name">
          <a-input v-model:value="userForm.name" placeholder="可留空，默认使用账号" />
        </a-form-item>
        <a-form-item label="头像 URL" name="avatar">
          <a-input v-model:value="userForm.avatar" placeholder="可选" />
        </a-form-item>
        <a-form-item label="职位">
          <a-select
            v-model:value="userForm.roleIds"
            mode="multiple"
            :disabled="editingUser?.id === 1"
            :options="roleOptions"
            placeholder="请选择职位"
          />
        </a-form-item>
      </a-form>
      <template #extra>
        <a-space>
          <a-button @click="userDrawerOpen = false">取消</a-button>
          <a-button type="primary" :loading="savingUser" @click="submitUser">保存</a-button>
        </a-space>
      </template>
    </a-drawer>

    <a-modal
      v-model:open="passwordModalOpen"
      title="重置密码"
      :confirm-loading="savingPassword"
      @ok="submitPassword"
    >
      <a-form ref="passwordFormRef" :model="passwordForm" :rules="passwordRules" layout="vertical">
        <a-form-item label="新密码" name="password">
          <a-input-password
            v-model:value="passwordForm.password"
            placeholder="至少 8 位，包含字母和数字"
          />
        </a-form-item>
      </a-form>
    </a-modal>

    <a-modal
      v-model:open="importModalOpen"
      title="导入用户"
      :confirm-loading="importing"
      @ok="submitImport"
    >
      <a-form :model="importForm" layout="vertical">
        <a-form-item label="格式">
          <a-radio-group v-model:value="importForm.format">
            <a-radio-button value="xlsx">Excel</a-radio-button>
            <a-radio-button value="csv">CSV</a-radio-button>
            <a-radio-button value="sql">SQL</a-radio-button>
          </a-radio-group>
        </a-form-item>
        <a-form-item v-if="importForm.format === 'xlsx'" label="Excel 文件">
          <a-upload
            accept=".xlsx"
            :before-upload="beforeImportUpload"
            :show-upload-list="false"
          >
            <a-button>
              <UploadOutlined />
              选择文件
            </a-button>
          </a-upload>
          <p class="muted-text">{{ importFileName || '支持导入导出的 users.xlsx 格式' }}</p>
        </a-form-item>
        <a-form-item v-else label="导入内容">
          <a-textarea
            v-model:value="importForm.content"
            :rows="8"
            placeholder="CSV 表头支持 username,password,name,avatar,role_ids；SQL 使用导出的 INSERT 格式"
          />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import {
  ClearOutlined,
  DownloadOutlined,
  PlusOutlined,
  ReloadOutlined,
  SearchOutlined,
  UploadOutlined,
} from '@ant-design/icons-vue';
import { message, type FormInstance, type UploadProps } from 'ant-design-vue';
import { computed, onMounted, reactive, ref } from 'vue';

import { getRbacOverview, type RbacRole } from '../api/rbac';
import {
  createUser,
  deleteUser,
  exportUsers,
  getUsers,
  importUsers,
  resetUserPassword,
  updateUser,
  type ManagedUser,
  type UserImportExportFormat,
} from '../api/users';

type TablePagination = {
  current?: number;
  pageSize?: number;
};

const loading = ref(false);
const savingUser = ref(false);
const savingPassword = ref(false);
const importing = ref(false);
const users = ref<ManagedUser[]>([]);
const roles = ref<RbacRole[]>([]);
const userDrawerOpen = ref(false);
const passwordModalOpen = ref(false);
const importModalOpen = ref(false);
const importFileName = ref('');
const editingUser = ref<ManagedUser | null>(null);
const passwordUser = ref<ManagedUser | null>(null);
const userFormRef = ref<FormInstance>();
const passwordFormRef = ref<FormInstance>();

const filters = reactive({
  keyword: '',
  department: '',
  role: '',
});

const userForm = reactive({
  username: '',
  password: '',
  name: '',
  avatar: '',
  roleIds: [] as number[],
});

const passwordForm = reactive({
  password: '',
});

const importForm = reactive<{
  format: UserImportExportFormat;
  content: string;
}>({
  format: 'xlsx',
  content: '',
});

const pagination = reactive({
  current: 1,
  pageSize: 10,
  showSizeChanger: true,
  showTotal: (total: number) => `共 ${total} 人`,
});

const columns = [
  { title: '用户', key: 'user', width: 220, fixed: 'left' },
  { title: '部门', key: 'departments', width: 220 },
  { title: '职位', key: 'roles', width: 260 },
  { title: '创建时间', dataIndex: 'createdAt', key: 'createdAt', width: 160 },
  { title: '操作', key: 'action', width: 190, fixed: 'right' },
];

const userRules = computed(() => ({
  username: [{ required: true, message: '请输入账号' }],
  password: editingUser.value
    ? []
    : [
        { required: true, message: '请输入初始密码' },
        { min: 8, message: '密码至少 8 位' },
      ],
}));

const passwordRules = {
  password: [
    { required: true, message: '请输入新密码' },
    { min: 8, message: '密码至少 8 位' },
  ],
};

const roleOptions = computed(() =>
  roles.value.map((role) => ({
    label: `${role.name}（${role.slug}）`,
    value: role.id,
  })),
);

onMounted(() => {
  void Promise.all([loadUsers(), loadRoles()]);
});

async function loadUsers() {
  loading.value = true;
  try {
    const data = await getUsers({
      keyword: filters.keyword.trim(),
      department: filters.department.trim(),
      role: filters.role.trim(),
    });
    users.value = (data.items || []).map((user) => ({
      ...user,
      roleIds: user.roleIds || [],
      roles: user.roles || [],
      departmentIds: user.departmentIds || [],
      departments: user.departments || [],
    }));
  } catch (error) {
    message.error(error instanceof Error ? error.message : '加载用户失败');
  } finally {
    loading.value = false;
  }
}

async function loadRoles() {
  try {
    const data = await getRbacOverview();
    roles.value = data.roles || [];
  } catch {
    roles.value = [];
  }
}

function resetSearch() {
  filters.keyword = '';
  filters.department = '';
  filters.role = '';
  pagination.current = 1;
  void loadUsers();
}

function searchUsers() {
  pagination.current = 1;
  void loadUsers();
}

function handleTableChange(pager: TablePagination) {
  pagination.current = pager.current || 1;
  pagination.pageSize = pager.pageSize || 10;
}

function openUserDrawer(record?: ManagedUser) {
  editingUser.value = record || null;
  userForm.username = record?.username || '';
  userForm.password = '';
  userForm.name = record?.name || '';
  userForm.avatar = record?.avatar || '';
  userForm.roleIds = record?.roleIds ? [...record.roleIds] : [];
  userDrawerOpen.value = true;
}

async function submitUser() {
  await userFormRef.value?.validate();
  savingUser.value = true;
  try {
    const payload = {
      username: userForm.username.trim(),
      name: userForm.name.trim(),
      avatar: userForm.avatar.trim(),
      roleIds: userForm.roleIds,
    };
    if (editingUser.value) {
      await updateUser(editingUser.value.id, payload);
    } else {
      await createUser({
        ...payload,
        password: userForm.password.trim(),
      });
    }
    userDrawerOpen.value = false;
    await loadUsers();
    message.success('用户已保存');
  } catch (error) {
    message.error(error instanceof Error ? error.message : '保存用户失败');
  } finally {
    savingUser.value = false;
  }
}

function openPasswordModal(record?: ManagedUser) {
  passwordUser.value = record || null;
  passwordForm.password = '';
  if (!passwordUser.value) {
    return;
  }
  passwordModalOpen.value = true;
}

async function submitPassword() {
  await passwordFormRef.value?.validate();
  if (!passwordUser.value) return;
  savingPassword.value = true;
  try {
    await resetUserPassword(passwordUser.value.id, passwordForm.password.trim());
    passwordModalOpen.value = false;
    message.success('密码已重置');
  } catch (error) {
    message.error(error instanceof Error ? error.message : '重置密码失败');
  } finally {
    savingPassword.value = false;
  }
}

async function removeUser(record: ManagedUser) {
  if (record.id === 1) return;
  try {
    await deleteUser(record.id);
    await loadUsers();
    message.success('用户已删除');
  } catch (error) {
    message.error(error instanceof Error ? error.message : '删除用户失败');
  }
}

function openImportModal() {
  importForm.format = 'xlsx';
  importForm.content = '';
  importFileName.value = '';
  importModalOpen.value = true;
}

const beforeImportUpload: UploadProps['beforeUpload'] = async (file) => {
  try {
    importForm.content = await readFileAsBase64(file);
    importFileName.value = file.name;
    message.success(`${file.name} 已选择`);
  } catch {
    message.error('读取 Excel 文件失败');
  }
  return false;
};

async function submitImport() {
  if (!importForm.content.trim()) {
    message.warning(importForm.format === 'xlsx' ? '请选择 Excel 文件' : '请粘贴导入内容');
    return;
  }
  importing.value = true;
  try {
    const data = await importUsers(importForm.format, importForm.content);
    importModalOpen.value = false;
    await loadUsers();
    message.success(`已导入 ${data.imported} 个用户`);
  } catch (error) {
    message.error(error instanceof Error ? error.message : '导入失败');
  } finally {
    importing.value = false;
  }
}

async function handleExport(event: { key: UserImportExportFormat }) {
  try {
    const { blob, filename } = await exportUsers(event.key);
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = filename;
    link.click();
    URL.revokeObjectURL(url);
  } catch (error) {
    message.error(error instanceof Error ? error.message : '导出失败');
  }
}

function readFileAsBase64(file: File) {
  return new Promise<string>((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => {
      const result = reader.result;
      if (!(result instanceof ArrayBuffer)) {
        reject(new Error('invalid file'));
        return;
      }
      resolve(arrayBufferToBase64(result));
    };
    reader.onerror = () => reject(reader.error);
    reader.readAsArrayBuffer(file);
  });
}

function arrayBufferToBase64(buffer: ArrayBuffer) {
  const bytes = new Uint8Array(buffer);
  const chunkSize = 0x8000;
  let binary = '';
  for (let i = 0; i < bytes.length; i += chunkSize) {
    const chunk = bytes.subarray(i, i + chunkSize);
    binary += String.fromCharCode(...chunk);
  }
  return window.btoa(binary);
}
</script>
