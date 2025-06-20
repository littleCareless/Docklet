<template>
  <div style="display: flex; flex-direction: column; height: 100%;">
    <header>
      <h1>Docklet 服务面板</h1>
    </header>
    <main style="flex-grow: 1; overflow-y: auto; padding: 20px;">
      <div
        v-if="loading"
        id="loading-message"
    >
      正在加载服务...
    </div>
    <div
      v-if="error"
      id="error-message"
      style="color: red"
    >
      {{ error }}
    </div>
    <div
      v-if="!loading && !error && services.length === 0"
      id="no-services-message"
    >
      未找到可用的服务。
    </div>
    <div
      v-if="!loading && !error && services.length > 0"
      id="services-grid"
      class="services-grid"
    >
      <div
        v-for="service in sortedServices"
        :key="service.id"
        class="service-card"
      >
        <h2>
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
              class="icon-image"
            />
          </span>
          <span
            v-else-if="service.icon && service.icon.startsWith('fa-')"
            class="icon"
          >
            <i :class="['fas', service.icon]"></i>
          </span>
          <span
            v-else-if="service.icon"
            class="icon"
            >{{ service.icon }}</span
          >
          {{ service.title || service.name || "未知服务" }}
        </h2>
        <p>{{ service.description || `容器名: ${service.container_name}` }}</p>
        <a
          v-if="service.url"
          :href="service.url"
          target="_blank"
          class="service-link"
          >访问服务</a
        >
        <p v.else>无可用访问链接</p>
      </div>
      </div>
    </main>
    <footer>
      <p>Docklet - 自动发现您的 Docker 服务</p>
    </footer>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from "vue";

const services = ref([]);
const loading = ref(true);
const error = ref(null);
const apiBaseUrl = ""; // Served from the same domain, or use import.meta.env.VITE_API_BASE_URL if configured

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

async function fetchServices() {
  loading.value = true;
  error.value = null;
  try {
    // apiBaseUrl is empty, so fetch will use relative path /api/services
    // Vite proxy will handle forwarding this to the backend
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
    console.error("获取服务失败:", e);
    error.value = `加载服务失败: ${e.message}. 请确保后端服务正在运行并且 Docker Socket 可访问。`;
    services.value = []; // Clear services on error
  } finally {
    loading.value = false;
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

onMounted(() => {
  fetchServices();
  // Optional: Refresh services periodically
  // setInterval(fetchServices, 30000);
});
</script>

<style>
/* body styles are now in frontend/src/style.css */

header {
  background-color: #333; /* Kept for specific header background */
  color: #fff; /* Kept for specific header text color */
  padding: 1rem 0;
  text-align: center;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  flex-shrink: 0; /* Prevent header from shrinking */
}

header h1 {
  margin: 0;
  font-size: 1.8rem;
}

/* main styles are now mostly inline, max-width and margin:auto removed */
/* padding is applied inline to the main element */

.services-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 20px;
}

.service-card {
  background-color: #fff;
  border: 1px solid #ddd;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 5px rgba(0, 0, 0, 0.05);
  transition: transform 0.2s ease-in-out, box-shadow 0.2s ease-in-out;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.service-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 4px 10px rgba(0, 0, 0, 0.1);
}

.service-card h2 {
  margin-top: 0;
  margin-bottom: 10px;
  font-size: 1.4rem;
  color: #007bff;
  display: flex; /* For icon alignment */
  align-items: center; /* For icon alignment */
}

.service-card .icon {
  font-size: 1.2em;
  margin-right: 8px;
  display: inline-block; /* Ensure span takes space */
}
.service-card .icon-image {
  width: 24px;
  height: 24px;
  margin-right: 8px;
  vertical-align: middle; /* Better alignment with text */
  object-fit: contain; /* Prevents distortion */
}

.service-card p {
  margin-bottom: 15px;
  font-size: 0.95rem;
  color: #555;
  flex-grow: 1;
}

.service-card a.service-link {
  display: inline-block;
  background-color: #007bff;
  color: white;
  padding: 10px 15px;
  text-decoration: none;
  border-radius: 5px;
  text-align: center;
  transition: background-color 0.2s ease;
  margin-top: auto; /* Pushes link to the bottom if card content is short */
}

.service-card a.service-link:hover {
  background-color: #0056b3;
}

#loading-message,
#error-message,
#no-services-message {
  text-align: center;
  font-size: 1.2rem;
  padding: 20px;
}

footer {
  text-align: center;
  padding: 20px;
  background-color: #333; /* Kept for specific footer background */
  color: #fff; /* Kept for specific footer text color */
  /* margin-top: 40px; Removed as flex layout handles spacing */
  flex-shrink: 0; /* Prevent footer from shrinking */
}

/* Responsive adjustments */
@media (max-width: 600px) {
  .services-grid {
    grid-template-columns: 1fr;
  }

  header h1 {
    font-size: 1.5rem;
  }

  .service-card h2 {
    font-size: 1.2rem;
  }
}
</style>
