<template>
  <div class="register-page">
    <div class="register-card">
      <h1 class="register-title">注册</h1>
      <p class="register-subtitle">{{ registerMode === 'open' ? '创建你的账号' : '使用邀请码创建账号' }}</p>

      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        @submit.prevent="handleRegister"
      >
        <el-form-item prop="username">
          <el-input
            v-model="form.username"
            placeholder="用户名"
            size="large"
            :prefix-icon="User"
          />
        </el-form-item>

        <el-form-item prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="密码"
            size="large"
            show-password
            :prefix-icon="Lock"
          />
        </el-form-item>

        <el-form-item prop="confirmPassword">
          <el-input
            v-model="form.confirmPassword"
            type="password"
            placeholder="确认密码"
            size="large"
            show-password
            :prefix-icon="Lock"
          />
        </el-form-item>

        <el-form-item v-if="registerMode === 'invite'" prop="invite_code">
          <el-input
            v-model="form.invite_code"
            placeholder="邀请码"
            size="large"
            :prefix-icon="Ticket"
          />
        </el-form-item>

        <el-form-item prop="captcha_code">
          <div class="captcha-row">
            <el-input
              v-model="form.captcha_code"
              placeholder="验证码"
              size="large"
              :prefix-icon="Key"
            />
            <div class="captcha-img-wrap" @click="refreshCaptcha" title="点击刷新">
              <img
                v-if="captchaImg"
                :src="captchaImg"
                class="captcha-img"
                alt="验证码"
              />
              <div v-else class="captcha-placeholder">
                <el-icon class="is-loading"><Loading /></el-icon>
              </div>
            </div>
          </div>
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            size="large"
            :loading="loading"
            class="register-btn"
            @click="handleRegister"
          >
            注册
          </el-button>
        </el-form-item>
      </el-form>

      <div class="register-footer">
        <router-link to="/admin/login" class="login-link">
          ← 返回登录
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { User, Lock, Ticket, Key } from '@element-plus/icons-vue'
import { register, getCaptcha, getPublicSettings } from '@/api/bff'

const router = useRouter()
const formRef = ref<FormInstance>()
const loading = ref(false)
const captchaImg = ref('')
const captchaId = ref('')
const registerMode = ref('invite')

const form = reactive({
  username: '',
  password: '',
  confirmPassword: '',
  invite_code: '',
  captcha_code: '',
})

const validateConfirm = (_rule: any, value: string, callback: any) => {
  if (value !== form.password) {
    callback(new Error('两次密码不一致'))
  } else {
    callback()
  }
}

const rules: FormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '3-20 个字符', trigger: 'blur' },
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, max: 50, message: '6-50 个字符', trigger: 'blur' },
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    { validator: validateConfirm, trigger: 'blur' },
  ],
  invite_code: [
    { required: false, message: '请输入邀请码', trigger: 'blur' },
  ],
  captcha_code: [
    { required: true, message: '请输入验证码', trigger: 'blur' },
  ],
}

async function fetchRegisterMode() {
  try {
    const res = await getPublicSettings()
    if (res.code === 0 && res.data) {
      registerMode.value = res.data.register_mode || 'invite'
    }
  } catch {
    // 默认 invite
  }
}

async function refreshCaptcha() {
  try {
    const res = await getCaptcha()
    if (res.code === 0) {
      captchaImg.value = res.data.captcha_img
      captchaId.value = res.data.captcha_id
      form.captcha_code = ''
    }
  } catch {
    // ignore
  }
}

async function handleRegister() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    const res = await register(form.username, form.password, form.invite_code, captchaId.value, form.captcha_code)
    if (res.code === 0) {
      ElMessage.success('注册成功，请登录')
      router.push('/admin/login')
    } else {
      ElMessage.error(res.message || '注册失败')
      refreshCaptcha()
    }
  } catch (err: any) {
    ElMessage.error(err.response?.data?.message || '注册失败')
    refreshCaptcha()
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchRegisterMode()
  refreshCaptcha()
})
</script>

<style lang="scss" scoped>
.register-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: white;
  padding: 24px;
}

.register-card {
  width: 100%;
  max-width: 400px;
}

.register-title {
  font-size: 36px;
  font-weight: 700;
  font-family: -apple-system, BlinkMacSystemFont, 'SF Pro Display', sans-serif;
  text-align: center;
  margin-bottom: 8px;
  color: #1D1D1F;
}

.register-subtitle {
  text-align: center;
  color: #86868B;
  font-size: 16px;
  margin-bottom: 48px;
}

.captcha-row {
  display: flex;
  gap: 12px;
  width: 100%;
}

.captcha-img-wrap {
  flex-shrink: 0;
  cursor: pointer;
  border-radius: 8px;
  overflow: hidden;
  border: 1px solid #E8E8ED;
  height: 40px;
  width: 120px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: border-color 0.15s;

  &:hover {
    border-color: #0071E3;
  }
}

.captcha-img {
  width: 120px;
  height: 40px;
  display: block;
}

.captcha-placeholder {
  color: #AEAEB2;
}

.register-btn {
  width: 100%;
  height: 44px;
  font-size: 16px;
  border-radius: 8px;
}

.register-footer {
  text-align: center;
  margin-top: 24px;
}

.login-link {
  color: #0071E3;
  font-size: 14px;
  text-decoration: none;
  &:hover {
    text-decoration: underline;
  }
}
</style>
