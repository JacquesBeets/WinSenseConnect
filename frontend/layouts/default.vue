<template>
  <div class="block h-full">
    <Menu />
    <div class="border-t">
      <div class="">
        <div class="grid lg:grid-cols-5">
          <!-- <Sidebar :playlists="playlists" class="hidden lg:block" /> -->
          <div class="col-span-3 lg:col-span-4 lg:border-l">
            <div class="h-full px-4 py-6 lg:px-8">
              <slot />
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
// import Menu from './components/Menu.vue'
// import Sidebar from './components/Sidebar.vue'
</script>