<template>
  <div>MQTT</div>
  <form @submit.prevent="saveConfig">
    <div class="form-control">
      <label for="brokerAddress">Broker Address</label>
      <input type="text" id="brokerAddress" v-model="config.brokerAddress" />
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
      <input type="text" id="clientID" v-model="config.clientID" />
    </div>
    <div class="form-control">
      <label for="topic">Topic</label>
      <input type="text" id="topic" v-model="config.topic" />
    </div>
    <div class="form-control">
      <label for="logLevel">Log Level</label>
      <select id="logLevel" v-model="config.logLevel">
        <option value="debug">Debug</option>
        <option value="info">Info</option>
        <option value="warn">Warn</option>
        <option value="error">Error</option>
      </select>
    </div>
    <div class="form-control">
      <label for="scriptTimeout">Script Timeout</label>
      <input type="number" id="scriptTimeout" v-model="config.scriptTimeout" />
    </div>
    <div class="form-control">
      <button type="submit">Save</button>
    </div>
  </form>
</template>

<script setup>
  const config = ref({})  
  console.log(config.value)
  
  const saveConfig = () => {
    store.dispatch('saveConfig', config.value)
  }
</script>