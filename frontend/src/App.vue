<template>
  <div class="app-container">
    <header class="modern-header">
      <div class="header-content">
        <div class="logo-section">
          <i class="fas fa-cube logo-icon"></i>
          <h1>Docklet 服务面板</h1>
        </div>
        <div class="header-subtitle">
          智能服务发现与管理平台
        </div>
      </div>
    </header>
    <main class="main-content">
      <!-- Search and Filter Section -->
      <div v-if="!loading && !error && (services.length > 0 || systemWebServices.length > 0)" class="search-section">
        <div class="search-container">
          <div class="search-input-wrapper">
            <i class="fas fa-search search-icon"></i>
            <input 
              v-model="searchQuery" 
              type="text" 
              placeholder="搜索服务名称、描述或端口..."
              class="search-input"
            />
            <button 
              v-if="searchQuery" 
              @click="clearSearch" 
              class="clear-search-btn"
            >
              <i class="fas fa-times"></i>
            </button>
          </div>
          <div class="view-controls">
            <button 
              @click="toggleAllSections" 
              class="toggle-all-btn"
              :class="{ 'collapsed': allSectionsCollapsed }"
            >
              <i :class="allSectionsCollapsed ? 'fas fa-expand-alt' : 'fas fa-compress-alt'"></i>
              {{ allSectionsCollapsed ? '展开全部' : '收起全部' }}
            </button>
          </div>
        </div>
        <div v-if="searchQuery" class="search-results-info">
          <span class="search-count">
            找到 {{ filteredDockerServices.length + filteredSystemServices.length }} 个匹配的服务
          </span>
          <button @click="clearSearch" class="clear-results-btn">
            清除搜索
          </button>
        </div>
      </div>

      <div v-if="loading" class="status-card loading-card">
      <div class="loading-spinner"></div>
      <p>正在扫描服务中...</p>
    </div>
    
    <div v-if="error" class="status-card error-card">
      <i class="fas fa-exclamation-triangle"></i>
      <p>{{ error }}</p>
    </div>
    
    <div v-if="!loading && !error && services.length === 0 && systemWebServices.length === 0" class="status-card empty-card">
      <i class="fas fa-search"></i>
      <h3>未发现服务</h3>
      <p>暂未找到可用的 Docker 服务或系统 Web 服务</p>
    </div>

    <!-- Docker Services Section -->
    <div v-if="!loading && !error && (searchQuery ? filteredDockerServices.length > 0 : services.length > 0)" class="services-section">
       <div class="section-header" @click="toggleDockerSection">
         <i class="fab fa-docker section-icon"></i>
         <h2>Docker 容器服务</h2>
         <span class="service-count">{{ searchQuery ? filteredDockerServices.length : services.length }} 个服务</span>
         <button class="collapse-btn" :class="{ 'collapsed': dockerSectionCollapsed }">
           <i :class="dockerSectionCollapsed ? 'fas fa-chevron-down' : 'fas fa-chevron-up'"></i>
         </button>
       </div>
       <div
           id="services-grid"
           class="services-grid"
           v-show="!dockerSectionCollapsed"
       >
           <div
           v-for="service in (searchQuery ? filteredDockerServices : sortedServices)"
           :key="service.id"
           class="service-card"
           >
           <div class="service-header">
             <div class="service-icon-wrapper">
               <span
               v-if="
                   service.icon &&
                   (service.icon.startsWith('http://') ||
                   service.icon.startsWith('https://') ||
                   service.icon.startsWith('/'))
               "
               >
               <img
                   :src="service.icon"
                   :alt="(service.title || service.name) + ' icon'"
                   class="service-icon-image"
               />
               </span>
               <span
               v-else-if="service.icon && service.icon.startsWith('fa-')"
               class="service-icon"
               >
               <i :class="['fas', service.icon]"></i>
               </span>
               <span
               v-else-if="service.icon"
               class="service-icon emoji-icon"
               >{{ service.icon }}</span>
               <span v-else class="service-icon default-icon">
               <i class="fas fa-cube"></i>
               </span>
             </div>
             <div class="service-info">
               <h3 class="service-title">{{ service.title || service.name || "未知Docker服务" }}</h3>
               <p class="service-description">{{ service.description || `容器: ${service.container_name}` }}</p>
             </div>
           </div>
           <div class="service-actions">
             <a
                 v-if="service.url"
                 :href="service.url"
                 target="_blank"
                 class="service-button primary"
             >
               <i class="fas fa-external-link-alt"></i>
               访问服务
             </a>
             <div v-else class="service-status offline">
               <i class="fas fa-times-circle"></i>
               暂无访问链接
             </div>
           </div>
           </div>
       </div>
    </div>

     <!-- System Web Services Section -->
     <div v-if="!loadingSystemServices && !errorSystemServices && (searchQuery ? filteredSystemServices.length > 0 : systemWebServices.length > 0)" class="services-section">
       <div class="section-header" @click="toggleSystemSection">
         <i class="fas fa-server section-icon"></i>
         <h2>本机 Web 服务</h2>
         <span class="service-count">{{ searchQuery ? filteredSystemServices.length : systemWebServices.length }} 个服务</span>
         <button class="collapse-btn" :class="{ 'collapsed': systemSectionCollapsed }">
           <i :class="systemSectionCollapsed ? 'fas fa-chevron-down' : 'fas fa-chevron-up'"></i>
         </button>
       </div>
       <div
         id="system-services-grid"
         class="services-grid"
         v-show="!systemSectionCollapsed"
       >
         <div
           v-for="sysService in (searchQuery ? filteredSystemServices : systemWebServices)"
           :key="sysService.name"
           class="service-card"
         >
           <div class="service-header">
             <div class="service-icon-wrapper">
               <span class="service-icon system-icon">
                 <i class="fas fa-cogs"></i>
               </span>
             </div>
             <div class="service-info">
               <h3 class="service-title">{{ sysService.display_name || sysService.name }}</h3>
               <div class="service-meta">
                 <span class="status-badge" :class="getStatusClass(sysService.status)">
                   <i :class="getStatusIcon(sysService.status)"></i>
                   {{ sysService.status }}
                 </span>
                 <span v-if="sysService.pid && sysService.pid !== '-'" class="pid-info">
                   PID: {{ sysService.pid }}
                 </span>
               </div>
             </div>
           </div>
           <div class="service-actions">
             <div v-if="sysService.listening_ports && sysService.listening_ports.length > 0" class="port-list">
               <div class="port-header" @click="toggleServicePorts(sysService.name)">
                 <span>监听端口:</span>
                 <span class="port-count-badge">{{ sysService.listening_ports.length }} 个端口</span>
                 <button class="port-toggle-btn" :class="{ 'collapsed': isServicePortsCollapsed(sysService.name) }">
                   <i :class="isServicePortsCollapsed(sysService.name) ? 'fas fa-chevron-down' : 'fas fa-chevron-up'"></i>
                 </button>
               </div>
               <div class="port-buttons" v-show="!isServicePortsCollapsed(sysService.name)">
                 <a 
                   v-for="port in sysService.listening_ports" 
                   :key="port"
                   :href="`http://${currentHostname}:${port}`" 
                   target="_blank" 
                   class="service-button port-button"
                 >
                   <i class="fas fa-plug"></i>
                   端口 {{ port }}
                 </a>
               </div>
             </div>
             <div v-else class="service-status no-ports">
               <i class="fas fa-info-circle"></i>
               未检测到监听端口
             </div>
           </div>
         </div>
       </div>
     </div>
      <div v-if="loadingSystemServices" class="status-card loading-card">
        <div class="loading-spinner"></div>
        <p>正在扫描本机 Web 服务...</p>
      </div>
      <div v-if="errorSystemServices" class="status-card error-card">
        <i class="fas fa-exclamation-triangle"></i>
        <p>{{ errorSystemServices }}</p>
      </div>

    </main>
    <footer class="modern-footer">
      <div class="footer-content">
        <div class="footer-info">
          <i class="fas fa-cube"></i>
          <span>Docklet</span>
          <span class="footer-separator">•</span>
          <span>智能服务发现平台</span>
        </div>
        <div class="footer-tech">
          Powered by Vue.js & Go
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from "vue";

