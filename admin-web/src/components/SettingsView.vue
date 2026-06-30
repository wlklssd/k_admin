<template>
  <div class="page-stack">
    <section class="page-heading">
      <div>
        <h1>系统设置</h1>
        <p>基础参数</p>
      </div>
      <a-button :loading="initialLoading" @click="loadSettings">
        <ReloadOutlined />
        重新加载
      </a-button>
    </section>

    <section class="panel">
      <a-skeleton :loading="initialLoading" active>
        <a-alert
          v-if="apiError"
          class="form-alert"
          type="error"
          show-icon
          closable
          :message="apiError"
          @close="apiError = ''"
        />
        <a-form ref="formRef" :model="formState" :rules="rules" layout="vertical">
          <a-row :gutter="16">
            <a-col :xs="24" :md="12">
              <a-form-item label="系统名称" name="appName">
                <a-input v-model:value="formState.appName" />
              </a-form-item>
            </a-col>
            <a-col :xs="24" :md="12">
              <a-form-item label="接口前缀" name="apiPrefix">
                <a-input v-model:value="formState.apiPrefix" />
              </a-form-item>
            </a-col>
            <a-col :xs="24" :md="12">
              <a-form-item label="存储引擎" name="storage">
                <a-select v-model:value="formState.storage" :options="storageOptions" />
              </a-form-item>
            </a-col>
            <a-col :xs="24" :md="12">
              <a-form-item label="会话时长" name="sessionMinutes">
                <a-input-number
                  v-model:value="formState.sessionMinutes"
                  class="full-width"
                  :min="10"
                  :max="1440"
                  addon-after="分钟"
                />
              </a-form-item>
            </a-col>
            <a-col :xs="24" :md="12">
              <a-form-item label="验证码">
                <a-switch v-model:checked="formState.captcha" checked-children="开" un-checked-children="关" />
              </a-form-item>
            </a-col>
            <a-col :xs="24" :md="12">
              <a-form-item label="操作日志">
                <a-switch v-model:checked="formState.operationLog" checked-children="开" un-checked-children="关" />
              </a-form-item>
            </a-col>
          </a-row>

          <a-divider />

          <a-space wrap>
            <a-button type="primary" :loading="submitLoading" @click="submit">
              <SaveOutlined />
              保存
            </a-button>
            <a-button @click="reset">重置</a-button>
          </a-space>
        </a-form>
      </a-skeleton>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ReloadOutlined, SaveOutlined } from '@ant-design/icons-vue';
import { message, type FormInstance } from 'ant-design-vue';
import { onMounted, reactive, ref } from 'vue';

interface SettingsForm {
  appName: string;
  apiPrefix: string;
  storage: string;
  sessionMinutes: number;
  captcha: boolean;
  operationLog: boolean;
}

const defaultSettings: SettingsForm = {
  appName: 'PezMax Admin',
  apiPrefix: '/api',
  storage: 'local',
  sessionMinutes: 120,
  captcha: true,
  operationLog: true,
};

const formRef = ref<FormInstance>();
const initialLoading = ref(false);
const submitLoading = ref(false);
const apiError = ref('');
const formState = reactive<SettingsForm>({ ...defaultSettings });

const storageOptions = [
  { label: '本地存储', value: 'local' },
  { label: 'MinIO', value: 'minio' },
  { label: 'OSS', value: 'oss' },
];

const rules = {
  appName: [{ required: true, message: '请输入系统名称' }],
  apiPrefix: [{ required: true, message: '请输入接口前缀' }],
  storage: [{ required: true, message: '请选择存储引擎' }],
  sessionMinutes: [{ required: true, message: '请输入会话时长' }],
};

function loadSettings() {
  initialLoading.value = true;
  window.setTimeout(() => {
    Object.assign(formState, defaultSettings);
    initialLoading.value = false;
  }, 350);
}

async function submit() {
  apiError.value = '';
  await formRef.value?.validate();
  submitLoading.value = true;
  window.setTimeout(() => {
    if (formState.storage === 'minio' && !formState.apiPrefix.startsWith('/')) {
      apiError.value = '接口前缀需要以 / 开头';
      submitLoading.value = false;
      return;
    }
    submitLoading.value = false;
    message.success('保存成功');
  }, 450);
}

function reset() {
  Object.assign(formState, defaultSettings);
  apiError.value = '';
}

onMounted(loadSettings);
</script>
