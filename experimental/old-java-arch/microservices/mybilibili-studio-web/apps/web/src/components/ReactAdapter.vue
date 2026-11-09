<template>
  <div ref="container" class="w-full h-full" />
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch, type PropType } from "vue";
import React from "react";
import { createRoot, type Root } from "react-dom/client";

const props = defineProps({
  component: {
    type: Function as PropType<React.ComponentType<any>>,
    required: true,
  },
  componentProps: {
    type: Object,
    default: () => ({}),
  },
});

const container = ref<HTMLDivElement | null>(null);
let reactRoot: Root | null = null;

const renderReactComponent = () => {
  if (!reactRoot && container.value) {
    reactRoot = createRoot(container.value);
  }
  if (reactRoot) {
    reactRoot.render(React.createElement(props.component, props.componentProps));
  }
};

onMounted(() => {
  renderReactComponent();
});

watch(
  () => props.componentProps,
  () => {
    renderReactComponent();
  },
  { deep: true }
);

watch(
  () => props.component,
  () => {
    renderReactComponent();
  }
);

onBeforeUnmount(() => {
  if (reactRoot) {
    reactRoot.unmount();
    reactRoot = null;
  }
});
</script>
