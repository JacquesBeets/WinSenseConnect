<template>
  <form @submit.prevent="saveConfig" class="max-w-3xl mx-auto">
    <h1 class="text-3xl font-bold mb-6">MQTT Configuration</h1>
    <div class="form-control">
      <label for="brokerAddress">Broker Address</label>
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
      <label for="clientID">Client ID</label>
      <input type="text" id="clientID" v-model="config.client_id" />
    </div>
    <div class="form-control">
      <label for="topic">Topic</label>
      <input type="text" id="topic" v-model="config.topic" />
    </div>
    <div class="form-control">
      <label for="logLevel">Log Level</label>
      <select id="logLevel" v-model="config.log_level">
        <option value="none">No Logs</option>
        <option value="debug">Debug</option>
        <option value="error">Error</option>
      </select>
    </div>
    <div class="form-control">
      <label for="scriptTimeout">Script Timeout</label>
      <input type="number" id="scriptTimeout" v-model="config.script_timeout" />
    </div>
    <div class="form-control">
      <button class="btn-primary ml-auto" type="submit">Save</button>
    </div>
  </form>
</template>

<script setup>
  const config = ref({})  

  const {data: configData} = await useFetch('http://localhost:8077/api/config')
  config.value = JSON.parse(configData.value)

  console.log(config.value)
  
  const saveConfig = () => {
    const data = JSON.stringify(config.value)
    useFetch('http://localhost:8077/api/config', {
      method: 'POST',
      body: data
    })
  }
</script>