const services = ref([]); // For Docker services
const loading = ref(true); // For Docker services
const error = ref(null); // For Docker services

const systemWebServices = ref([]);
const loadingSystemServices = ref(true);
const errorSystemServices = ref(null);

// Search and filter states
const searchQuery = ref('');
const dockerSectionCollapsed = ref(false);
const systemSectionCollapsed = ref(false);
const collapsedServicePorts = ref(new Set());

const apiBaseUrl = ""; // Served from the same domain, or use import.meta.env.VITE_API_BASE_URL if configured
const currentHostname = computed(() => window.location.hostname);

// Function to rewrite URLs that point to localhost if accessed via IP/domain
function rewriteServiceUrl(originalUrlString) {
  if (!originalUrlString || typeof originalUrlString !== "string") {
    return originalUrlString;
  }

  // Only process http and https URLs
  if (
    !originalUrlString.startsWith("http://") &&
    !originalUrlString.startsWith("https://")
  ) {
    return originalUrlString;
  }

  try {
    const url = new URL(originalUrlString);
    const currentHostname = window.location.hostname;

    // If original URL's hostname is localhost or 127.0.0.1,
    // and current browser window's hostname is not localhost or 127.0.0.1
    if (
      (url.hostname === "localhost" || url.hostname === "127.0.0.1") &&
      currentHostname !== "localhost" &&
      currentHostname !== "127.0.0.1"
    ) {
      // Replace the URL's hostname with the current browser window's hostname
      url.hostname = currentHostname;
    }
    return url.toString();
  } catch (e) {
    // If URL is invalid or other error occurs, return original string
    console.warn(`Failed to rewrite URL "${originalUrlString}":`, e);
    return originalUrlString;
  }
}

