import { createApp } from 'vue';
import { createPinia } from 'pinia';
import App from './App.vue';
import router from './router';

// Import Bootstrap CSS
import 'bootstrap/dist/css/bootstrap.min.css';

// Import Bootstrap JS and attach to window
import * as bootstrap from 'bootstrap';
// @ts-ignore
window.bootstrap = bootstrap;

// Import Highlight.js CSS for syntax highlighting
import 'highlight.js/styles/github.css';

const app = createApp(App);
const pinia = createPinia();

app.use(pinia);
app.use(router);

app.mount('#app');
