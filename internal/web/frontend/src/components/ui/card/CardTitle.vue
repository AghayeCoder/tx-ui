<script setup lang="ts">
import { cn } from "@/lib/utils";
defineProps<{ class?: string }>();
</script>

<template>
  <h3 :class="cn('text-lg font-semibold leading-none tracking-tight', $props.class)">
    <slot />
  </h3>
</template>