async function fetchDockerServices() {
  loading.value = true;
  error.value = null;
  try {
    const response = await fetch(`${apiBaseUrl}/api/services`);
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    const rawData = await response.json();
    services.value = (rawData || []).map((service) => ({
      ...service,
      url: service.url ? rewriteServiceUrl(service.url) : null,
      icon: service.icon ? rewriteServiceUrl(service.icon) : service.icon,
    }));
  } catch (e) {
    console.error("获取Docker服务失败:", e);
    error.value = `加载Docker服务失败: ${e.message}.`;
    services.value = [];
  } finally {
    loading.value = false;
  }
}

async function fetchSystemWebServices() {
  loadingSystemServices.value = true;
  errorSystemServices.value = null;
  try {
    const response = await fetch(`${apiBaseUrl}/api/system-services`);
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    const services = await response.json() || [];
    systemWebServices.value = services;
    
    // 默认折叠所有服务的端口
    collapsedServicePorts.value.clear();
    services.forEach(service => {
      if (service.listening_ports && service.listening_ports.length > 0) {
        collapsedServicePorts.value.add(service.name);
      }
    });
  } catch (e) {
    console.error("获取本机Web服务失败:", e);
    errorSystemServices.value = `加载本机Web服务失败: ${e.message}.`;
    systemWebServices.value = [];
  } finally {
    loadingSystemServices.value = false;
  }
}

const sortedServices = computed(() => {
  return [...services.value].sort((a, b) => {
    const orderA =
      a.raw_labels && a.raw_labels["docklet.order"]
        ? String(a.raw_labels["docklet.order"])
        : "";
    const orderB =
      b.raw_labels && b.raw_labels["docklet.order"]
        ? String(b.raw_labels["docklet.order"])
        : "";

    if (orderA && !orderB) return -1;
    if (!orderA && orderB) return 1;
    if (orderA && orderB && orderA !== orderB) {
      // Attempt numeric sort if both are numbers, otherwise lexicographical
      const numA = parseFloat(orderA);
      const numB = parseFloat(orderB);
      if (!isNaN(numA) && !isNaN(numB)) {
        return numA - numB;
      }
      return orderA.localeCompare(orderB);
    }
    // Secondary sort: title (case-insensitive)
    const titleA = a.title || a.name || "";
    const titleB = b.title || b.name || "";
    return titleA.localeCompare(titleB, undefined, { sensitivity: "base" });
  });
});

