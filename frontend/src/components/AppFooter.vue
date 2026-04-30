<template>
  <footer class="app-footer" role="contentinfo">
    <div class="footer-content">
      <div class="footer-section">
        <h3 id="footer-products">产品服务</h3>
        <ul aria-labelledby="footer-products">
          <li><a href="/universities">院校查询</a></li>
          <li><a href="/majors">专业分析</a></li>
          <li><a href="/recommendation">智能推荐</a></li>
          <li><a href="/analysis">数据分析</a></li>
        </ul>
      </div>

      <div class="footer-section">
        <h3 id="footer-user-services">用户服务</h3>
        <ul aria-labelledby="footer-user-services">
          <!-- <li><a href="/membership">会员服务</a></li> 暂时隐藏，支付功能待开发 -->
          <li><a href="/profile">个人中心</a></li>
          <li><a href="#help">帮助中心</a></li>
          <li><a href="#faq">常见问题</a></li>
        </ul>
      </div>

      <div class="footer-section">
        <h3 id="footer-about">关于我们</h3>
        <ul aria-labelledby="footer-about">
          <li><a href="#about">公司介绍</a></li>
          <li><a href="#contact">联系我们</a></li>
          <li><a href="#privacy">隐私政策</a></li>
          <li><a href="#terms">服务条款</a></li>
        </ul>
      </div>

      <div class="footer-section contact-info">
        <h3 id="footer-contact">联系方式</h3>
        <p>
          <el-icon aria-hidden="true"><phone /></el-icon>
          <a href="tel:400-888-8888" aria-label="客服电话：400-888-8888"
            >400-888-8888</a
          >
        </p>
        <p>
          <el-icon aria-hidden="true"><message /></el-icon>
          <a
            href="mailto:support@gaokaohub.com"
            aria-label="客服邮箱：support@gaokaohub.com"
            >support@gaokaohub.com</a
          >
        </p>
        <p>
          <el-icon aria-hidden="true"><location /></el-icon>
          北京市海淀区中关村大街
        </p>
        <div class="social-links" role="group" aria-label="社交媒体链接">
          <el-button circle aria-label="微信公众号">
            <el-icon aria-hidden="true"><ChatDotSquare /></el-icon>
          </el-button>
          <el-button circle aria-label="分享到社交媒体">
            <el-icon aria-hidden="true"><Share /></el-icon>
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
import {
  Phone,
  Message,
  Location,
  ChatDotSquare,
  Share,
} from '@element-plus/icons-vue';
</script>

<style scoped>
.app-footer {
  background:
    radial-gradient(
      900px circle at 5% 2%,
      rgb(14 165 233 / 0.22),
      transparent 45%
    ),
    radial-gradient(
      820px circle at 92% 95%,
      rgb(20 184 166 / 0.18),
      transparent 45%
    ),
    linear-gradient(180deg, #020617 0%, #0f172a 100%);
  color: #cbd5e1;
  margin-top: 60px;
  border-top: 1px solid rgb(56 189 248 / 0.26);
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
  color: #67e8f9;
  margin-bottom: 16px;
  font-size: 18px;
  font-weight: 600;
  letter-spacing: 0.01em;
}

.footer-section ul {
  list-style: none;
  padding: 0;
}

.footer-section li {
  margin-bottom: 8px;
}

.footer-section a {
  color: #cbd5e1;
  text-decoration: none;
  transition:
    color 0.2s ease,
    transform 0.2s ease;
}

.footer-section a:hover {
  color: #f8fafc;
  transform: translateX(2px);
}

.contact-info p {
  display: flex;
  align-items: center;
  margin-bottom: 12px;
  color: #cbd5e1;
}

.contact-info .el-icon {
  margin-right: 8px;
  color: #67e8f9;
}

.social-links {
  margin-top: 16px;
  display: flex;
  gap: 12px;
}

.social-links .el-button {
  background: rgb(30 41 59 / 0.88);
  border-color: rgb(71 85 105 / 0.8);
  color: #67e8f9;
}

.social-links .el-button:hover {
  background: #67e8f9;
  border-color: #67e8f9;
  color: #0b1120;
}

.footer-bottom {
  border-top: 1px solid rgb(56 189 248 / 0.2);
  padding: 20px 0;
  text-align: center;
  background: rgb(2 6 23 / 0.44);
}

.copyright p {
  margin: 0;
  color: #94a3b8;
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
