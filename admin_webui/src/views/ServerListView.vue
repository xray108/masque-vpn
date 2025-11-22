<template>
  <div class="server-list-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ t('clientManagement.title') }}</span>
          <div>
            <el-input
              v-model="searchQuery"
              :placeholder="t('clientManagement.searchPlaceholder')"
              clearable
              style="width: 250px; margin-right: 10px"
              @clear="fetchClients"
              @keyup.enter="fetchClients"
            >
              <template #append>
                <el-button :icon="Search" @click="fetchClients" />
              </template>
            </el-input>
            <el-button type="primary" :icon="Plus" @click="openGenerateClientDialog">{{ t('clientManagement.addClient') }}</el-button>
            <el-button :icon="Refresh" @click="handleRefresh" style="margin-left:10px;">{{ t('actions.refresh') }}</el-button>
          </div>
        </div>
      </template>

      <el-table :data="filteredClients" v-loading="loading" stripe style="width: 100%">
        <el-table-column prop="client_name" :label="t('clientManagement.table.clientName')" sortable />
        <el-table-column prop="created_at" :label="t('clientManagement.table.createdAt')" sortable />
        <el-table-column :label="t('clientManagement.table.groups')" min-width="150">
          <template #default="{ row }">
            <el-tooltip
              v-if="getClientGroupNames(row.group_ids).length"
              effect="dark"
              :content="getClientGroupNames(row.group_ids).join(', ')"
              placement="top"
            >
              <el-tag
                v-for="groupName in getClientGroupNames(row.group_ids).slice(0, 2)"
                :key="groupName"
                size="small"
                style="margin-right: 5px;"
              >
                {{ groupName }}
              </el-tag>
            </el-tooltip>
            <span v-if="getClientGroupNames(row.group_ids).length > 2">...</span>
            <span v-if="!getClientGroupNames(row.group_ids).length">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="online" :label="t('clientManagement.table.onlineStatus')" sortable>
          <template #default="{ row }">
            <el-tag :type="row.online ? 'success' : 'danger'">
              {{ row.online ? t('clientManagement.table.online') : t('clientManagement.table.offline') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('clientManagement.table.actions')" width="250">
          <template #default="{ row }">
            <el-button size="small" type="success" :icon="Download" @click="downloadClientConfig(row.client_id, row.client_name)">
              {{ t('clientManagement.downloadConfig') }}
            </el-button>
            <el-button size="small" type="danger" :icon="Delete" @click="confirmDeleteClient(row.client_id, row.client_name)">
              {{ t('clientManagement.deleteClient') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="generateDialog.visible" :title="t('clientManagement.generateClientTitle')" width="500px" @closed="resetGenerateForm">
      <el-form ref="generateFormRef" :model="generateForm" :rules="generateFormRules" label-width="150px">
        <el-form-item :label="t('clientManagement.form.clientName')" prop="client_name">
          <el-input v-model="generateForm.client_name" :placeholder="t('clientManagement.form.clientNamePlaceholder')" />
        </el-form-item>
        <el-form-item :label="t('clientManagement.form.serverAddress')" prop="server_addr">
          <el-input v-model="generateForm.server_addr" :placeholder="t('clientManagement.form.serverAddressPlaceholder')" />
        </el-form-item>
        <el-form-item :label="t('clientManagement.form.serverName')" prop="server_name">
          <el-input v-model="generateForm.server_name" :placeholder="t('clientManagement.form.serverNamePlaceholder')" />
        </el-form-item>
        
        <el-divider content-position="left">FEC Configuration</el-divider>
        
        <el-form-item label="Enable FEC" prop="fec_enabled">
          <el-switch v-model="generateForm.fec_enabled" />
        </el-form-item>
        
        <el-form-item label="Redundancy (%)" prop="fec_redundancy" v-if="generateForm.fec_enabled">
          <el-input-number v-model="generateForm.fec_redundancy" :min="0" :max="100" />
        </el-form-item>
        
        <el-form-item label="Block Size" prop="fec_block_size" v-if="generateForm.fec_enabled">
          <el-input-number v-model="generateForm.fec_block_size" :min="1" :max="100" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="generateDialog.visible = false">{{ t('actions.cancel') }}</el-button>
        <el-button type="primary" @click="handleGenerateClient" :loading="generateDialog.loading">
          {{ t('clientManagement.generateClientButton') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script lang="ts" setup>
import { ref, reactive, onMounted, computed } from 'vue';
import apiClient from '@/api';
import { ElMessage, ElMessageBox, FormInstance, FormRules } from 'element-plus';
import { Plus, Download, Delete, Refresh, Search } from '@element-plus/icons-vue';
import { saveAs } from 'file-saver';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

interface Client {
  client_id: string;
  client_name: string;
  created_at: string;
  online: boolean;
  group_ids: string[];
}

interface Group {
  group_id: string;
  group_name: string;
}

interface ServerConfigFromAPI {
  server_addr: string;
  server_name: string;
  mtu: number;
}

const clients = ref<Client[]>([]);
const allGroups = ref<Group[]>([]);
const loading = ref(false);
const searchQuery = ref('');
const defaultServerSettings = ref<{ addr: string; name: string } | null>(null);

const generateDialog = reactive({
  visible: false,
  loading: false,
});
const generateFormRef = ref<FormInstance>();
const generateForm = reactive({
  client_name: '',
  server_addr: '',
  server_name: '',
  fec_enabled: false,
  fec_redundancy: 10,
  fec_block_size: 20,
});

const generateFormRules = reactive<FormRules>({
  client_name: [
    { required: true, message: () => t('clientManagement.validation.clientNameRequired'), trigger: 'blur' },
  ],
});

// ... (existing code) ...

const openGenerateClientDialog = async () => {
  await fetchDefaultServerSettings();
  if (defaultServerSettings.value) {
    generateForm.server_addr = defaultServerSettings.value.addr;
    generateForm.server_name = defaultServerSettings.value.name;
  } else {
    generateForm.server_addr = '';
    generateForm.server_name = '';
  }
  generateForm.client_name = '';
  generateForm.fec_enabled = false;
  generateForm.fec_redundancy = 10;
  generateForm.fec_block_size = 20;
  generateDialog.visible = true;
};

const resetGenerateForm = () => {
  generateFormRef.value?.resetFields();
  generateForm.client_name = '';
  generateForm.server_addr = defaultServerSettings.value?.addr || '';
  generateForm.server_name = defaultServerSettings.value?.name || '';
  generateForm.fec_enabled = false;
  generateForm.fec_redundancy = 10;
  generateForm.fec_block_size = 20;
};

const handleGenerateClient = async () => {
  if (!generateFormRef.value) return;
  await generateFormRef.value.validate(async (valid) => {
    if (valid) {
      generateDialog.loading = true;
      try {
        const params: Record<string, string> = {
          client_name: generateForm.client_name,
          fec_enabled: String(generateForm.fec_enabled),
          fec_redundancy: String(generateForm.fec_redundancy),
          fec_block_size: String(generateForm.fec_block_size),
        };
        if (generateForm.server_addr) params.server_addr = generateForm.server_addr;
        if (generateForm.server_name) params.server_name = generateForm.server_name;

        const apiResponse:any = await apiClient.post<{ client_id: string; client_name: string }>('/clients/gen_v2', null, { params });
        const responseData = apiResponse;

        if (responseData && responseData.client_id) {
          ElMessage.success(t('clientManagement.generateSuccess', { clientName: responseData.client_name }));
          generateDialog.visible = false;
          await fetchClients();
        } else {
          ElMessage.error(t('clientManagement.generateFail'));
        }
      } catch (error: any) {
        const apiError = error?.response?.data?.error || error?.message;
        ElMessage.error(apiError || t('clientManagement.generateError'));
        console.error('Error generating client:', error);
      } finally {
        generateDialog.loading = false;
      }
    }
  });
};

const downloadClientConfig = async (clientId: string, clientName: string) => {
  try {
    const response:any = await apiClient.get<string>(`/download_client?id=${clientId}`, {
      responseType: 'text',
    });
    const configContent = response;

    if (typeof configContent === 'string') {
      const blob = new Blob([configContent], { type: 'text/plain;charset=utf-8' });
      saveAs(blob, `${clientName || clientId}.toml`);
      ElMessage.success(t('clientManagement.downloadSuccess'));
    } else {
      ElMessage.error(t('clientManagement.downloadError'));
    }
  } catch (error: any) {
    const apiError = error?.response?.data?.error || error?.message;
    ElMessage.error(apiError || t('clientManagement.downloadError'));
    console.error('Error downloading client config:', error);
  }
};

const confirmDeleteClient = (clientId: string, clientName: string) => {
  ElMessageBox.confirm(
    t('clientManagement.confirmDeleteClientMessage', { clientId: clientName || clientId }),
    t('clientManagement.confirmDeleteClientTitle'),
    {
      confirmButtonText: t('actions.delete'),
      cancelButtonText: t('actions.cancel'),
      type: 'warning',
    }
  ).then(async () => {
    try {
      await apiClient.post(`/delete_client?id=${clientId}`);
      ElMessage.success(t('clientManagement.deleteSuccess', { clientId: clientName || clientId }));
      await fetchClients();
    } catch (error: any) {
      const apiError = error?.response?.data?.error || error?.message;
      ElMessage.error(apiError || t('clientManagement.deleteError'));
      console.error('Error deleting client:', error);
    }
  }).catch(() => {
    ElMessage.info(t('clientManagement.deleteCancelled'));
  });
};

const handleRefresh = async () => {
  loading.value = true;
  try {
    await fetchAllGroups();
    await fetchClients();
    // ElMessage.success(t('clientManagement.refreshSuccess'));
  } catch (error) {
    console.error("Error during refresh:", error);
    ElMessage.error(t('clientManagement.fetchError'));
  } finally {
    loading.value = false;
  }
};

onMounted(async () => {
  await handleRefresh();
});
</script>

<style scoped>
.server-list-container {
  padding: 20px;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