// Search and filter computed properties
const filteredDockerServices = computed(() => {
  if (!searchQuery.value.trim()) return sortedServices.value;
  
  const query = searchQuery.value.toLowerCase();
  return sortedServices.value.filter(service => {
    const title = (service.title || service.name || '').toLowerCase();
    const description = (service.description || '').toLowerCase();
    const containerName = (service.container_name || '').toLowerCase();
    const url = (service.url || '').toLowerCase();
    
    return title.includes(query) || 
           description.includes(query) || 
           containerName.includes(query) ||
           url.includes(query);
  });
});

const filteredSystemServices = computed(() => {
  if (!searchQuery.value.trim()) return systemWebServices.value;
  
  const query = searchQuery.value.toLowerCase();
  return systemWebServices.value.filter(service => {
    const name = (service.name || '').toLowerCase();
    const displayName = (service.display_name || '').toLowerCase();
    const status = (service.status || '').toLowerCase();
    const ports = (service.listening_ports || []).join(' ').toLowerCase();
    
    return name.includes(query) || 
           displayName.includes(query) || 
           status.includes(query) ||
           ports.includes(query);
  });
});

const allSectionsCollapsed = computed(() => {
  return dockerSectionCollapsed.value && systemSectionCollapsed.value;
});

// Helper functions for system service status
function getStatusClass(status) {
  const statusLower = status.toLowerCase();
  if (statusLower.includes('active') || statusLower.includes('running')) {
    return 'status-active';
  } else if (statusLower.includes('inactive') || statusLower.includes('stopped')) {
    return 'status-inactive';
  } else if (statusLower.includes('failed') || statusLower.includes('error')) {
    return 'status-error';
  }
  return 'status-unknown';
}

function getStatusIcon(status) {
  const statusLower = status.toLowerCase();
  if (statusLower.includes('active') || statusLower.includes('running')) {
    return 'fas fa-check-circle';
  } else if (statusLower.includes('inactive') || statusLower.includes('stopped')) {
    return 'fas fa-pause-circle';
  } else if (statusLower.includes('failed') || statusLower.includes('error')) {
    return 'fas fa-exclamation-circle';
  }
  return 'fas fa-question-circle';
}

// Search and toggle methods
function clearSearch() {
  searchQuery.value = '';
}

function toggleDockerSection() {
  dockerSectionCollapsed.value = !dockerSectionCollapsed.value;
}

function toggleSystemSection() {
  systemSectionCollapsed.value = !systemSectionCollapsed.value;
}

function toggleAllSections() {
  const newState = !allSectionsCollapsed.value;
  dockerSectionCollapsed.value = newState;
  systemSectionCollapsed.value = newState;
}

function toggleServicePorts(serviceName) {
  if (collapsedServicePorts.value.has(serviceName)) {
    collapsedServicePorts.value.delete(serviceName);
  } else {
    collapsedServicePorts.value.add(serviceName);
  }
}

function isServicePortsCollapsed(serviceName) {
  return collapsedServicePorts.value.has(serviceName);
}

onMounted(() => {
  fetchDockerServices();
  fetchSystemWebServices();
  // Optional: Refresh services periodically
  // setInterval(fetchDockerServices, 30000);
  // setInterval(fetchSystemWebServices, 60000); // System services might not change as often
});
</script>

<style>
/* App Container */
.app-container {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  /* 确保在内容超出视口时，背景能够完全覆盖 */
  width: 100%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  background-attachment: fixed; /* 固定背景，防止滚动时背景移动 */
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  /* 确保容器能够扩展到内容的实际高度 */
  position: relative;
}

/* Modern Header */
.modern-header {
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(20px);
  border-bottom: 1px solid rgba(255, 255, 255, 0.2);
  padding: 2rem 0;
  text-align: center;
  color: white;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
}

.header-content {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 2rem;
}

.logo-section {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  margin-bottom: 0.5rem;
}

.logo-icon {
  font-size: 2.5rem;
  color: #ffffff;
  text-shadow: 0 0 20px rgba(255, 255, 255, 0.5);
}

