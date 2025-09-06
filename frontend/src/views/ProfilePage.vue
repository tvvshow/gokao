<template>
  <div class="profile-page">
    <div class="container">
      <div class="page-header">
        <h1 class="page-title">个人中心</h1>
      </div>

      <el-row :gutter="20">
        <!-- 左侧菜单 -->
        <el-col :span="6">
          <el-card class="menu-card">
            <el-menu :default-active="activeMenu" @select="handleMenuSelect">
              <el-menu-item index="profile">
                <el-icon><user /></el-icon>
                <span>个人信息</span>
              </el-menu-item>
              <el-menu-item index="membership">
                <el-icon><star /></el-icon>
                <span>会员服务</span>
              </el-menu-item>
              <el-menu-item index="history">
                <el-icon><clock /></el-icon>
                <span>使用记录</span>
              </el-menu-item>
              <el-menu-item index="favorites">
                <el-icon><collection /></el-icon>
                <span>我的收藏</span>
              </el-menu-item>
            </el-menu>
          </el-card>
        </el-col>

        <!-- 右侧内容 -->
        <el-col :span="18">
          <el-card class="content-card">
            <!-- 个人信息 -->
            <div v-if="activeMenu === 'profile'" class="profile-content">
              <h3>个人信息</h3>
              <el-form :model="userForm" label-width="100px">
                <el-form-item label="用户名">
                  <el-input v-model="userForm.username" />
                </el-form-item>
                <el-form-item label="邮箱">
                  <el-input v-model="userForm.email" />
                </el-form-item>
                <el-form-item label="手机号">
                  <el-input v-model="userForm.phone" />
                </el-form-item>
                <el-form-item>
                  <el-button type="primary">保存修改</el-button>
                </el-form-item>
              </el-form>
            </div>

            <!-- 会员服务 -->
            <div v-else-if="activeMenu === 'membership'" class="membership-content">
              <h3>会员服务</h3>
              <p>会员功能正在开发中...</p>
            </div>

            <!-- 使用记录 -->
            <div v-else-if="activeMenu === 'history'" class="history-content">
              <h3>使用记录</h3>
              <p>使用记录功能正在开发中...</p>
            </div>

            <!-- 我的收藏 -->
            <div v-else-if="activeMenu === 'favorites'" class="favorites-content">
              <h3>我的收藏</h3>
              <p>收藏功能正在开发中...</p>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { User, Star, Clock, Collection } from '@element-plus/icons-vue'

const activeMenu = ref('profile')

const userForm = reactive({
  username: '测试用户',
  email: 'test@example.com',
  phone: '13800138000'
})

const handleMenuSelect = (key: string) => {
  activeMenu.value = key
}
</script>

<style scoped>
.profile-page {
  padding: 20px 0;
  min-height: calc(100vh - 160px);
}

.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 20px;
}

.page-header {
  margin-bottom: 30px;
}

.page-title {
  font-size: 28px;
  color: #2c3e50;
}

.menu-card,
.content-card {
  min-height: 500px;
}

.profile-content,
.membership-content,
.history-content,
.favorites-content {
  padding: 20px;
}

.profile-content h3,
.membership-content h3,
.history-content h3,
.favorites-content h3 {
  margin-bottom: 20px;
  color: #2c3e50;
}
</style>