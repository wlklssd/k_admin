<template>
  <div class="page-stack">
    <section class="page-heading">
      <div>
        <h1>工作台</h1>
        <p>系统运行概览</p>
      </div>
      <a-segmented v-model:value="range" :options="rangeOptions" />
    </section>

    <a-row :gutter="[16, 16]">
      <a-col v-for="item in stats" :key="item.title" :xs="24" :sm="12" :lg="6">
        <div class="stat-box">
          <span>{{ item.title }}</span>
          <strong>{{ item.value }}</strong>
          <a-progress :percent="item.percent" :show-info="false" size="small" />
        </div>
      </a-col>
    </a-row>

    <a-row :gutter="[16, 16]">
      <a-col :xs="24" :lg="15">
        <section class="panel">
          <div class="panel-title">
            <h2>模块进度</h2>
            <a-button type="text" size="small" @click="refresh">
              <ReloadOutlined />
            </a-button>
          </div>
          <a-table
            row-key="key"
            size="middle"
            :columns="columns"
            :data-source="tableData"
            :pagination="false"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'status'">
                <a-tag :color="statusColor(record.status)">
                  {{ statusText(record.status) }}
                </a-tag>
              </template>
              <template v-else-if="column.key === 'progress'">
                <a-progress :percent="record.progress" size="small" />
              </template>
            </template>
          </a-table>
        </section>
      </a-col>

      <a-col :xs="24" :lg="9">
        <section class="panel">
          <div class="panel-title">
            <h2>近期动态</h2>
            <a-tag color="processing">实时</a-tag>
          </div>
          <a-timeline>
            <a-timeline-item color="green">用户管理接口完成联调</a-timeline-item>
            <a-timeline-item color="blue">Ant Design Vue 页面骨架已接入</a-timeline-item>
            <a-timeline-item color="orange">文件存储等待 MinIO 配置</a-timeline-item>
            <a-timeline-item color="gray">代码生成器排入 P0</a-timeline-item>
          </a-timeline>
        </section>
      </a-col>
    </a-row>
  </div>
</template>

<script setup lang="ts">
import { ReloadOutlined } from '@ant-design/icons-vue';
import { message } from 'ant-design-vue';
import { computed, ref } from 'vue';

import { initialResources } from '../mock';
import type { ResourceStatus } from '../types';

const range = ref('今日');
const rangeOptions = ['今日', '本周', '本月'];

const stats = computed(() => {
  const boost = range.value === '今日' ? 0 : range.value === '本周' ? 8 : 16;
  return [
    { title: '在线模块', value: `${3 + Math.floor(boost / 8)}`, percent: 78 + boost / 2 },
    { title: '待办任务', value: `${12 + boost}`, percent: 46 + boost },
    { title: '接口成功率', value: `${98.2 - boost / 20}%`, percent: 92 },
    { title: '文件处理', value: `${264 + boost * 5}`, percent: 64 + boost / 2 },
  ];
});

const columns = [
  { title: '模块', dataIndex: 'name', key: 'name' },
  { title: '归属', dataIndex: 'owner', key: 'owner' },
  { title: '状态', dataIndex: 'status', key: 'status', width: 110 },
  { title: '进度', dataIndex: 'progress', key: 'progress', width: 180 },
];

const tableData = computed(() => initialResources.slice(0, 5));

function statusColor(status: ResourceStatus) {
  return {
    online: 'green',
    pending: 'orange',
    offline: 'red',
  }[status];
}

function statusText(status: ResourceStatus) {
  return {
    online: '运行中',
    pending: '待配置',
    offline: '停用',
  }[status];
}

function refresh() {
  message.success('已刷新');
}
</script>
