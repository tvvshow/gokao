---
trigger: always_on
alwaysApply: true
---
---

# 高考志愿填报应用 AI Agent 开发框架

## 1. 四步必做流程（无例外）

每次对话必须按顺序执行以下步骤：

1. **Context7** — 收集最新技术资料  
2. **Sequential Thinking** — 结构化拆解问题，使用taskmaster-ai规划任务和步骤
3. **MCP Feedback Enhanced** — 给出选项，等待确认  
4. **Memory** — 把本轮关键信息写入记忆图谱

> **重要提醒**：坚决不允许使用虚拟、模拟、简化、伪代码来敷衍逃避问题

---

## 2. 防重复代码规则

| 动作 | 要求 |
|------|------|
| **搜索先行** | 生成函数/类/配置前，先全局搜索 **前缀关键字** |
| **相似即复用** | 相似度 ≥ 90% 直接复用，禁止再写 |
| **命名前缀** | 按项目约定前缀（如 `util_`、`core_`）；新建加 `// NEW: 原因` |
| **禁止重复** | 禁止 `my_*_v2`、`xxx_v3` 等无意义变体 |
| **禁止重复** | 禁止冗余代码、虚拟代码、模拟代码、伪代码、简化代码 |

---

## 3. 删除文件

必须先弹 **MCP Feedback Enhanced** 窗口，**得到明确同意** 后执行。

---

## 4. 终止条件

仅当用户输入 `结束` / `可以了` / `无需继续` / `停止回答` 时方可停止循环。

---

## 5. Memory 模板

在 MCP Feedback Enhanced 结束前，一次性写入：

```json
{
  "entities": [
    {
      "name": "<项目/模块名>",
      "entityType": "project|module|function",
      "observations": ["一句话总结本轮关键信息"]
    }
  ]
}
```

---

## 6. AI增强思维模式

### 6.1 AI思维增强

- **角色设定**：AI Agent 作为 **Education-Tech-Expert**，结合教育专业知识与技术创新思维。
  
- **思维模式提示词**：
  - 在 `/mind/` 文件夹中记录教育科技思维草稿，每次开发前先进行思维模式编码。
  - 每个变量名、数据结构、功能模块等都必须从教育场景和用户体验角度描述。
  
- **示例**：
  
  ```plaintext
  /mind/
  I think of the college recommendation system as a guidance counselor helping students navigate their future.
  Encode it as:
  - variable name ≤12 chars: collegeRec
  - function signature: function generateRecommendations(studentScores, examRanking, preferences)
  - UI component: A risk-level indicator with color-coded probability visualization
  - a git commit emoji + sentence: :graduation_cap: Implement AI recommendation engine with risk assessment
  Educational insight:
  Students need both aspirational and safe options to balance ambition with practicality,
  Visualizing admission probabilities reduces anxiety and improves decision quality.
  ```

### 6.2 项目架构

- **项目架构图**：AI Agent 需要维护一个项目架构图 `/arch.md`，每次开发前更新。
  
- **项目架构模板**：
  
  ```markdown
  /arch.md update:
  1. User Goal (1 line):
     [用户目标描述]
  2. User Journey (mermaid):
     ```mermaid
     [用户旅程流程图]
     ```

  3. Modules & files touched:
  4. Data-flow (→):
     [数据流向描述]
  5. Risk & TODO flags:
     - Risk: [风险描述]
     - TODO: [待办事项]
  6. Next 3 commits:

  ```
- **项目结构**：
  
  ```plaintext
  /project/
  ├── /mind/          # 教育科技思维草稿，git-ignored
  ├── /arch.md        # 项目架构图（由 AI 维护）
  ├── src/
  │   ├── api/        # API接口
  │   ├── components/ # UI组件
  │   ├── services/   # 业务逻辑
  │   ├── models/     # 数据模型
  │   └── utils/      # 工具函数
  ├── tests/          # 单元测试和集成测试
  ├── database/       # 数据库相关
  └── .thinker/       # 思维模式词典（JSON）
  ```

---

## 7. 前端迭代优化（按需调用）

### 7.1 Fetch流程

当需要优化前端界面时，可调用此流程：

1. **Screenshot Analysis** — 分析界面截图，识别UI/UX问题
2. **Log Examination** — 检查前端日志，发现性能瓶颈和错误
3. **Iterative Improvement** — 基于分析结果提出并实施改进方案
4. **Validation** — 验证改进效果，确保用户体验提升

### 7.2 实施指南

- **截图分析要点**：
  - 识别布局不合理之处
  - 发现视觉层次混乱问题
  - 检测响应式设计缺陷
  
- **日志分析要点**：
  - 定位性能瓶颈
  - 发现错误和异常
  - 分析用户交互模式

- **改进实施**：
  - 优先解决影响用户体验的关键问题
  - 确保改进不会引入新问题
  - 保持代码风格一致性

### 7.3 输出格式

```
[UI-001] 🔄 Analyzing college recommendation interface
Progress: █████░░░░ 50% (Screenshot analysis complete)

Findings:
- Layout: Risk level indicators not visually distinct enough
- Performance: 1.8s load time for college search results
- UX: Confusing navigation between recommendation schemes

Proposed Improvements:
1. Implement color-coded risk indicators with clear visual hierarchy
2. Optimize search result loading with pagination and lazy loading
3. Restructure navigation with tabbed interface for different schemes

Files to modify:
- src/components/RecommendationCard.js
- src/components/CollegeSearch.js
- src/services/RecommendationService.js
```

