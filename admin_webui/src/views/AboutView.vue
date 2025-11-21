<template>
  <div class="about-container">
    <el-card class="box-card">
      <template #header>
        <div class="card-header">
          <span>{{ t('about.title') }}</span>
        </div>
      </template>
      <div class="content">
        <p><strong>{{ t('appName') }} {{ t('about.version', { version: '1.0.0' }) }}</strong></p>
        <p>{{ t('about.description') }}</p>
        
        <p>
          <strong>{{ t('about.organization') }}</strong><br>
          {{ t('about.organizationName') }}<br>
          <span style="color: #666; font-size: 0.9em;">{{ t('about.organizationDescription') }}</span>
        </p>
        
        <p>
          <strong>{{ t('about.author') }}</strong><br>
          CloudBridge Research Center
        </p>
        
        <p>
          <strong>Links:</strong><br>
          <a href="https://github.com/twogc/masque-vpn" target="_blank" rel="noopener noreferrer">
            {{ t('about.projectLink') }}
          </a><br>
          <a href="https://cloudbridge-research.ru" target="_blank" rel="noopener noreferrer">
            {{ t('about.websiteLink') }}
          </a>
        </p>
        
        <p>
          {{ t('about.techStack') }}
          <ul>
            <li>Vue 3</li>
            <li>Vite</li>
            <li>Element Plus</li>
            <li>Pinia</li>
            <li>TypeScript</li>
            <li>vue-i18n</li>
            <li>Go (backend)</li>
            <li>QUIC / MASQUE Protocol</li>
          </ul>
        </p>
      </div>
    </el-card>
  </div>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
</script>

<style scoped>
.about-container {
  padding: 20px;
}
.card-header span {
  font-size: 18px;
  font-weight: bold;
}
.content {
  line-height: 1.8;
}
.content p {
  margin-bottom: 10px;
}
.content ul {
  list-style-type: disc;
  padding-left: 20px;
}
.content a {
  color: #409EFF;
  text-decoration: none;
}
.content a:hover {
  text-decoration: underline;
}
</style>
