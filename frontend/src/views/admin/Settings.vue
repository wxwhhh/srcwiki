<template>
  <div class="settings-page">
    <h2 class="page-title">系统设置</h2>

    <el-card class="settings-card">
      <template #header>
        <div class="card-header">
          <span>注册模式</span>
        </div>
      </template>

      <div class="setting-item">
        <div class="setting-label">
          <span class="label-text">用户注册方式</span>
          <span class="label-desc">控制新用户如何注册账号</span>
        </div>
        <div class="setting-control">
          <el-radio-group v-model="form.register_mode" :disabled="saving">
            <el-radio value="open">
              <div class="radio-option">
                <span class="radio-title">开放注册</span>
                <span class="radio-desc">任何人都可以注册，无需邀请码</span>
              </div>
            </el-radio>
            <el-radio value="invite">
              <div class="radio-option">
                <span class="radio-title">邀请制</span>
                <span class="radio-desc">必须使用有效邀请码才能注册</span>
              </div>
            </el-radio>
          </el-radio-group>
        </div>
      </div>

      <div class="setting-actions">
        <el-button type="primary" :loading="saving" @click="handleSave">
          保存设置
        </el-button>
      </div>
    </el-card>

    <el-card class="settings-card info-card">
      <el-alert
        title="安全说明"
        type="info"
        :closable="false"
        show-icon
      >
        <template #default>
          <p>邀请制模式下，注册接口会在后端强制校验邀请码，前端界面仅做展示控制。</p>
          <p>切换模式后立即生效，无需重启服务。</p>
        </template>
      </el-alert>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getAdminSettings, updateAdminSettings } from '@/api/admin'

const saving = ref(false)

const form = reactive({
  register_mode: 'invite' as string,
})

async function fetchSettings() {
  try {
    const res = await getAdminSettings()
    if (res.code === 0 && res.data) {
      form.register_mode = res.data.register_mode || 'invite'
    }
  } catch {
    ElMessage.error('获取设置失败')
  }
}

async function handleSave() {
  saving.value = true
  try {
    const res = await updateAdminSettings({ register_mode: form.register_mode })
    if (res.code === 0) {
      ElMessage.success('设置已保存')
      if (res.data) {
        form.register_mode = res.data.register_mode || form.register_mode
      }
    } else {
      ElMessage.error(res.message || '保存失败')
    }
  } catch {
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  fetchSettings()
})
</script>

<style lang="scss" scoped>
.settings-page {
  max-width: 700px;
}

.page-title {
  font-size: 20px;
  font-weight: 600;
  margin-bottom: 24px;
  color: #1D1D1F;
}

.settings-card {
  margin-bottom: 20px;
}

.card-header {
  font-weight: 600;
  font-size: 15px;
}

.setting-item {
  display: flex;
  align-items: flex-start;
  gap: 32px;
}

.setting-label {
  min-width: 140px;

  .label-text {
    display: block;
    font-weight: 500;
    font-size: 14px;
    color: #1D1D1F;
    margin-bottom: 4px;
  }

  .label-desc {
    display: block;
    font-size: 12px;
    color: #86868B;
  }
}

.setting-control {
  flex: 1;
}

.radio-option {
  display: flex;
  flex-direction: column;
  margin-left: 8px;

  .radio-title {
    font-weight: 500;
    font-size: 14px;
  }

  .radio-desc {
    font-size: 12px;
    color: #86868B;
    margin-top: 2px;
  }
}

:deep(.el-radio) {
  display: flex;
  align-items: flex-start;
  height: auto;
  margin-bottom: 16px;
}

:deep(.el-radio__label) {
  padding-left: 8px;
}

.setting-actions {
  margin-top: 24px;
  padding-top: 16px;
  border-top: 1px solid #E8E8ED;
}

.info-card {
  :deep(.el-alert__content p) {
    margin: 4px 0;
    font-size: 13px;
  }
}
</style>
