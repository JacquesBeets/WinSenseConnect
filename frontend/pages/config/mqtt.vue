<template>
  <form  class="max-w-3xl mx-auto">
    <h1 class="text-3xl font-bold mb-6">MQTT Configuration</h1>
    <div class="form-control">
      <label for="brokerAddress">Broker IP Address and port</label>
      <input type="text" id="brokerAddress" v-model="config.broker_address" />
    </div>
    <div class="form-control">
      <label for="username">Username</label>
      <input type="text" id="username" v-model="config.username" />
    </div>
    <div class="form-control">
      <label for="password">Password</label>
      <input type="password" id="password" v-model="config.password" />
    </div>
    <div class="form-control">
      <label for="clientID">Client ID <small class="opacity-30">(Must be unique and identifiable)</small></label>
      <input type="text" id="clientID" v-model="config.client_id" />
    </div>
    <div class="form-control">
      <label for="topic">Topic <small class="opacity-30">(eg: winsense/{{ config.topic }}/{{ config.client_id }})</small></label>
      <input type="text" id="topic" v-model="config.topic" />
    </div>
    <div class="form-control">
      <label for="logLevel">Log Level</label>
      <select id="logLevel" v-model="config.log_level">
        <option value="none">No Logs</option>
        <option value="debug">Debug - Logs everything</option>
        <option value="errors">Log Errors Only</option>
      </select>
    </div>
    <div class="form-control">
      <label for="scriptTimeout">Script Timeout</label>
      <input type="number" id="scriptTimeout" v-model="config.script_timeout" />
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

const config = ref({})
const isSaving = ref(false)


const { data: configData } = await useFetch('http://localhost:8077/api/config')
if (configData.value) {
  config.value = JSON.parse(configData.value)
  console.log(config.value)
} else {
  console.error('Failed to fetch configuration')
  $toast.error('Failed to load configuration')
}

const saveConfig = async () => {
  isSaving.value = true
  try {
    const { error: saveError } = await useFetch('http://localhost:8077/api/config', {
      method: 'POST',
      body: config.value
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