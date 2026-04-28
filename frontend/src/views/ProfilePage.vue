<template>
  <div class="profile-page">
    <div class="container">
      <div class="page-header">
        <h1 class="page-title">个人中心</h1>
      </div>

      <el-row :gutter="20">
        <el-col :span="6">
          <el-card class="menu-card">
            <el-menu :default-active="activeMenu" @select="handleMenuSelect">
              <el-menu-item index="profile">
                <el-icon><User /></el-icon>
                <span>个人信息</span>
              </el-menu-item>
              <el-menu-item index="membership">
                <el-icon><Star /></el-icon>
                <span>会员服务</span>
              </el-menu-item>
              <el-menu-item index="history">
                <el-icon><Clock /></el-icon>
                <span>已保存方案</span>
              </el-menu-item>
              <el-menu-item index="favorites">
                <el-icon><Collection /></el-icon>
                <span>我的收藏</span>
              </el-menu-item>
            </el-menu>
          </el-card>
        </el-col>

        <el-col :span="18">
          <el-card class="content-card">
            <div v-if="loading" class="loading-state">
              <el-skeleton :rows="6" animated />
            </div>

            <div v-else-if="activeMenu === 'profile'" class="section-content">
              <h3>个人信息</h3>
              <el-form :model="userForm" label-width="100px">
                <el-form-item label="用户名">
                  <el-input v-model="userForm.username" disabled />
                </el-form-item>
                <el-form-item label="邮箱">
                  <el-input v-model="userForm.email" />
                </el-form-item>
                <el-form-item label="手机号">
                  <el-input v-model="userForm.phone" />
                </el-form-item>
                <el-form-item>
                  <el-button
                    type="primary"
                    :loading="saving"
                    @click="saveProfile"
                  >
                    {{ saving ? '保存中...' : '保存修改' }}
                  </el-button>
                </el-form-item>
              </el-form>
            </div>

            <div
              v-else-if="activeMenu === 'membership'"
              class="section-content"
            >
              <h3>会员服务</h3>
              <div class="summary-card">
                <div class="summary-row">
                  <span>当前等级</span>
                  <strong>{{ membershipInfo.level }}</strong>
                </div>
                <div class="summary-row">
                  <span>到期时间</span>
                  <strong>{{ membershipInfo.expiry || '未开通' }}</strong>
                </div>
                <div class="summary-row">
                  <span>推荐额度</span>
                  <strong>
                    {{ membershipInfo.usageCount.recommendations }} /
                    {{ membershipInfo.limits.recommendations || '未设置' }}
                  </strong>
                </div>
                <div class="summary-row">
                  <span>检索额度</span>
                  <strong>
                    {{ membershipInfo.usageCount.searches }} /
                    {{ membershipInfo.limits.searches || '未设置' }}
                  </strong>
                </div>
              </div>
            </div>

            <div v-else-if="activeMenu === 'history'" class="section-content">
              <div class="section-header">
                <h3>已保存方案</h3>
                <el-button size="small" @click="loadSavedSchemes"
                  >刷新</el-button
                >
              </div>

              <el-empty
                v-if="savedSchemes.length === 0"
                description="还没有保存的方案"
              />
              <div v-else class="card-list">
                <div
                  v-for="scheme in savedSchemes"
                  :key="scheme.id"
                  class="info-card"
                >
                  <div class="info-card-header">
                    <strong>{{ scheme.name }}</strong>
                    <span>{{ formatDate(scheme.updatedAt) }}</span>
                  </div>
                  <div class="info-card-body">
                    <span>成绩：{{ scheme.studentInfo.score || '-' }}</span>
                    <span>省份：{{ scheme.studentInfo.province || '-' }}</span>
                    <span>推荐数：{{ scheme.recommendations.length }}</span>
                  </div>
                </div>
              </div>
            </div>

            <div v-else-if="activeMenu === 'favorites'" class="section-content">
              <div class="section-header">
                <h3>我的收藏</h3>
                <el-button size="small" @click="loadFavorites">刷新</el-button>
              </div>

              <el-empty
                v-if="favorites.length === 0"
                description="还没有收藏院校"
              />
              <div v-else class="card-list">
                <div v-for="item in favorites" :key="item.id" class="info-card">
                  <div class="info-card-header">
                    <strong>{{ item.name }}</strong>
                    <span>{{ item.province }} {{ item.city }}</span>
                  </div>
                  <div class="info-card-body">
                    <span>{{ item.type || '未知类型' }}</span>
                    <span>{{ item.level || '未知层次' }}</span>
                    <span>{{
                      item.rank ? `排名 ${item.rank}` : '暂无排名'
                    }}</span>
                  </div>
                  <div class="info-card-actions">
                    <el-button size="small" @click="viewUniversity(item.id)">
                      查看详情
                    </el-button>
                    <el-button
                      size="small"
                      type="danger"
                      plain
                      @click="removeFavorite(item.id)"
                    >
                      取消收藏
                    </el-button>
                  </div>
                </div>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';
