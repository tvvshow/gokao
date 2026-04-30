<template>
  <el-card class="content-card" aria-labelledby="form-heading">
    <template #header>
      <div class="card-header">
        <el-icon aria-hidden="true"><edit /></el-icon>
        <span id="form-heading">填写考生信息</span>
      </div>
    </template>

    <el-form
      ref="formRef"
      :model="studentForm"
      :rules="formRules"
      label-width="100px"
      @submit.prevent="handleSubmit"
      aria-label="考生信息表单"
    >
      <!-- 基础信息 -->
      <div class="form-section">
        <h3 class="section-title">基础信息</h3>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="高考分数" prop="score" required>
              <el-input
                v-model.number="studentForm.score"
                type="number"
                placeholder="请输入高考分数"
                :min="0"
                :max="750"
              >
                <template #append>分</template>
              </el-input>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="所在省份" prop="province" required>
              <el-select
                v-model="studentForm.province"
                placeholder="选择省份"
                style="width: 100%"
              >
                <el-option
                  v-for="province in provinces"
                  :key="province"
                  :label="province"
                  :value="province"
                />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="文理科类" prop="scienceType" required>
              <el-radio-group v-model="studentForm.scienceType">
                <el-radio label="理科">理科</el-radio>
                <el-radio label="文科">文科</el-radio>
                <el-radio label="新高考">新高考</el-radio>
              </el-radio-group>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="高考年份" prop="year">
              <el-select
                v-model="studentForm.year"
                placeholder="选择年份"
                style="width: 100%"
              >
                <el-option
                  v-for="year in years"
                  :key="year"
                  :label="year"
                  :value="year"
                />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="位次排名" prop="rank">
          <el-input
            v-model.number="studentForm.rank"
            type="number"
            placeholder="请输入位次排名（可选）"
          >
            <template #append>名</template>
          </el-input>
        </el-form-item>
      </div>

      <!-- 偏好设置 -->
      <div class="form-section">
        <h3 class="section-title">志愿偏好</h3>

        <el-form-item label="意向地区">
          <el-select
            v-model="studentForm.preferences.regions"
            multiple
            placeholder="选择意向地区"
            style="width: 100%"
          >
            <el-option
              v-for="region in regions"
              :key="region"
              :label="region"
              :value="region"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="专业类别">
          <el-select
            v-model="studentForm.preferences.majorCategories"
            multiple
            placeholder="选择感兴趣的专业类别"
            style="width: 100%"
          >
            <el-option
              v-for="category in majorCategories"
              :key="category"
              :label="category"
              :value="category"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="院校类型">
          <el-checkbox-group v-model="studentForm.preferences.universityTypes">
            <el-row :gutter="10">
              <el-col :span="12">
                <el-checkbox label="985工程">985工程</el-checkbox>
              </el-col>
              <el-col :span="12">
                <el-checkbox label="211工程">211工程</el-checkbox>
              </el-col>
              <el-col :span="12">
                <el-checkbox label="双一流">双一流</el-checkbox>
              </el-col>
              <el-col :span="12">
                <el-checkbox label="普通本科">普通本科</el-checkbox>
              </el-col>
              <el-col :span="12">
                <el-checkbox label="师范类">师范类</el-checkbox>
              </el-col>
              <el-col :span="12">
                <el-checkbox label="财经类">财经类</el-checkbox>
              </el-col>
              <el-col :span="12">
                <el-checkbox label="理工类">理工类</el-checkbox>
              </el-col>
              <el-col :span="12">
                <el-checkbox label="医药类">医药类</el-checkbox>
              </el-col>
              <el-col :span="12">
                <el-checkbox label="农林类">农林类</el-checkbox>
              </el-col>
              <el-col :span="12">
                <el-checkbox label="艺术类">艺术类</el-checkbox>
              </el-col>
              <el-col :span="12">
                <el-checkbox label="体育类">体育类</el-checkbox>
              </el-col>
              <el-col :span="12">
                <el-checkbox label="民族类">民族类</el-checkbox>
              </el-col>
            </el-row>
          </el-checkbox-group>
        </el-form-item>

        <el-form-item label="风险承受度">
          <el-radio-group v-model="studentForm.preferences.riskTolerance">
            <el-radio label="conservative">保守型（冲1保5稳4）</el-radio>
            <el-radio label="moderate">稳健型（冲2保3稳5）</el-radio>
            <el-radio label="aggressive">激进型（冲4保2稳4）</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item label="特殊要求">
          <el-input
            v-model="studentForm.preferences.specialRequirements"
            type="textarea"
            :rows="3"
            placeholder="如：不接受中外合作办学、希望在大城市、对某专业有特别偏好等"
          />
        </el-form-item>
      </div>

      <div class="form-actions" role="group" aria-label="表单操作">
        <el-button
          type="primary"
          size="large"
          :loading="loading"
          @click="handleSubmit"
          aria-label="生成智能推荐方案"
        >
          <el-icon aria-hidden="true"><magic-stick /></el-icon>
          生成智能推荐
        </el-button>
        <el-button size="large" @click="handleReset" aria-label="重置表单"
          >重置</el-button
        >
      </div>
    </el-form>
  </el-card>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import type { FormInstance, FormRules } from 'element-plus';
