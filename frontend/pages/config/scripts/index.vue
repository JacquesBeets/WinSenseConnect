<template>
  <div class="p-6">
    <h1>Scripts</h1>

      <template v-if="scripts.length === 0">
        <p>No scripts found.</p>
      </template>

      <template v-else>
        <table class="table w-full">
          <thead>
            <tr>
              <th>MQTT Command Name</th>
              <th>Path</th>
              <th>Run as user</th>
              <th>Timeout</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="script in scripts" :key="script.id">
              <td>{{ script.name }}</td>
              <td>{{ script.script_path }}</td>
              <td>{{ script.run_as_user }}</td>
              <td>{{ script.script_timeout }}</td>
              <td>
                <nuxt-link  :to="`/config/scripts/${script.id}`" class="btn btn-primary cursor-pointer">Edit</nuxt-link>  
              </td>
            </tr>
          </tbody>
        </table>
      </template>
  </div>
</template>

<script setup>
const { $toast } = useNuxtApp()
const isSaving = ref(false)

const scripts = ref([])

const { data: scriptsData } = await useFetch('http://localhost:8077/api/scripts')
if (scriptsData) {
  scripts.value = JSON.parse(scriptsData.value)
  console.log(scripts.value)
} else {
  console.error('Failed to fetch configuration')
  $toast.error('Failed to load configuration')
}

const saveConfig = async () => {
  isSaving.value = true
  try {
    const { error: saveError } = await useFetch('http://localhost:8077/api/config', {
      method: 'POST',
      body: scripts.value
    })

    if (saveError.value) {
      throw new Error('Failed to save configuration')
    }

    $toast.success('Configuration saved successfully, Restarting service...')

    // Restart the service
    const { error: restartError } = await useFetch('http://localhost:8077/api/restart', {
      method: 'POST'
    })

    if (restartError.value) {
      throw new Error('Failed to restart service')
    }

    $toast.success('Service restarted successfully')
  } catch (error) {
    console.error('Error:', error)
    $toast.error(error.message)
  } finally {
    isSaving.value = false
  }
}
</script>