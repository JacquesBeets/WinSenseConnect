<template>
  <form  class="max-w-3xl mx-auto">
    <div class="form-control">
      <label for="scriptId">Script ID</label>
      <input type="text" id="scriptId" v-model="script.id" disabled/>
    </div>
    <div class="form-control">
      <label for="name">MQTT Command Name</label>
      <input type="text" id="name" v-model="script.name" />
    </div>
    <div class="form-control">
      <label for="scriptPath">Script Path</label>
      <input type="text" id="scriptPath" disabled v-model="script.script_path" />
    </div>
    <div class="form-control">
      <label for="scriptTimeout">Script Timeout</label>
      <input type="number" id="scriptTimeout" v-model="script.script_timeout" />
    </div>
    <div class="form-control mt-6">
      <div class="flex items-center me-6">
        <input v-model="script.run_as_user" checked id="primary-checkbox" type="checkbox" value="" class="w-6 h-6 text-primary-400 bg-gray-100 border-gray-300 rounded focus:ring-primary-500 dark:focus:ring-primary-600 dark:ring-offset-gray-800 focus:ring-2 dark:bg-gray-700 dark:border-gray-600">
        <label for="primary-checkbox" class="ms-2 text-sm font-medium text-gray-900 dark:text-gray-300" >Run as User</label>
      </div>
    </div>
    <div class="form-control">
      <button @click.stop="saveConfig" class="btn-primary ml-auto" :disabled="isSaving">
        {{ isSaving ? 'Saving...' : 'Save' }}
      </button>
    </div>
  </form>
</template>

<script setup>
const { $toast } = useNuxtApp()

const route = useRoute()
const id = route.params.id
const script = ref({
    "id": 1,
    "name": "test_notification",
    "script_path": "test_notification.ps1",
    "run_as_user": true,
    "script_timeout": 300,
    "created_at": "2023-07-01T12:00:00Z",
    "updated_at": "2023-07-01T12:00:00Z"
})
const isSaving = ref(false)


const { data: scriptData } = await useFetch(`http://localhost:8077/api/scripts/${id}`)
if (scriptData.value) {
  script.value = JSON.parse(scriptData.value)
  console.log("script.value", script.value)
} else {
  console.error('Failed to fetch configuration')
  $toast.error('Failed to load configuration')
}

const saveConfig = async () => {
  isSaving.value = true
  try {
    const { error: saveError } = await useFetch('http://localhost:8077/api/config', {
      method: 'POST',
      body: script.value
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