<template>
  <BaseDialog
    :show="show"
    :title="t('admin.users.createUser')"
    width="normal"
    @close="$emit('close')"
  >
    <form id="create-user-form" @submit.prevent="submit" class="space-y-5">
      <div>
        <label class="input-label">{{ t('admin.users.email') }}</label>
        <input v-model="form.email" type="email" required class="input" :placeholder="t('admin.users.enterEmail')" />
      </div>
      <div>
        <label class="input-label">{{ t('admin.users.password') }}</label>
        <div class="flex gap-2">
          <div class="relative flex-1">
            <input v-model="form.password" type="text" required class="input pr-10" :placeholder="t('admin.users.enterPassword')" />
          </div>
          <button type="button" @click="generateRandomPassword" class="btn btn-secondary px-3">
            <Icon name="refresh" size="md" />
          </button>
        </div>
      </div>
      <div>
        <label class="input-label">{{ t('admin.users.username') }}</label>
        <input v-model="form.username" type="text" class="input" :placeholder="t('admin.users.enterUsername')" />
      </div>
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <div>
          <label class="input-label">{{ t('admin.users.columns.balance') }}</label>
          <input v-model="form.balance" type="number" step="any" class="input" />
        </div>
        <div>
          <label class="input-label">{{ t('admin.users.columns.concurrency') }}</label>
          <input v-model.number="form.concurrency" type="number" class="input" />
        </div>
      </div>
      <div>
        <label class="input-label">{{ t('admin.users.form.rpmLimit') }}</label>
        <input
          v-model.number="form.rpm_limit"
          type="number"
          min="0"
          step="1"
          class="input"
          :placeholder="t('admin.users.form.rpmLimitPlaceholder')"
        />
        <p class="input-hint">{{ t('admin.users.form.rpmLimitHint') }}</p>
      </div>

      <div v-if="exclusiveGroups.length > 0" class="space-y-2">
        <label class="input-label">{{ t('admin.users.exclusiveGroups') }}</label>
        <p class="input-hint">{{ t('admin.users.exclusiveGroupsAssignHint') }}</p>
        <div class="grid gap-2 rounded-lg border border-gray-200 bg-gray-50 p-3 dark:border-dark-600 dark:bg-dark-800">
          <label
            v-for="group in exclusiveGroups"
            :key="group.id"
            class="flex cursor-pointer items-center gap-3 rounded-lg px-3 py-2 transition-colors hover:bg-white dark:hover:bg-dark-700"
          >
            <input
              type="checkbox"
              :checked="form.allowed_groups.includes(group.id)"
              @change="toggleAllowedGroup(group.id, ($event.target as HTMLInputElement).checked)"
              class="h-4 w-4 rounded border-gray-300 text-primary-500 focus:ring-primary-500 dark:border-dark-500"
            />
            <div class="min-w-0 flex-1">
              <div class="truncate text-sm font-medium text-gray-900 dark:text-white">{{ group.name }}</div>
              <div class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">
                {{ group.platform }} · {{ t('admin.users.defaultRate') }} {{ group.rate_multiplier }}x
              </div>
            </div>
          </label>
        </div>
      </div>
    </form>
    <template #footer>
      <div class="flex justify-end gap-3">
        <button @click="$emit('close')" type="button" class="btn btn-secondary">{{ t('common.cancel') }}</button>
        <button type="submit" form="create-user-form" :disabled="loading" class="btn btn-primary">
          {{ loading ? t('admin.users.creating') : t('common.create') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'; import { adminAPI } from '@/api/admin'
import { useForm } from '@/composables/useForm'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import type { AdminGroup } from '@/types'

const props = defineProps<{ show: boolean }>()
const emit = defineEmits(['close', 'success']); const { t } = useI18n()

const form = reactive({ email: '', password: '', username: '', notes: '', balance: '', concurrency: 1, rpm_limit: 0, allowed_groups: [] as number[] })
const groups = ref<AdminGroup[]>([])
const exclusiveGroups = computed(() =>
  groups.value.filter((group) => group.status === 'active' && group.subscription_type === 'standard' && group.is_exclusive)
)

const loadGroups = async () => {
  try {
    groups.value = await adminAPI.groups.getAll()
  } catch (error) {
    console.error('Failed to load exclusive groups for create user modal:', error)
  }
}

const { loading, submit } = useForm({
  form,
  submitFn: async (data) => {
    const { balance: rawBalance, ...rest } = data
    const balance = String(rawBalance).trim()
    const payload: typeof rest & { balance?: number } = { ...rest }
    if (balance !== '') {
      payload.balance = Number(balance)
    }
    await adminAPI.users.create(payload)
    emit('success'); emit('close')
  },
  successMsg: t('admin.users.userCreated')
})

watch(() => props.show, (v) => {
  if (v) {
    Object.assign(form, { email: '', password: '', username: '', notes: '', balance: '', concurrency: 1, rpm_limit: 0, allowed_groups: [] })
    void loadGroups()
  }
})

const generateRandomPassword = () => {
  const chars = 'ABCDEFGHJKLMNPQRSTUVWXYZabcdefghjkmnpqrstuvwxyz23456789!@#$%^&*'
  let p = ''; for (let i = 0; i < 16; i++) p += chars.charAt(Math.floor(Math.random() * chars.length))
  form.password = p
}

const toggleAllowedGroup = (groupId: number, checked: boolean) => {
  form.allowed_groups = checked
    ? [...form.allowed_groups, groupId]
    : form.allowed_groups.filter((id) => id !== groupId)
}
</script>
