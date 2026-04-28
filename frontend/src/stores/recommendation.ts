import { defineStore } from 'pinia';
import { ref } from 'vue';
import { recommendationApi } from '@/api/recommendation';
import type {
  StudentInfo,
  Recommendation,
  RecommendationScheme,
} from '@/types/recommendation';

export const useRecommendationStore = defineStore('recommendation', () => {
  // 状态
  const studentInfo = ref<StudentInfo | null>(null);
  const recommendations = ref<Recommendation[]>([]);
  const savedSchemes = ref<RecommendationScheme[]>([]);
  const isLoading = ref(false);
  const error = ref<string | null>(null);

  // 中文type映射为英文enum
  const mapRecommendationType = (
    chineseType: string
  ): 'aggressive' | 'moderate' | 'conservative' => {
    const typeMap: Record<string, 'aggressive' | 'moderate' | 'conservative'> =
      {
        冲刺: 'aggressive',
        稳妥: 'moderate',
        保底: 'conservative',
      };
    return typeMap[chineseType] || 'moderate';
  };

  // 获取推荐
  const generateRecommendations = async (studentData: StudentInfo) => {
    isLoading.value = true;
    error.value = null;

    try {
      const response =
        await recommendationApi.generateRecommendations(studentData);

      if (response.success) {
        studentInfo.value = studentData;
        // 转换type字段：中文 -> 英文enum
        recommendations.value = response.data.recommendations.map((rec) => ({
          ...rec,
          type: mapRecommendationType(rec.type),
        }));
        return recommendations.value;
      } else {
        throw new Error(response.message || '推荐生成失败');
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : '推荐生成失败';
      throw err;
    } finally {
      isLoading.value = false;
    }
  };

  // 保存推荐方案
  const saveScheme = async (schemeData: {
    name: string;
    studentInfo: StudentInfo;
    recommendations: Recommendation[];
  }) => {
    try {
      const response = await recommendationApi.saveScheme(schemeData);

      if (response.success) {
        const newScheme: RecommendationScheme = {
          id: response.data.id,
          name: schemeData.name,
          studentInfo: schemeData.studentInfo,
          recommendations: schemeData.recommendations,
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
        };

        savedSchemes.value.push(newScheme);
        return newScheme;
      } else {
        throw new Error(response.message || '保存失败');
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : '保存失败';
      throw err;
    }
  };

  // 加载保存的方案
  const loadSavedSchemes = async () => {
    try {
      const response = await recommendationApi.getSchemes();

      if (response.success) {
        savedSchemes.value = response.data;
        return savedSchemes.value;
      } else {
        throw new Error(response.message || '加载失败');
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : '加载失败';
      throw err;
    }
  };

  // 删除方案
  const deleteScheme = async (schemeId: string) => {
    try {
      // Note: deleteScheme API not implemented yet, using local deletion
      savedSchemes.value = savedSchemes.value.filter(
        (scheme) => scheme.id !== schemeId
      );
    } catch (err) {
      error.value = err instanceof Error ? err.message : '删除失败';
      throw err;
    }
  };

  // 清空状态
  const clearRecommendations = () => {
    recommendations.value = [];
    studentInfo.value = null;
    error.value = null;
  };

  // 获取推荐统计
  const getRecommendationStats = () => {
    if (recommendations.value.length === 0) {
      return {
        total: 0,
        successRate: 0,
        riskLevel: 'unknown',
        matchScore: 0,
      };
    }

    const avgProbability =
      recommendations.value.reduce(
        (sum, rec) => sum + rec.admissionProbability,
        0
      ) / recommendations.value.length;
    const avgMatch =
      recommendations.value.reduce((sum, rec) => sum + rec.matchScore, 0) /
      recommendations.value.length;

    return {
      total: recommendations.value.length,
      successRate: Math.round(avgProbability),
      riskLevel: studentInfo.value?.preferences?.riskTolerance || 'moderate',
      matchScore: Math.round(avgMatch),
    };
  };

  // 按类型筛选推荐
  const getRecommendationsByType = (type: string) => {
    return recommendations.value.filter((rec) => rec.type === type);
  };

  // 导出推荐报告
  const exportReport = async (): Promise<Blob> => {
    if (!studentInfo.value || recommendations.value.length === 0) {
      throw new Error('没有可导出的推荐数据');
    }

    try {
      const blob = await recommendationApi.exportReport(recommendations.value);
      return blob;
    } catch (err) {
      error.value = err instanceof Error ? err.message : '导出失败';
      throw err;
    }
  };

  return {
    // 状态
    studentInfo,
    recommendations,
    savedSchemes,
    isLoading,
    error,

    // 操作
    generateRecommendations,
    saveScheme,
    loadSavedSchemes,
    deleteScheme,
    clearRecommendations,
    getRecommendationStats,
    getRecommendationsByType,
    exportReport,
  };
});