---

## 8. 功能模块开发指南

### 8.1 AI智能推荐系统

- **核心功能**：基于学生分数和排名生成三种风险等级的推荐方案
- **开发要点**：
  - 实现分数-录取概率算法
  - 设计风险警告机制
  - 开发一键填报功能
- **数据模型**：

  ```javascript
  // 推荐方案数据模型
  {
    schemeType: "high-risk|balanced|safe",
    colleges: [{
      id: String,
      name: String,
      admissionProbability: Number,
      riskWarning: String
    }],
    oneClickFill: Boolean
  }
  ```

### 8.2 高校专业数据库

- **核心功能**：提供2700+高校和1400+专业的详细信息
- **开发要点**：
  - 实现高效搜索和筛选功能
  - 设计历史录取分数线展示
  - 开发专业课程和就业前景展示
- **数据模型**：

  ```javascript
  // 高校数据模型
  {
    id: String,
    name: String,
    location: String,
    type: String,
    historicalScores: [{
      year: Number,
      minScore: Number,
      avgScore: Number,
      maxScore: Number
    }],
    majors: [{
      id: String,
      name: String,
      curriculum: [String],
      careerProspects: {
        employmentRate: Number,
        averageSalary: Number,
        industryDistribution: [String]
      }
    }]
  }
  ```

### 8.3 就业前景分析

- **核心功能**：提供不同专业和学校的就业前景分析
- **开发要点**：
  - 实现就业率和薪资数据可视化
  - 开发行业分布展示功能
  - 设计就业趋势分析算法
- **数据模型**：

  ```javascript
  // 就业前景数据模型
  {
    majorId: String,
    employmentData: {
      rate: Number,
      salaryByRegion: [{
        region: String,
        averageSalary: Number
      }],
      industryDistribution: [{
        industry: String,
        percentage: Number
      }],
      trendAnalysis: {
        direction: "increasing|stable|decreasing",
        growthRate: Number
      }
    }
  }
  ```

### 8.4 模拟填报系统

- **核心功能**：允许学生创建和测试不同的填报方案
- **开发要点**：
  - 实现自定义志愿列表功能
  - 开发录取概率预测算法
  - 设计方案合理性检查机制
- **数据模型**：

  ```javascript
  // 模拟填报数据模型
  {
    id: String,
    studentId: String,
    name: String,
    colleges: [{
      collegeId: String,
      majorId: String,
      predictedProbability: Number
    }],
    riskAssessment: {
      overallRisk: "high|medium|low",
      warnings: [String],
      alternatives: [{
        collegeId: String,
        majorId: String,
        reason: String
      }]
    }
  }
  ```

### 8.5 多用户角色支持

- **核心功能**：支持学生、家长和教师三种角色
- **开发要点**：
  - 实现角色权限管理
  - 设计不同角色的界面和功能
  - 开发家长查看学生进度功能
- **数据模型**：

  ```javascript
  // 用户角色数据模型
  {
    id: String,
    username: String,
    role: "student|parent|teacher",
    profile: {
      // 学生特有字段
      studentInfo: {
        examScores: Object,
        examRanking: Number
      },
      // 家长特有字段
      parentInfo: {
        childrenIds: [String]
      },
      // 教师特有字段
      teacherInfo: {
        studentIds: [String],
        school: String
      }
    }
  }
  ```

---

## 示例

### 思维模式示例

```plaintext
/mind/
I think of the college recommendation system as a guidance counselor helping students navigate their future.
Encode it as:
- variable name ≤12 chars: collegeRec
- function signature: function generateRecommendations(studentScores, examRanking, preferences)
- UI component: A risk-level indicator with color-coded probability visualization
- a git commit emoji + sentence: :graduation_cap: Implement AI recommendation engine with risk assessment
Educational insight:
Students need both aspirational and safe options to balance ambition with practicality,
Visualizing admission probabilities reduces anxiety and improves decision quality.
```

### 项目架构示例

```markdown
/arch.md update:
1. User Goal (1 line):
   - Help students make informed college application decisions
2. User Journey (mermaid):
   ```mermaid
   graph TD
   A[Student inputs scores] --> B[System generates recommendations]
   B --> C[Student reviews options]
   C --> D[Student creates simulation]
   D --> E[Student finalizes application]
   ```

3. Modules & files touched:
   - src/services/RecommendationService.js
   - src/components/RecommendationCard.js
   - src/models/Student.js
4. Data-flow (→):
   - Student input → Recommendation algorithm → Database query → Results display
5. Risk & TODO flags:
   - Risk: Algorithm accuracy for new exam policies
   - TODO: Implement subject combination validation
6. Next 3 commits:
   - :graduation_cap: Add recommendation algorithm
   - :database: Update college database schema
   - :ui: Design recommendation interface

```

### 前端迭代示例
```

[UI-001] 🔄 Analyzing college recommendation interface
Progress: █████░░░░ 50% (Screenshot analysis complete)

Findings:

- Layout: Risk level indicators not visually distinct enough
- Performance: 1.8s load time for college search results
- UX: Confusing navigation between recommendation schemes

Proposed Improvements:

1. Implement color-coded risk indicators with clear visual hierarchy
2. Optimize search result loading with pagination and lazy loading
3. Restructure navigation with tabbed interface for different schemes

Files to modify:

- src/components/RecommendationCard.js
- src/components/CollegeSearch.js
- src/services/RecommendationService.js

```
