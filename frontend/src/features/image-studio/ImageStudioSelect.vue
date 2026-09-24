<template>
  <div
    ref="rootRef"
    class="studio-select"
    :class="{
      'studio-select--open': open,
      'studio-select--down': direction === 'down',
    }"
  >
    <button
      type="button"
      class="studio-select__trigger"
      :disabled="disabled"
      :aria-expanded="open"
      aria-haspopup="listbox"
      @click="toggle"
      @keydown.esc.prevent="close"
    >
      <span>{{ selectedOption?.label ?? placeholder }}</span>
      <Icon name="chevronDown" size="sm" :class="{ 'rotate-180': open }" />
    </button>

    <Transition name="studio-select-menu">
      <div v-if="open" class="studio-select__menu" role="listbox">
        <button
          v-for="option in options"
          :key="option.value"
          type="button"
          role="option"
          class="studio-select__option"
          :disabled="option.disabled"
          :class="{
            'studio-select__option--selected': option.value === modelValue,
            'studio-select__option--disabled': option.disabled,
          }"
          :aria-selected="option.value === modelValue"
          @click="option.disabled || select(option.value)"
        >
          {{ option.label }}
        </button>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'

import Icon from '@/components/icons/Icon.vue'

interface SelectOption {
  value: string
  label: string
  disabled?: boolean
}

const props = withDefaults(defineProps<{
  modelValue: string
  options: SelectOption[]
  placeholder?: string
  disabled?: boolean
  direction?: 'up' | 'down'
}>(), {
  placeholder: '请选择',
  disabled: false,
  direction: 'up',
})

const emit = defineEmits<{
  (event: 'update:modelValue', value: string): void
}>()

const rootRef = ref<HTMLElement | null>(null)
const open = ref(false)
const selectedOption = computed(() => props.options.find((option) => option.value === props.modelValue) ?? null)

function toggle() {
  if (props.disabled) return
  open.value = !open.value
}

function close() {
  open.value = false
}

function select(value: string) {
  emit('update:modelValue', value)
  close()
}

function handleClickOutside(event: MouseEvent) {
  if (!rootRef.value?.contains(event.target as Node)) {
    close()
  }
}

onMounted(() => document.addEventListener('click', handleClickOutside))
onBeforeUnmount(() => document.removeEventListener('click', handleClickOutside))
</script>

<style scoped>
.studio-select {
  position: relative;
  min-width: 5.8rem;
}

.studio-select__trigger {
  display: flex;
  width: 100%;
  height: 2.35rem;
  align-items: center;
  justify-content: space-between;
  gap: 0.55rem;
  border: 1px solid rgb(226 232 240);
  border-radius: 0.75rem;
  background: rgb(255 255 255 / 0.92);
  padding: 0 0.7rem;
  color: rgb(30 41 59);
  font-size: 0.82rem;
  box-shadow: 0 1px 2px rgb(92 67 55 / 0.04);
  transition:
    border-color 0.18s ease,
    box-shadow 0.18s ease,
    transform 0.18s ease;
}

.studio-select__trigger:hover,
.studio-select--open .studio-select__trigger {
  border-color: rgb(45 212 191);
  box-shadow: 0 6px 16px rgb(20 184 166 / 0.14);
}

.studio-select__trigger:disabled {
  cursor: not-allowed;
  opacity: 0.55;
  box-shadow: none;
}

.studio-select__menu {
  position: absolute;
  right: 0;
  bottom: calc(100% + 0.45rem);
  z-index: 40;
  min-width: 100%;
  overflow: hidden;
  border: 1px solid rgb(226 232 240);
  border-radius: 0.85rem;
  background: rgb(255 255 255 / 0.98);
  box-shadow: 0 16px 34px rgb(15 23 42 / 0.14);
  backdrop-filter: blur(16px);
}

.studio-select--down .studio-select__menu {
  top: calc(100% + 0.45rem);
  bottom: auto;
  transform-origin: top;
}

.studio-select__option {
  display: block;
  width: 100%;
  border: 0;
  border-bottom: 1px solid rgb(241 245 249);
  background: transparent;
  padding: 0.68rem 0.9rem;
  color: rgb(51 65 85);
  font-size: 0.82rem;
  text-align: left;
  white-space: nowrap;
  transition:
    background 0.16s ease,
    color 0.16s ease;
}

.studio-select__option:last-child {
  border-bottom: 0;
}

.studio-select__option:hover {
  background: rgb(240 253 250);
}

.studio-select__option--selected {
  background: rgb(204 251 241);
  color: rgb(15 118 110);
}

.studio-select__option--disabled {
  cursor: not-allowed;
  color: rgb(148 163 184);
}

.studio-select-menu-enter-active,
.studio-select-menu-leave-active {
  transition:
    opacity 0.16s ease,
    transform 0.16s ease;
  transform-origin: bottom;
}

.studio-select-menu-enter-from,
.studio-select-menu-leave-to {
  opacity: 0;
  transform: translateY(6px) scale(0.98);
}

.studio-select--down .studio-select-menu-enter-from,
.studio-select--down .studio-select-menu-leave-to {
  transform: translateY(-6px) scale(0.98);
}

:global(.dark .studio-select__trigger) {
  border-color: rgb(51 65 85);
  background: rgb(30 41 59);
  color: rgb(241 245 249);
}

:global(.dark .studio-select__menu) {
  border-color: rgb(51 65 85);
  background: rgb(30 41 59 / 0.98);
}

:global(.dark .studio-select__option) {
  border-bottom-color: rgb(51 65 85);
  color: rgb(226 232 240);
}

:global(.dark .studio-select__option--selected) {
  background: rgb(19 78 74 / 0.75);
  color: rgb(94 234 212);
}
</style>
