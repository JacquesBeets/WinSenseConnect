<template>
  <div class="flex flex-col gap-2">
    <NuxtLink to="/" class="text-lg">
      Dashboard
    </NuxtLink>
    <!-- <NuxtLink to="/settings" class="text-lg">
      Sensors
    </NuxtLink> -->
    <NuxtLink to="/scripts" class="text-lg">
      Scripts
    </NuxtLink>
    <NuxtLink to="/config/mqtt" class="text-lg">
      MQTT Setup
    </NuxtLink>
    <NuxtLink to="/logs" class="text-lg">
      LOGS
    </NuxtLink>
  </div>
</template>