<template>
  <Teleport to="body">
    <transition name="fade">
      <div v-if="url" class="lightbox" @click="$emit('close')">
        <img :src="url" alt="照片预览" />
        <span class="lightbox-close">×</span>
      </div>
    </transition>
  </Teleport>
</template>

<script setup>
defineProps({ url: { type: String, default: '' } })
defineEmits(['close'])
</script>

<style scoped>
.lightbox {
  position: fixed;
  inset: 0;
  z-index: 1100;
  background: rgba(40, 30, 18, 0.82);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 30px;
  cursor: zoom-out;
}
.lightbox img {
  max-width: 100%;
  max-height: 92vh;
  border-radius: 12px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.4);
}
.lightbox-close {
  position: absolute;
  top: 22px;
  right: 32px;
  color: #fff;
  font-size: 38px;
  line-height: 1;
}
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.18s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
