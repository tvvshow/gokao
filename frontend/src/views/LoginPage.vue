<template>
  <div class="login-page">
    <div class="login-container">
      <div class="login-card">
        <div class="login-header">
          <h1>高考志愿填报助手</h1>
          <p>登录您的账户，开启智能志愿推荐之旅</p>
        </div>

        <el-tabs v-model="activeTab" class="login-tabs">
          <!-- 登录 -->
          <el-tab-pane label="登录" name="login">
            <el-form
              ref="loginFormRef"
              :model="loginForm"
              :rules="loginRules"
              size="large"
              @submit.prevent="handleLogin"
            >
              <el-form-item prop="username">
                <el-input
                  v-model="loginForm.username"
                  placeholder="请输入用户名/邮箱/手机号"
                  prefix-icon="User"
                />
              </el-form-item>
              <el-form-item prop="password">
                <el-input
                  v-model="loginForm.password"
                  type="password"
                  placeholder="请输入密码"
                  prefix-icon="Lock"
                  show-password
                />
              </el-form-item>
              <el-form-item>
                <div class="login-options">
                  <el-checkbox v-model="rememberMe">记住我</el-checkbox>
                  <el-link type="primary">忘记密码？</el-link>
                </div>
              </el-form-item>
              <el-form-item>
                <el-button
                  type="primary"
                  size="large"
                  style="width: 100%"
                  :loading="loginLoading"
                  @click="handleLogin"
                >
                  登录
                </el-button>
              </el-form-item>
            </el-form>
          </el-tab-pane>

          <!-- 注册 -->
          <el-tab-pane label="注册" name="register">
            <el-form
              ref="registerFormRef"
              :model="registerForm"
              :rules="registerRules"
              size="large"
              @submit.prevent="handleRegister"
            >
              <el-form-item prop="username">
                <el-input
                  v-model="registerForm.username"
                  placeholder="请输入用户名"
                  prefix-icon="User"
                />
              </el-form-item>
              <el-form-item prop="email">
                <el-input
                  v-model="registerForm.email"
                  placeholder="请输入邮箱"
                  prefix-icon="Message"
                />
              </el-form-item>
              <el-form-item prop="phone">
                <el-input
                  v-model="registerForm.phone"
                  placeholder="请输入手机号"
                  prefix-icon="Phone"
                />
              </el-form-item>
              <el-form-item prop="password">
                <el-input
                  v-model="registerForm.password"
                  type="password"
                  placeholder="请输入密码（至少8位，包含大小写字母和数字）"
                  prefix-icon="Lock"
                  show-password
                />
                <div v-if="registerForm.password" class="password-strength">
                  <el-progress 
                    :percentage="passwordStrength.score * 25" 
                    :color="passwordStrengthColor"
                    :show-text="false"
                    :stroke-width="4"
                  />
                  <span class="strength-text" :style="{ color: passwordStrengthColor }">
                    {{ passwordStrength.message }}
                  </span>
                </div>
              </el-form-item>
              <el-form-item prop="confirmPassword">
                <el-input
                  v-model="registerForm.confirmPassword"
                  type="password"
                  placeholder="请确认密码"
                  prefix-icon="Lock"
                  show-password
                />
              </el-form-item>
              <el-form-item>
                <el-checkbox v-model="agreeTerms">
                  我已阅读并同意 <el-link type="primary">用户协议</el-link> 和 <el-link type="primary">隐私政策</el-link>
                </el-checkbox>
              </el-form-item>
              <el-form-item>
                <el-button
                  type="primary"
                  size="large"
                  style="width: 100%"
                  :loading="registerLoading"
                  @click="handleRegister"
                >
                  注册
                </el-button>
              </el-form-item>
            </el-form>
          </el-tab-pane>
        </el-tabs>

        <div class="login-footer">
          <el-divider>或</el-divider>
          <el-button type="info" size="large" style="width: 100%" @click="guestLogin">
            游客体验
          </el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { 
  createPasswordValidator, 
  validatePasswordStrength,
  sanitizeFormData,
  createSafeTextValidator
} from '@/utils/validators'

const router = useRouter()
const userStore = useUserStore()

const activeTab = ref('login')
const rememberMe = ref(false)
const agreeTerms = ref(false)
const loginLoading = ref(false)
const registerLoading = ref(false)