import { Edit, MagicStick } from '@element-plus/icons-vue';
import type { StudentInfo } from '@/types/recommendation';
import { PROVINCES, REGIONS, MAJOR_CATEGORIES } from '@/config/constants';

interface Props {
  studentInfo: StudentInfo;
  loading?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
});

const emit = defineEmits<{
  submit: [];
  reset: [];
}>();

const formRef = ref<FormInstance>();

// Direct reference to studentInfo prop - parent handles reactivity
const studentForm = computed(() => props.studentInfo);

// Form validation rules
const formRules: FormRules = {
  score: [
    { required: true, message: '请输入高考分数', trigger: 'blur' },
    {
      type: 'number',
      min: 0,
      max: 750,
      message: '分数应在0-750之间',
      trigger: 'blur',
    },
  ],
  province: [{ required: true, message: '请选择所在省份', trigger: 'change' }],
  scienceType: [
    { required: true, message: '请选择文理科类', trigger: 'change' },
  ],
};

// Use constants from config
const provinces = ref(PROVINCES);
const regions = ref(REGIONS);
const majorCategories = ref(MAJOR_CATEGORIES);

const years = computed(() => {
  const currentYear = new Date().getFullYear();
  return Array.from({ length: 5 }, (_, i) => currentYear - i);
});

// Handle form submission
const handleSubmit = async () => {
  if (!formRef.value) return;

  try {
    await formRef.value.validate();
    emit('submit');
  } catch {
    // Validation failed, handled by Element Plus
  }
};

// Handle form reset
const handleReset = () => {
  if (formRef.value) {
    formRef.value.resetFields();
  }
  emit('reset');
};

// Expose methods for parent component
defineExpose({
  validate: () => formRef.value?.validate(),
  resetFields: () => formRef.value?.resetFields(),
});
</script>

<style scoped>
.card-header {
  display: flex;
  align-items: center;
  font-weight: 600;
  color: #0f172a;
}

.card-header .el-icon {
  margin-right: 8px;
  color: #0ea5e9;
}

.form-section {
  margin-bottom: 28px;
}

.section-title {
  font-size: 16px;
  color: #0f172a;
  margin-bottom: 18px;
  padding-bottom: 8px;
  border-bottom: 2px solid rgb(14 165 233 / 0.28);
}

.content-card :deep(.el-form-item__label) {
  white-space: normal;
}

.content-card :deep(.el-select),
.content-card :deep(.el-input),
.content-card :deep(.el-textarea) {
  width: 100%;
}

.form-actions {
  display: flex;
  gap: 12px;
  justify-content: center;
  align-items: center;
  padding-top: 20px;
  border-top: 1px solid rgb(148 163 184 / 0.32);
}

.form-actions .el-button {
  padding: 12px 24px;
  min-height: 44px;
}

/* 移动端优化 */
@media (max-width: 768px) {
  .form-section {
    margin-bottom: 22px;
  }

  .content-card :deep(.el-form-item) {
    margin-bottom: 18px;
  }

  .form-actions .el-button {
    display: block;
    width: 100%;
    margin: 0;
  }

  /* 院校类型复选框移动端优化 */
  .content-card :deep(.el-checkbox-group .el-row) {
    display: grid !important;
    grid-template-columns: repeat(2, 1fr) !important;
    gap: 10px !important;
  }
}
</style>
