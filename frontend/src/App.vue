<template>
  <router-view v-slot="{ Component }">
    <transition name="page">
      <component :is="Component" />
    </transition>
  </router-view>
</template>

<script setup lang="ts">
// App root — 仅负责路由视图和过渡动画
</script>

<style lang="scss">
#app {
  min-height: 100vh;
}
</style>
