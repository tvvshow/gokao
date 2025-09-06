<template>
  <footer class="app-footer">
    <div class="footer-content">
      <div class="footer-section">
        <h3>产品服务</h3>
        <ul>
          <li><a href="/universities">院校查询</a></li>
          <li><a href="/majors">专业分析</a></li>
          <li><a href="/recommendation">智能推荐</a></li>
          <li><a href="/analysis">数据分析</a></li>
        </ul>
      </div>
      
      <div class="footer-section">
        <h3>用户服务</h3>
        <ul>
          <li><a href="/membership">会员服务</a></li>
          <li><a href="/profile">个人中心</a></li>
          <li><a href="#help">帮助中心</a></li>
          <li><a href="#faq">常见问题</a></li>
        </ul>
      </div>
      
      <div class="footer-section">
        <h3>关于我们</h3>
        <ul>
          <li><a href="#about">公司介绍</a></li>
          <li><a href="#contact">联系我们</a></li>
          <li><a href="#privacy">隐私政策</a></li>
          <li><a href="#terms">服务条款</a></li>
        </ul>
      </div>
      
      <div class="footer-section contact-info">
        <h3>联系方式</h3>
        <p><el-icon><phone /></el-icon> 400-888-8888</p>
        <p><el-icon><message /></el-icon> support@gaokaohub.com</p>
        <p><el-icon><location /></el-icon> 北京市海淀区中关村大街</p>
        <div class="social-links">
          <el-button circle>
            <el-icon><ChatDotSquare /></el-icon>
          </el-button>
          <el-button circle>
            <el-icon><Share /></el-icon>
          </el-button>
        </div>
      </div>
    </div>
    
    <div class="footer-bottom">
      <div class="copyright">
        <p>&copy; 2023 高考志愿填报助手. All rights reserved.</p>
        <p>ICP备案号：京ICP备xxxxxxxx号</p>
      </div>
    </div>
  </footer>
</template>

<script setup lang="ts">
import { Phone, Message, Location, ChatDotSquare, Share } from '@element-plus/icons-vue'
</script>

<style scoped>
.app-footer {
  background: #2c3e50;
  color: #ecf0f1;
  margin-top: 60px;
}

.footer-content {
  max-width: 1200px;
  margin: 0 auto;
  padding: 40px 20px;
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 30px;
}

.footer-section h3 {
  color: #3498db;
  margin-bottom: 16px;
  font-size: 18px;
  font-weight: 600;
}

.footer-section ul {
  list-style: none;
  padding: 0;
}

.footer-section li {
  margin-bottom: 8px;
}

.footer-section a {
  color: #bdc3c7;
  text-decoration: none;
  transition: color 0.3s;
}

.footer-section a:hover {
  color: #3498db;
}

.contact-info p {
  display: flex;
  align-items: center;
  margin-bottom: 12px;
  color: #bdc3c7;
}

.contact-info .el-icon {
  margin-right: 8px;
  color: #3498db;
}

.social-links {
  margin-top: 16px;
  display: flex;
  gap: 12px;
}

.social-links .el-button {
  background: #34495e;
  border-color: #34495e;
  color: #3498db;
}

.social-links .el-button:hover {
  background: #3498db;
  border-color: #3498db;
  color: white;
}

.footer-bottom {
  border-top: 1px solid #34495e;
  padding: 20px 0;
  text-align: center;
  background: #263238;
}

.copyright p {
  margin: 0;
  color: #95a5a6;
  font-size: 14px;
  line-height: 1.5;
}

@media (max-width: 768px) {
  .footer-content {
    grid-template-columns: 1fr;
    padding: 30px 15px;
    gap: 20px;
  }
  
  .footer-section {
    text-align: center;
  }
}
</style>