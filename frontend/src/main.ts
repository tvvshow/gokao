import { createApp } from 'vue';
import { createPinia } from 'pinia';
import ElementPlus from 'element-plus';
import 'element-plus/dist/index.css';
import 'element-plus/theme-chalk/dark/css-vars.css';
import App from './App.vue';
import router from './router';
import './style.css';
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
