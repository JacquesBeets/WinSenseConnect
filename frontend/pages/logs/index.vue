<template>
  <div class="container mx-auto px-4 py-8">
    <h1 class="text-3xl font-bold mb-4">Logs</h1>
    <div v-if="error" class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative mb-4" role="alert">
      <strong class="font-bold">Error:</strong>
      <span class="block sm:inline">{{ error }}</span>
    </div>
    <div v-if="displayedLogs.length === 0" class="bg-yellow-100 border border-yellow-400 text-yellow-700 px-4 py-3 rounded relative mb-4">
      No logs available at the moment.
    </div>
    <div v-else class="bg-secondary-400 shadow-md rounded-lg overflow-hidden">
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-secondary-300">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-black uppercase tracking-wider">Timestamp</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-black uppercase tracking-wider">Level</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-black uppercase tracking-wider">Message</th>
            </tr>
          </thead>
          <tbody class="bg-secondary-500 divide-y divide-secondary-200">
            <tr v-for="(log, index) in displayedLogs" :key="index" :class="{'bg-secondary-400': index % 2 === 0}">
              <td class="px-6 py-4 whitespace-nowrap text-sm text-black">{{ formatDate(log.timestamp) }}</td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span :class="getLevelClass(log.level)" class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full">
                  {{ log.level }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-black">{{ log.message }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue';
import { useFetch } from '#app';

const { $subscribeToLogs } = useNuxtApp();

const allLogs = ref([]);
const newLogs = ref([]);
const error = ref(null);

const unsubscribe = ref(null);

const API_BASE_URL = 'http://localhost:8077'; // Adjust this to match your API server address

const fetchLogs = async () => {
  const { data, error: fetchError } = await useFetch(`${API_BASE_URL}/api/logs`, {
    method: 'GET',
    headers: {
      'Accept': 'application/json',
    },
  });

  if (fetchError.value) {
    console.error('Failed to fetch logs:', fetchError.value);
    error.value = `Failed to fetch logs: ${fetchError.value.message}`;
    return;
  }
  
  if (!data.value) {
    console.error('Unexpected API response:', data.value);
    error.value = 'Received unexpected data format from the server.';
    return;
  }

  console.log('API Response:', data.value); // Log the entire response for debugging

  if (typeof data.value === 'string') {
    try {
      // If the response is a string, try to parse it as JSON
      const parsedData = JSON.parse(data.value);
      if (parsedData && parsedData.logs !== undefined) {
        data.value = parsedData;
      } else {
        throw new Error('Invalid JSON structure');
      }
    } catch (jsonError) {
      console.error('Error parsing JSON:', jsonError);
      error.value = 'Error processing log data: Invalid JSON';
      return;
    }
  }

  if (data.value.logs === undefined) {
    console.error('Logs property missing in API response:', data.value);
    error.value = 'Received unexpected data format from the server.';
    return;
  }

  if (data.value.logs === "") {
    console.log('No logs available');
    allLogs.value = [];
    return;
  }

  try {
    // Parse the logs string into an array of log objects
    const logsArray = data.value.logs.split('\n').filter(Boolean).map(logLine => {
      const [timestamp, level, ...messageParts] = logLine.split(' ');
      return {
        timestamp: new Date(timestamp).getTime(),
        level,
        message: messageParts.join(' ')
      };
    });

    allLogs.value = logsArray.reverse(); // Reverse to show newest first
  } catch (parseError) {
    console.error('Error parsing logs:', parseError);
    error.value = 'Error processing log data.';
  }
};

onMounted(async () => {
  await fetchLogs();

  if ($subscribeToLogs) {
    unsubscribe.value = $subscribeToLogs((logEvent) => {
      newLogs.value.unshift(logEvent);
      if (newLogs.value.length > 100) {
        newLogs.value.pop();
      }
    });
  } else {
    console.error('$subscribeToLogs is not available');
    error.value = 'Real-time log updates are not available.';
  }
});

onUnmounted(() => {
  if (unsubscribe.value) {
    unsubscribe.value();
  }
});

const displayedLogs = computed(() => {
  return [...newLogs.value, ...allLogs.value];
});

const formatDate = (timestamp) => {
  const date = new Date(timestamp);
  return date.toLocaleString();
};

const getLevelClass = (level) => {
  switch (level.toLowerCase()) {
    case 'debug':
      return 'bg-blue-100 text-blue-800';
    case 'error':
      return 'bg-red-100 text-red-800';
    default:
      return 'bg-gray-100 text-gray-800';
  }
};
</script>
