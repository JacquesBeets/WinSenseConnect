<template>
  <div class="flex flex-col gap-2">
    <NuxtLink to="/" class="text-lg">
      <Icon name="material-symbols:dashboard" class="text-primary-500"/> Dashboard
    </NuxtLink>
    <!-- <NuxtLink to="/settings" class="text-lg">
      <Icon name="material-symbols:sensors" class="text-primary-500" /> Sensors
    </NuxtLink> -->
    <NuxtLink to="/config/scripts" class="text-lg">
      <Icon name="material-symbols:settings-ethernet-rounded" class="text-primary-500" /> Scripts
    </NuxtLink>
    <NuxtLink to="/config/mqtt" class="text-lg">
      <Icon name="material-symbols:settings-b-roll" class="text-primary-500" /> MQTT
    </NuxtLink>
    <!-- <NuxtLink to="/config/home-assistant" class="text-lg">
      <Icon name="material-symbols:format-list-numbered" class="text-primary-500" /> Home Asssistant
    </NuxtLink> -->
    <NuxtLink to="/logs" class="text-lg">
      <Icon name="material-symbols:format-list-numbered" class="text-primary-500" /> LOGS
    </NuxtLink>
  </div>
</template>

<style scoped>
a {
  @apply py-2 px-4 flex gap-2 items-center;
}
a.router-link-active {
  @apply bg-primary-500/10 font-semibold border-b-2 border-primary-500;
}
</style>