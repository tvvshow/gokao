import { createApp } from 'vue';
import { createPinia } from 'pinia';
import ElementPlus from 'element-plus';
import 'element-plus/dist/index.css';
import 'element-plus/theme-chalk/dark/css-vars.css';
import App from './App.vue';
import router from './router';
import './style.css';
// UI-01：theme.css 必须在 style.css 之后导入，让新的 token 与 Element Plus 主题变量
// 覆盖旧深蓝主视觉。后续 UI-02 重写首页时再决定是否拆 style.css。
import './styles/theme.css';
import { useUserStore } from './stores/user';
import { setupErrorHandler } from './utils/errorHandler';

const app = createApp(App);
const pinia = createPinia();

app.use(pinia);
app.use(router);
app.use(ElementPlus);

// Setup global error handling
setupErrorHandler(app);

// 初始化用户状态
const userStore = useUserStore();
userStore.init();

app.mount('#app');