import { User, Star, Clock, Collection } from '@element-plus/icons-vue';
import { ElMessage } from 'element-plus';
import { useUserStore } from '@/stores/user';
import { useRecommendationStore } from '@/stores/recommendation';
import { universityApi } from '@/api/university';
import type { MembershipInfo } from '@/types/user';
import type { University } from '@/types/university';

const router = useRouter();
const userStore = useUserStore();
const recommendationStore = useRecommendationStore();

const activeMenu = ref('profile');
const loading = ref(true);
const saving = ref(false);
const favorites = ref<University[]>([]);
const membershipInfo = ref<MembershipInfo>({
  level: 'free',
  features: [],
  usageCount: {
    recommendations: 0,
    searches: 0,
    analyses: 0,
  },
  limits: {
    recommendations: 0,
    searches: 0,
    analyses: 0,
  },
});

const userForm = reactive({
  username: '',
  email: '',
  phone: '',
});

const savedSchemes = computed(() => recommendationStore.savedSchemes);

const syncUserForm = () => {
  userForm.username = userStore.user?.username || '';
  userForm.email = userStore.user?.email || '';
  userForm.phone = userStore.user?.phone || '';
};

const loadProfile = async () => {
  if (!userStore.user) {
    userStore.init();
  }

  if (userStore.isLoggedIn) {
    try {
      const response = await userStore.getMembershipInfo();
      if (response.success && 'data' in response && response.data) {
        membershipInfo.value = response.data;
      }
    } catch {
      // ignore membership load failure
    }
  }

  syncUserForm();
};

const loadFavorites = async () => {
  const response = await universityApi.getFavorites();
  favorites.value = response.data;
};

const loadSavedSchemes = async () => {
  try {
    await recommendationStore.loadSavedSchemes();
  } catch {
    ElMessage.error('加载保存方案失败');
  }
};

const handleMenuSelect = async (key: string) => {
  loading.value = true;
  activeMenu.value = key;

  if (key === 'favorites') {
    await loadFavorites();
  } else if (key === 'history') {
    await loadSavedSchemes();
  } else if (key === 'membership') {
    await loadProfile();
  }

  loading.value = false;
};

const saveProfile = async () => {
  saving.value = true;
  try {
    const result = await userStore.updateProfile({
      email: userForm.email,
      phone: userForm.phone,
    });

    if (result.success) {
      ElMessage.success('保存成功');
      syncUserForm();
    } else {
      ElMessage.error(result.message || '保存失败');
    }
  } finally {
    saving.value = false;
  }
};

const removeFavorite = async (id: string) => {
  const response = await universityApi.toggleFavorite(id);
  if (response.success) {
    favorites.value = favorites.value.filter((item) => item.id !== id);
    ElMessage.success('已取消收藏');
  }
};

const viewUniversity = (id: string) => {
  router.push(`/universities/${id}`);
};

const formatDate = (value?: string) => {
  if (!value) {
    return '未记录';
  }
  return new Date(value).toLocaleString('zh-CN');
};

onMounted(async () => {
  await loadProfile();
  await loadSavedSchemes();
  loading.value = false;
});
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

.loading-state {
  padding: 40px;
}

.section-content {
  padding: 20px;
}

.section-content h3 {
  margin-bottom: 20px;
  color: #2c3e50;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
}

.summary-card {
  display: grid;
  gap: 12px;
}

.summary-row {
  display: flex;
  justify-content: space-between;
  padding: 12px 16px;
  background: #f8fafc;
  border-radius: 10px;
}

.card-list {
  display: grid;
  gap: 16px;
}

.info-card {
  padding: 16px;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  background: #fff;
}

.info-card-header {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.info-card-body {
  display: flex;
  gap: 16px;
  flex-wrap: wrap;
  color: #64748b;
}

.info-card-actions {
  display: flex;
  gap: 12px;
  margin-top: 16px;
}

@media (max-width: 768px) {
  .el-row {
    flex-direction: column;
  }

  .el-col {
    width: 100%;
    max-width: 100%;
  }

  .menu-card {
    margin-bottom: 20px;
    min-height: auto;
  }

  .info-card-header {
    flex-direction: column;
  }
}
</style>