.modern-header h1 {
  margin: 0;
  font-size: 2.5rem;
  font-weight: 700;
  background: linear-gradient(45deg, #ffffff, #f0f8ff);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.header-subtitle {
  font-size: 1.1rem;
  opacity: 0.9;
  font-weight: 400;
}

/* Main Content */
.main-content {
  flex: 1;
  padding: 2rem;
  max-width: 1400px;
  margin: 0 auto;
  width: 100%;
  box-sizing: border-box;
}

/* Search Section */
.search-section {
  margin-bottom: 2rem;
}

.search-container {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1rem;
  flex-wrap: wrap;
}

.search-input-wrapper {
  position: relative;
  flex: 1;
  min-width: 300px;
}

.search-input {
  width: 100%;
  padding: 1rem 3rem 1rem 3rem;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(20px);
  border: 2px solid rgba(255, 255, 255, 0.2);
  border-radius: 16px;
  font-size: 1rem;
  color: #2d3748;
  outline: none;
  transition: all 0.3s ease;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.1);
}

.search-input:focus {
  border-color: #667eea;
  box-shadow: 0 15px 40px rgba(102, 126, 234, 0.2);
  transform: translateY(-2px);
}

.search-input::placeholder {
  color: #a0aec0;
}

.search-icon {
  position: absolute;
  left: 1rem;
  top: 50%;
  transform: translateY(-50%);
  color: #718096;
  font-size: 1.1rem;
}

.clear-search-btn {
  position: absolute;
  right: 1rem;
  top: 50%;
  transform: translateY(-50%);
  background: none;
  border: none;
  color: #718096;
  cursor: pointer;
  padding: 0.5rem;
  border-radius: 50%;
  transition: all 0.2s ease;
}

.clear-search-btn:hover {
  background: rgba(113, 128, 150, 0.1);
  color: #4a5568;
}

.view-controls {
  display: flex;
  gap: 0.5rem;
}

.toggle-all-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 1rem 1.5rem;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(20px);
  border: 2px solid rgba(255, 255, 255, 0.2);
  border-radius: 16px;
  color: #4a5568;
  cursor: pointer;
  font-size: 0.9rem;
  font-weight: 500;
  transition: all 0.3s ease;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.1);
}

