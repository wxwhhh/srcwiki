<template>
  <div class="profile-page">
    <h1 class="page-title">个人信息</h1>

    <div class="profile-section">
      <h2 class="section-title">基本信息</h2>
      <div class="info-row">
        <span class="info-label">用户名</span>
        <span class="info-value">{{ auth.user?.username }}</span>
      </div>
      <div class="info-row">
        <span class="info-label">角色</span>
        <span class="info-value">{{ roleLabel }}</span>
      </div>
      <div class="info-row">
        <span class="info-label">注册时间</span>
        <span class="info-value">{{ formatDate(auth.user?.created_at || '') }}</span>
      </div>
    </div>

    <div class="profile-section">
      <h2 class="section-title">修改密码</h2>
      <el-form
        ref="pwdFormRef"
        :model="pwdForm"
        :rules="pwdRules"
        label-position="top"
        class="pwd-form"
      >
        <el-form-item label="当前密码" prop="old_password">
          <el-input
            v-model="pwdForm.old_password"
            type="password"
            show-password
            placeholder="输入当前密码"
          />
        </el-form-item>

        <el-form-item label="新密码" prop="new_password">
          <el-input
            v-model="pwdForm.new_password"
            type="password"
            show-password
            placeholder="输入新密码"
          />
        </el-form-item>

        <el-form-item label="确认新密码" prop="confirm_password">
          <el-input
            v-model="pwdForm.confirm_password"
            type="password"
            show-password
            placeholder="再次输入新密码"
          />
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            :loading="pwdLoading"
            @click="handleChangePassword"
          >
            修改密码
          </el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import { changePassword } from '@/api/bff'

const auth = useAuthStore()
const pwdFormRef = ref<FormInstance>()
const pwdLoading = ref(false)

const roleLabel = computed(() => {
  const map: Record<string, string> = { admin: '管理员', editor: '编辑者', reader: '读者' }
  return map[auth.user?.role || ''] || auth.user?.role
})

const pwdForm = reactive({
  old_password: '',
  new_password: '',
  confirm_password: '',
})

const validateConfirm = (_rule: any, value: string, callback: any) => {
  if (value !== pwdForm.new_password) {
    callback(new Error('两次密码不一致'))
  } else {
    callback()
  }
}

const pwdRules: FormRules = {
  old_password: [{ required: true, message: '请输入当前密码', trigger: 'blur' }],
  new_password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, max: 50, message: '6-50 个字符', trigger: 'blur' },
  ],
  confirm_password: [
    { required: true, message: '请确认新密码', trigger: 'blur' },
    { validator: validateConfirm, trigger: 'blur' },
  ],
}

async function handleChangePassword() {
  const valid = await pwdFormRef.value?.validate().catch(() => false)
  if (!valid) return

  pwdLoading.value = true
  try {
    const res = await changePassword(pwdForm.old_password, pwdForm.new_password)
    if (res.code === 0) {
      ElMessage.success('密码修改成功')
      pwdForm.old_password = ''
      pwdForm.new_password = ''
      pwdForm.confirm_password = ''
    } else {
      ElMessage.error(res.message || '修改失败')
    }
  } catch (err: any) {
    ElMessage.error(err.response?.data?.message || '修改失败')
  } finally {
    pwdLoading.value = false
  }
}

function formatDate(dateStr: string): string {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  return d.toLocaleDateString('zh-CN', { year: 'numeric', month: 'long', day: 'numeric' })
}
</script>

<style lang="scss" scoped>
.profile-page {
  max-width: 600px;
  margin: 0 auto;
}

.page-title {
  font-size: 36px;
  font-weight: 600;
  font-family: -apple-system, BlinkMacSystemFont, 'SF Pro Display', sans-serif;
  margin-bottom: 48px;
}

.profile-section {
  margin-bottom: 48px;
}

.section-title {
  font-size: 22px;
  font-weight: 600;
  margin-bottom: 24px;
  padding-bottom: 12px;
  border-bottom: 1px solid #F2F2F7;
}

.info-row {
  display: flex;
  padding: 12px 0;
  border-bottom: 1px solid #F2F2F7;
}

.info-label {
  width: 120px;
  flex-shrink: 0;
  font-size: 14px;
  color: #86868B;
}

.info-value {
  font-size: 16px;
  color: #1D1D1F;
}

.pwd-form {
  max-width: 400px;
}
</style>
