<template>
  <div class="mobile-image container">
    <image-mj-template v-if="activeMenu.mj" />
  </div>
</template>

<script setup>
import { onMounted, ref } from "vue";
import ImageMjTemplate from "@/views/mobile/pages/ImageMjTemplate.vue";
import { httpGet } from "@/utils/http";

const menus = ref([]);
const activeMenu = ref({
  mj: false,
  sd: false,
  dall: false,
});

onMounted(() => {
  httpGet("/api/menu/list").then((res) => {
    menus.value = res.data;
    activeMenu.value = {
      mj: menus.value.some((item) => item.url === "/mj"),
      sd: menus.value.some((item) => item.url === "/sd"),
      dall: menus.value.some((item) => item.url === "/dalle"),
    };
  });
});
</script>

<style lang="stylus">
.mobile-image {
  // 样式已移除，不再需要 tabs 相关样式
}
</style>