.toggle-all-btn:hover {
  background: linear-gradient(135deg, #667eea, #764ba2);
  color: white;
  transform: translateY(-2px);
  box-shadow: 0 15px 40px rgba(102, 126, 234, 0.3);
}

.toggle-all-btn.collapsed {
  background: linear-gradient(135deg, #00b894, #00a085);
  color: white;
}

.search-results-info {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.5rem;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(20px);
  border-radius: 12px;
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.search-count {
  color: #4a5568;
  font-weight: 500;
}

.clear-results-btn {
  background: none;
  border: none;
  color: #667eea;
  cursor: pointer;
  font-size: 0.9rem;
  font-weight: 500;
  padding: 0.5rem 1rem;
  border-radius: 8px;
  transition: all 0.2s ease;
}

.clear-results-btn:hover {
  background: rgba(102, 126, 234, 0.1);
  color: #5a67d8;
}

/* Status Cards */
.status-card {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(20px);
  border-radius: 16px;
  padding: 3rem 2rem;
  text-align: center;
  margin: 2rem auto;
  max-width: 500px;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.loading-card {
  background: linear-gradient(135deg, #74b9ff, #0984e3);
  color: white;
}

.error-card {
  background: linear-gradient(135deg, #fd79a8, #e84393);
  color: white;
}

.empty-card {
  background: linear-gradient(135deg, #fdcb6e, #e17055);
  color: white;
}

.status-card i {
  font-size: 3rem;
  margin-bottom: 1rem;
  opacity: 0.8;
}

.status-card h3 {
  margin: 1rem 0 0.5rem 0;
  font-size: 1.5rem;
  font-weight: 600;
}

.status-card p {
  margin: 0;
  font-size: 1.1rem;
  opacity: 0.9;
}

/* Loading Spinner */
.loading-spinner {
  width: 40px;
  height: 40px;
  border: 4px solid rgba(255, 255, 255, 0.3);
  border-top: 4px solid white;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin: 0 auto 1rem auto;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

/* Services Section */
.services-section {
  margin-bottom: 3rem;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 2rem;
  padding: 1.5rem 2rem;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(20px);
  border-radius: 16px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
  cursor: pointer;
  transition: all 0.3s ease;
  user-select: none;
}

.section-header:hover {
  transform: translateY(-2px);
  box-shadow: 0 15px 40px rgba(0, 0, 0, 0.15);
}

.section-icon {
  font-size: 1.8rem;
  color: #667eea;
}

.section-header h2 {
  margin: 0;
  font-size: 1.8rem;
  font-weight: 600;
  color: #2d3748;
  flex: 1;
}

.service-count {
  background: linear-gradient(135deg, #667eea, #764ba2);
  color: white;
  padding: 0.5rem 1rem;
  border-radius: 20px;
  font-size: 0.9rem;
  font-weight: 500;
}

.collapse-btn {
  background: none;
  border: none;
  color: #718096;
  cursor: pointer;
  padding: 0.5rem;
  border-radius: 8px;
  transition: all 0.3s ease;
  margin-left: auto;
}

.collapse-btn:hover {
  background: rgba(113, 128, 150, 0.1);
  color: #4a5568;
}

.collapse-btn i {
  font-size: 1.2rem;
  transition: transform 0.3s ease;
}

.collapse-btn.collapsed i {
  transform: rotate(180deg);
}

/* Services Grid */
.services-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 1.5rem;
}

/* Service Cards */
.service-card {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(20px);
  border-radius: 20px;
  padding: 2rem;
  box-shadow: 0 15px 35px rgba(0, 0, 0, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  display: flex;
  flex-direction: column;
  position: relative;
  overflow: hidden;
}

.service-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 4px;
  background: linear-gradient(90deg, #667eea, #764ba2);
}

.service-card:hover {
  transform: translateY(-8px) scale(1.02);
  box-shadow: 0 25px 50px rgba(0, 0, 0, 0.15);
}

/* Service Header */
.service-header {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.service-icon-wrapper {
  flex-shrink: 0;
}

.service-icon {
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 12px;
  background: linear-gradient(135deg, #667eea, #764ba2);
  color: white;
  font-size: 1.5rem;
}

.service-icon.system-icon {
  background: linear-gradient(135deg, #00b894, #00a085);
}

.service-icon.default-icon {
  background: linear-gradient(135deg, #74b9ff, #0984e3);
}

.service-icon.emoji-icon {
  background: linear-gradient(135deg, #fdcb6e, #e17055);
  font-size: 1.8rem;
}

.service-icon-image {
  width: 32px;
  height: 32px;
  object-fit: contain;
  border-radius: 6px;
}

.service-info {
  flex: 1;
  min-width: 0;
}

.service-title {
  margin: 0 0 0.5rem 0;
  font-size: 1.3rem;
  font-weight: 600;
  color: #2d3748;
  line-height: 1.3;
}

.service-description {
  margin: 0;
  color: #718096;
  font-size: 0.95rem;
  line-height: 1.4;
}

/* Service Meta */
.service-meta {
  display: flex;
  align-items: center;
  gap: 1rem;
  flex-wrap: wrap;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.4rem 0.8rem;
  border-radius: 12px;
  font-size: 0.8rem;
  font-weight: 500;
  text-transform: capitalize;
}

.status-active {
  background: rgba(72, 187, 120, 0.1);
  color: #2f855a;
  border: 1px solid rgba(72, 187, 120, 0.2);
}

.status-inactive {
  background: rgba(160, 174, 192, 0.1);
  color: #4a5568;
  border: 1px solid rgba(160, 174, 192, 0.2);
}

.status-error {
  background: rgba(245, 101, 101, 0.1);
  color: #c53030;
  border: 1px solid rgba(245, 101, 101, 0.2);
}

.status-unknown {
  background: rgba(237, 137, 54, 0.1);
  color: #c05621;
  border: 1px solid rgba(237, 137, 54, 0.2);
}

.pid-info {
  color: #718096;
  font-size: 0.85rem;
  background: rgba(160, 174, 192, 0.1);
  padding: 0.3rem 0.6rem;
  border-radius: 8px;
}

/* Service Actions */
.service-actions {
  margin-top: auto;
}

.service-button {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.8rem 1.5rem;
  background: linear-gradient(135deg, #667eea, #764ba2);
  color: white;
  text-decoration: none;
  border-radius: 12px;
  font-weight: 500;
  font-size: 0.9rem;
  transition: all 0.3s ease;
  border: none;
  cursor: pointer;
  box-shadow: 0 4px 15px rgba(102, 126, 234, 0.3);
}

.service-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(102, 126, 234, 0.4);
}

.service-button.primary {
  background: linear-gradient(135deg, #667eea, #764ba2);
}

.service-button.port-button {
  background: linear-gradient(135deg, #00b894, #00a085);
  margin: 0.25rem 0.5rem 0.25rem 0;
  padding: 0.6rem 1rem;
  font-size: 0.85rem;
}

.service-status {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.8rem 1rem;
  border-radius: 12px;
  font-size: 0.9rem;
}

.service-status.offline {
  background: rgba(245, 101, 101, 0.1);
  color: #c53030;
  border: 1px solid rgba(245, 101, 101, 0.2);
}

.service-status.no-ports {
  background: rgba(160, 174, 192, 0.1);
  color: #4a5568;
  border: 1px solid rgba(160, 174, 192, 0.2);
}

/* Port List */
.port-list {
  margin-top: 1rem;
}

.port-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 0.9rem;
  color: #4a5568;
  margin-bottom: 0.8rem;
  font-weight: 500;
  cursor: pointer;
  padding: 0.75rem 1rem;
  background: rgba(113, 128, 150, 0.05);
  border-radius: 12px;
  transition: all 0.3s ease;
  user-select: none;
}

.port-header:hover {
  background: rgba(113, 128, 150, 0.1);
  transform: translateY(-1px);
}

.port-count-badge {
  background: linear-gradient(135deg, #00b894, #00a085);
  color: white;
  padding: 0.3rem 0.8rem;
  border-radius: 16px;
  font-size: 0.8rem;
  font-weight: 500;
  margin-left: auto;
  margin-right: 1rem;
}

.port-toggle-btn {
  background: none;
  border: none;
  color: #718096;
  cursor: pointer;
  padding: 0.3rem;
  border-radius: 6px;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  justify-content: center;
}

.port-toggle-btn:hover {
  background: rgba(113, 128, 150, 0.1);
  color: #4a5568;
}

.port-toggle-btn i {
  font-size: 1rem;
  transition: transform 0.3s ease;
}

.port-toggle-btn.collapsed i {
  transform: rotate(180deg);
}

.port-buttons {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  transition: all 0.3s ease;
  overflow: hidden;
}

/* Modern Footer */
.modern-footer {
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(20px);
  border-top: 1px solid rgba(255, 255, 255, 0.2);
  padding: 2rem 0;
  margin-top: 2rem;
}

.footer-content {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 2rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  color: white;
}

.footer-info {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-weight: 500;
}

.footer-separator {
  opacity: 0.6;
}

.footer-tech {
  font-size: 0.9rem;
  opacity: 0.8;
}

/* Responsive Design */
@media (max-width: 768px) {
  .main-content {
    padding: 1rem;
  }
  
  .services-grid {
    grid-template-columns: 1fr;
  }
  
  .modern-header h1 {
    font-size: 2rem;
  }
  
  .logo-section {
    flex-direction: column;
    gap: 0.5rem;
  }
  
  .section-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
    text-align: left;
  }
  
  .service-card {
    padding: 1.5rem;
  }
  
  .footer-content {
    flex-direction: column;
    gap: 1rem;
    text-align: center;
  }
  
  .port-buttons {
    justify-content: center;
  }
  
  .search-container {
    flex-direction: column;
    align-items: stretch;
  }
  
  .search-input-wrapper {
    min-width: auto;
  }
  
  .view-controls {
    justify-content: center;
  }
  
  .search-results-info {
    flex-direction: column;
    gap: 1rem;
    text-align: center;
  }
}

@media (max-width: 480px) {
  .header-content {
    padding: 0 1rem;
  }
  
  .modern-header {
    padding: 1.5rem 0;
  }
  
  .modern-header h1 {
    font-size: 1.8rem;
  }
  
  .service-header {
    flex-direction: column;
    align-items: center;
    text-align: center;
    gap: 1rem;
  }
  
  .service-meta {
    justify-content: center;
  }
}
</style>
