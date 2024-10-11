<template>
  <div class="block h-full">
    <Menu />
    <div class="border-t border-slate-400 h-full">

        <div class="grid md:grid-cols-4 lg:grid-cols-5 h-full">
          <Sidebar class="block col-span-1 p-4" />
          <div class="border-slate-400 md:col-span-3 lg:border-l" >
            <div class="h-full w-full px-4 py-6 lg:px-8">
              <slot />
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