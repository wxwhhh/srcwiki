<template>
  <div class="login-page">
    <div class="login-card">
      <h1 class="login-title">LiteWiki</h1>
      <p class="login-subtitle">内部知识库</p>

      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        @submit.prevent="handleLogin"
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

        <el-form-item prop="captcha_code">
          <div class="captcha-row">
            <el-input
              v-model="form.captcha_code"
              placeholder="验证码"
              size="large"
              :prefix-icon="Key"
              @keyup.enter="handleLogin"
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
            class="login-btn"
            @click="handleLogin"
          >
            登录
          </el-button>
        </el-form-item>
      </el-form>

      <div class="login-footer">
        <router-link to="/admin/register" class="register-link">
          注册 →
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { User, Lock, Key, ArrowDown } from '@element-plus/icons-vue'
import { login, getMe, getCaptcha } from '@/api/bff'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const formRef = ref<FormInstance>()
const loading = ref(false)
const captchaImg = ref('')
const captchaId = ref('')


const form = reactive({
  username: '',
  password: '',
  captcha_code: '',
})

const rules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  captcha_code: [{ required: true, message: '请输入验证码', trigger: 'blur' }],
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

async function handleLogin() {
  if (false) {
    return
  }
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    const res = await login(form.username, form.password, captchaId.value, form.captcha_code)
    if (res.code === 0) {
      const meRes = await getMe()
      if (meRes.code === 0) {
        auth.setUser(meRes.data)
      }
      ElMessage.success('登录成功')
      const redirect = (route.query.redirect as string) || '/admin/'
      router.push(redirect)
    } else {
      ElMessage.error(res.message || '登录失败')
      refreshCaptcha()
    }
  } catch (err: any) {
    ElMessage.error(err.response?.data?.message || '登录失败')
    refreshCaptcha()
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  refreshCaptcha()
})
</script>

<style lang="scss" scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: white;
  padding: 24px;
}

.login-card {
  width: 100%;
  max-width: 400px;
}

.login-title {
  font-size: 36px;
  font-weight: 700;
  font-family: -apple-system, BlinkMacSystemFont, 'SF Pro Display', sans-serif;
  text-align: center;
  margin-bottom: 8px;
  color: #1D1D1F;
}

.login-subtitle {
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

.login-btn {
  width: 100%;
  height: 44px;
  font-size: 16px;
  border-radius: 8px;
}

.login-footer {
  text-align: center;
  margin-top: 24px;
}

.register-link {
  color: #0071E3;
  font-size: 14px;
  text-decoration: none;
  &:hover {
    text-decoration: underline;
  }
}

</style>
