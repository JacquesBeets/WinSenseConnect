<template>
  <h1>Scripts ID Page</h1>
</template>

<script setup>
const { $toast } = useNuxtApp()

const config = ref({})
const isSaving = ref(false)


const { data: configData } = await useFetch('http://localhost:8077/api/scripts')
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