const loginFormRef = ref<FormInstance>()
const registerFormRef = ref<FormInstance>()

// 登录表单
const loginForm = reactive({
  username: '',
  password: ''
})

// 注册表单
const registerForm = reactive({
  username: '',
  email: '',
  phone: '',
  password: '',
  confirmPassword: ''
})

// 密码强度计算
const passwordStrength = computed(() => {
  return validatePasswordStrength(registerForm.password)
})

// 密码强度颜色
const passwordStrengthColor = computed(() => {
  const score = passwordStrength.value.score
  if (score <= 1) return '#f56c6c'
  if (score === 2) return '#e6a23c'
  if (score === 3) return '#409eff'
  return '#67c23a'
})

// 验证规则
const loginRules: FormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { validator: createSafeTextValidator('用户名'), trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' }
  ]
}

const registerRules: FormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '用户名长度为3-20个字符', trigger: 'blur' },
    { pattern: /^[a-zA-Z0-9_\u4e00-\u9fa5]+$/, message: '用户名只能包含字母、数字、下划线和中文', trigger: 'blur' },
    { validator: createSafeTextValidator('用户名'), trigger: 'blur' }
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '邮箱格式不正确', trigger: 'blur' }
  ],
  phone: [
    { required: true, message: '请输入手机号', trigger: 'blur' },
    { pattern: /^1[3-9]\d{9}$/, message: '手机号格式不正确', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { validator: createPasswordValidator(8), trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (value !== registerForm.password) {
          callback(new Error('两次输入的密码不一致'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ]
}

// 登录处理
const handleLogin = async () => {
  if (!loginFormRef.value) return

  try {
    await loginFormRef.value.validate()
    
    loginLoading.value = true
    // Sanitize form data before sending
    const sanitizedForm = sanitizeFormData(loginForm)
    const result = await userStore.login(sanitizedForm)
    
    if (result.success) {
      ElMessage.success('登录成功')
      // Check for redirect query parameter
      const redirect = router.currentRoute.value.query.redirect as string
      router.push(redirect || '/')
    } else {
      ElMessage.error(result.message || '登录失败')
    }
  } catch {
    ElMessage.warning('请完善登录信息')
  } finally {
    loginLoading.value = false
  }
}

// 注册处理
const handleRegister = async () => {
  if (!registerFormRef.value) return

  if (!agreeTerms.value) {
    ElMessage.warning('请同意用户协议和隐私政策')
    return
  }

  try {
    await registerFormRef.value.validate()
    
    registerLoading.value = true
    // Sanitize form data before sending
    const sanitizedForm = sanitizeFormData(registerForm)
    const result = await userStore.register(sanitizedForm)
    
    if (result.success) {
      ElMessage.success('注册成功，请登录')
      activeTab.value = 'login'
    } else {
      ElMessage.error(result.message || '注册失败')
    }
  } catch {
    ElMessage.warning('请完善注册信息')
  } finally {
    registerLoading.value = false
  }
}

// 游客登录
const guestLogin = () => {
  ElMessage.success('以游客身份进入系统')
  router.push('/')
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}

.login-container {
  width: 100%;
  max-width: 400px;
}

.login-card {
  background: white;
  border-radius: 16px;
  padding: 40px;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.1);
  backdrop-filter: blur(10px);
}

.login-header {
  text-align: center;
  margin-bottom: 30px;
}

.login-header h1 {
  font-size: 24px;
  color: #2c3e50;
  margin-bottom: 8px;
}

.login-header p {
  color: #7f8c8d;
  font-size: 14px;
  margin: 0;
}

.login-tabs {
  margin-bottom: 20px;
}

.login-options {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.login-footer {
  margin-top: 20px;
}

.login-footer .el-divider {
  margin: 20px 0;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .login-card {
    padding: 30px 20px;
  }
  
  .login-header h1 {
    font-size: 20px;
  }
}

/* 密码强度指示器 */
.password-strength {
  margin-top: 8px;
  display: flex;
  align-items: center;
  gap: 10px;
}

.password-strength .el-progress {
  flex: 1;
}

.strength-text {
  font-size: 12px;
  white-space: nowrap;
}
</style>