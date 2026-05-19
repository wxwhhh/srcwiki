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
          <div class="disclaimer-section">
            <el-checkbox v-model="disclaimerAgreed" class="disclaimer-checkbox">
              <span class="disclaimer-label">
                我已阅读并同意
                <a class="disclaimer-toggle" @click.prevent="disclaimerExpanded = !disclaimerExpanded">
                  《免责声明》
                  <el-icon class="toggle-icon" :class="{ expanded: disclaimerExpanded }">
                    <ArrowDown />
                  </el-icon>
                </a>
              </span>
            </el-checkbox>
            <transition name="fade">
              <div v-if="disclaimerExpanded" class="disclaimer-detail">
                <p class="disclaimer-title">⚠️ 免责声明</p>
                <p>本平台所收录内容均为已公开的 Nday 漏洞信息，来源于互联网公开渠道，仅供安全研究、技术学习和漏洞修复参考使用。</p>
                <ol class="disclaimer-list">
                  <li>本平台不提供任何 0day 漏洞、未公开漏洞情报或攻击服务</li>
                  <li>本平台不鼓励、不支持任何形式的非法渗透测试或网络攻击行为</li>
                  <li>使用者应遵守《中华人民共和国网络安全法》《中华人民共和国刑法》等相关法律法规</li>
                  <li>任何未经授权对他人系统进行测试的行为均属违法，后果由使用者自行承担</li>
                  <li>如发现本平台内容存在侵权，请联系管理员删除</li>
                </ol>
                <p class="disclaimer-agreement">使用本平台即表示您已阅读并同意上述声明。</p>
              </div>
            </transition>
          </div>
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            size="large"
            :loading="loading"
            :disabled="!disclaimerAgreed"
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
const disclaimerAgreed = ref(false)
const disclaimerExpanded = ref(false)

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
  if (!disclaimerAgreed.value) {
    ElMessage.warning('请先阅读并同意免责声明')
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

.disclaimer-section {
  width: 100%;
}

.disclaimer-checkbox {
  height: auto;
}

.disclaimer-label {
  font-size: 13px;
  color: #606266;
  line-height: 1.5;
}

.disclaimer-toggle {
  color: #0071E3;
  cursor: pointer;
  text-decoration: none;
  display: inline-flex;
  align-items: center;
  gap: 2px;

  &:hover {
    text-decoration: underline;
  }
}

.toggle-icon {
  font-size: 12px;
  transition: transform 0.2s;

  &.expanded {
    transform: rotate(180deg);
  }
}

.disclaimer-detail {
  margin-top: 10px;
  padding: 12px 14px;
  background: #F5F5F7;
  border-radius: 8px;
  font-size: 13px;
  line-height: 1.7;
  color: #1D1D1F;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
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
.disclaimer-title {
  font-weight: 600;
  margin: 0 0 8px;
  color: #303133;
}

.disclaimer-list {
  margin: 8px 0;
  padding-left: 20px;
}

.disclaimer-list li {
  margin-bottom: 4px;
  line-height: 1.6;
}

.disclaimer-agreement {
  margin-top: 10px;
  font-weight: 500;
  color: #303133;
}
</style>
