<template>
  <section class="not-found-panel">
    <a-result
      status="404"
      title="404"
      sub-title="页面不存在或暂未配置到当前后台菜单。"
    >
      <template #extra>
        <a-space>
          <a-button type="primary" @click="$emit('goHome')">
            <HomeOutlined />
            返回工作台
          </a-button>
          <a-button @click="reloadPage">
            <ReloadOutlined />
            刷新
          </a-button>
        </a-space>
      </template>
    </a-result>
  </section>
</template>

<script setup lang="ts">
import { HomeOutlined, ReloadOutlined } from '@ant-design/icons-vue';

defineEmits<{
  goHome: [];
}>();

function reloadPage() {
  window.location.reload();
}
</script>
