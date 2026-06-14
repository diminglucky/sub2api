<template>
  <AppLayout>
    <div class="w-full space-y-6">
      <div v-if="loading" class="flex items-center justify-center py-20">
        <div class="h-8 w-8 animate-spin rounded-full border-4 border-primary-500 border-t-transparent"></div>
      </div>
      <template v-else>
        <!-- Tab Switcher (hide during payment and subscription confirm) -->
        <div v-if="tabs.length > 1 && paymentPhase === 'select' && !selectedPlan" class="flex space-x-1 rounded-xl bg-gray-100 p-1 dark:bg-dark-800">
          <button v-for="tab in tabs" :key="tab.key"
            class="flex-1 rounded-lg px-4 py-2.5 text-sm font-medium transition-all"
            :class="activeTab === tab.key ? 'bg-white text-gray-900 shadow dark:bg-dark-700 dark:text-white' : 'text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300'"
            @click="activeTab = tab.key">{{ tab.label }}</button>
        </div>
        <!-- Payment in progress (shared by recharge and subscription) -->
        <template v-if="paymentPhase === 'paying'">
          <PaymentStatusPanel
            :order-id="paymentState.orderId"
            :qr-code="paymentState.qrCode"
            :expires-at="paymentState.expiresAt"
            :payment-type="paymentState.paymentType"
            :pay-url="paymentState.payUrl"
            :order-type="paymentState.orderType"
            :currency="paymentState.currency || selectedCurrency"
            @done="onPaymentDone"
            @success="onPaymentSuccess"
            @settled="onPaymentSettled"
          />
        </template>
        <!-- Tab content (select phase) -->
        <template v-else>
          <!-- Top-up Tab -->
          <template v-if="activeTab === 'recharge'">
            <section class="card mx-auto max-w-5xl space-y-6 p-4 sm:p-5">
              <header class="flex items-center gap-3">
                <Icon name="creditCard" size="lg" class="text-orange-500" />
                <h2 class="text-xl font-bold text-gray-900 dark:text-white">{{ t('payment.onlineRecharge') }}</h2>
              </header>

              <div v-if="hasRechargePackages" class="space-y-3">
                <div class="flex flex-wrap items-center justify-between gap-3">
                  <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('payment.rechargePackages') }}</h3>
                </div>
                <div class="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-3">
                  <button
                    v-for="pkg in availableRechargePackages"
                    :key="pkg.id"
                    type="button"
                    :class="[
                      'group flex min-h-36 flex-col justify-between rounded-xl border p-4 text-left transition-all',
                      selectedRechargePackageId === pkg.id
                        ? 'border-orange-500 bg-orange-500/5 shadow-sm ring-1 ring-orange-500/30 dark:bg-orange-500/10'
                        : 'border-gray-200 bg-white hover:border-orange-300 hover:bg-orange-50/30 dark:border-dark-700 dark:bg-dark-800 dark:hover:border-orange-500/60 dark:hover:bg-orange-950/10',
                    ]"
                    @click="selectRechargePackage(pkg)"
                  >
                    <span class="flex items-start gap-3">
                      <span class="mt-0.5 flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-orange-50 text-orange-500 dark:bg-orange-950/40">
                        <Icon name="creditCard" size="md" />
                      </span>
                      <span class="min-w-0">
                        <span class="block truncate text-base font-bold text-gray-900 dark:text-white">{{ pkg.name }}</span>
                        <span class="mt-1 block text-sm font-semibold text-gray-500 dark:text-gray-400">
                          {{ t('payment.rechargePackageCredit', { amount: pkg.amount.toFixed(2) }) }}
                        </span>
                      </span>
                    </span>
                    <span class="mt-5 flex items-end justify-between gap-3">
                      <span class="text-2xl font-black text-gray-950 dark:text-white">
                        {{ formatPackagePaymentAmount(pkg.pay_amount) }}
                      </span>
                      <span class="text-sm font-bold text-orange-500">
                        {{ selectedRechargePackageId === pkg.id ? t('payment.selectedRechargePackage') : t('payment.selectRechargePackage') }}
                      </span>
                    </span>
                  </button>
                </div>
              </div>

              <div v-else class="space-y-3">
                <label class="block text-sm font-semibold text-gray-900 dark:text-white">
                  {{ t('payment.amountLabel') }}
                </label>
                <div class="relative">
                  <input
                    type="number"
                    inputmode="decimal"
                    :value="amountInputText"
                    :min="rechargeMinAmount"
                    :max="rechargeMaxAmount || undefined"
                    step="0.01"
                    :placeholder="amountPlaceholder"
                    class="input h-11 w-full rounded-lg border-gray-300 bg-transparent pr-4 text-base font-semibold dark:border-dark-600"
                    @input="handleRechargeAmountInput"
                  />
                </div>
                <p v-if="amountError" class="text-xs text-amber-600 dark:text-amber-300">{{ amountError }}</p>
                <p v-else class="flex items-center gap-2 text-sm font-semibold text-gray-500 dark:text-gray-400">
                  <Icon name="document" size="sm" />
                  <span>{{ rechargeEstimateText }}</span>
                </p>
              </div>

              <div v-if="!hasRechargePackages" class="space-y-3">
                <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('payment.quickAmounts') }}</h3>
                <div class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5">
                  <button
                    v-for="option in quickRechargeOptions"
                    :key="option.amount"
                    type="button"
                    :class="[
                      'flex h-16 flex-col items-center justify-center rounded-lg border text-center transition-colors',
                      amount === option.amount
                        ? 'border-primary-500 bg-primary-500/10 text-primary-600 dark:border-primary-400 dark:text-primary-300'
                        : 'border-gray-300 bg-transparent text-gray-900 hover:border-primary-400 dark:border-dark-600 dark:text-white dark:hover:border-primary-500',
                    ]"
                    @click="selectRechargeAmount(option.amount)"
                  >
                    <span class="text-base font-bold">{{ formatQuickAmount(option.amount) }}</span>
                    <span
                      v-if="option.badge"
                      class="mt-2 rounded-full bg-emerald-100 px-2 py-0.5 text-xs font-semibold text-emerald-700 dark:bg-emerald-950/60 dark:text-emerald-300"
                    >
                      {{ option.badge }}
                    </span>
                  </button>
                </div>
              </div>

              <div v-if="enabledMethods.length >= 1" class="space-y-3">
                <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('payment.paymentMethod') }}</h3>
                <div class="flex flex-wrap gap-3">
                  <button
                    v-for="method in methodOptions"
                    :key="method.type"
                    type="button"
                    :disabled="!method.available"
                    :class="[
                      'flex h-14 items-center gap-3 rounded-lg border px-4 transition-colors',
                      !method.available
                        ? 'cursor-not-allowed border-gray-200 opacity-50 dark:border-dark-700'
                        : selectedMethod === method.type
                          ? 'border-orange-500 bg-orange-500/10 text-gray-900 dark:text-white'
                          : 'border-gray-300 text-gray-700 hover:border-gray-400 dark:border-dark-600 dark:text-gray-200 dark:hover:border-dark-500',
                    ]"
                    @click="method.available && (selectedMethod = method.type)"
                  >
                    <span
                      class="flex h-4 w-4 items-center justify-center rounded-full border-2"
                      :class="selectedMethod === method.type ? 'border-orange-500' : 'border-gray-400 dark:border-dark-500'"
                    >
                      <span v-if="selectedMethod === method.type" class="h-2 w-2 rounded-full bg-orange-500"></span>
                    </span>
                    <img :src="paymentMethodIcon(method.type)" :alt="t(`payment.methods.${method.type}`)" class="h-8 w-8 object-contain" />
                    <span class="text-base font-bold">{{ t(`payment.methods.${method.type}`) }}</span>
                  </button>
                </div>
              </div>
              <div v-else class="space-y-3">
                <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('payment.paymentMethod') }}</h3>
                <div class="rounded-lg border border-dashed border-gray-300 px-4 py-4 text-sm font-semibold text-gray-500 dark:border-dark-600 dark:text-gray-400">
                  {{ t('payment.noPaymentMethodsConfigured') }}
                </div>
              </div>

              <div v-if="validAmount > 0 && (feeRate > 0 || balanceRechargeMultiplier !== 1)" class="max-w-xl space-y-2 rounded-lg border border-gray-200 p-4 text-sm dark:border-dark-700">
                <div class="flex justify-between">
                  <span class="text-gray-500 dark:text-gray-400">{{ t('payment.paymentAmount') }}</span>
                  <span class="text-gray-900 dark:text-white">{{ formatSelectedPaymentAmount(validAmount) }}</span>
                </div>
                <div v-if="feeRate > 0" class="flex justify-between">
                  <span class="text-gray-500 dark:text-gray-400">{{ t('payment.fee') }} ({{ feeRate }}%)</span>
                  <span class="text-gray-900 dark:text-white">{{ formatSelectedPaymentAmount(feeAmount) }}</span>
                </div>
                <div v-if="feeRate > 0" class="flex justify-between border-t border-gray-200 pt-2 dark:border-dark-600">
                  <span class="font-medium text-gray-700 dark:text-gray-300">{{ t('payment.actualPay') }}</span>
                  <span class="text-lg font-bold text-primary-600 dark:text-primary-400">{{ formatSelectedPaymentAmount(totalAmount) }}</span>
                </div>
                <div v-if="balanceRechargeMultiplier !== 1" class="flex justify-between" :class="{ 'border-t border-gray-200 pt-2 dark:border-dark-600': feeRate <= 0 }">
                  <span class="text-gray-500 dark:text-gray-400">{{ t('payment.creditedBalance') }}</span>
                  <span class="text-gray-900 dark:text-white">¥{{ creditedAmount.toFixed(2) }}</span>
                </div>
              </div>

              <button
                :class="['btn min-w-28 px-6 py-3 text-base font-bold', paymentButtonClass]"
                :disabled="!canSubmit || submitting"
                @click="handleSubmitRecharge"
              >
                <span v-if="submitting" class="flex items-center justify-center gap-2">
                  <span class="h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent"></span>
                  {{ t('common.processing') }}
                </span>
                <span v-else>{{ t('payment.payNow') }}</span>
              </button>
            </section>
          </template>
          <!-- Recharge Card Tab -->
          <template v-else-if="activeTab === 'rechargeCard'">
            <section class="card mx-auto max-w-5xl space-y-6 p-4 sm:p-5">
              <header class="flex items-center justify-between gap-3">
                <div class="flex items-center gap-3">
                  <Icon name="gift" size="lg" class="text-primary-500" />
                  <h2 class="text-xl font-bold text-gray-900 dark:text-white">{{ t('payment.buyRechargeCard') }}</h2>
                </div>
                <button type="button" class="btn btn-secondary" @click="activeTab = 'redeem'">
                  {{ t('payment.goRedeem') }}
                </button>
              </header>

              <div v-if="rechargeCardProducts.length === 0" class="rounded-lg border border-dashed border-gray-300 px-4 py-8 text-center text-sm font-semibold text-gray-500 dark:border-dark-600 dark:text-gray-400">
                {{ t('payment.noRechargeCardProducts') }}
              </div>
              <div v-else class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
                <button
                  v-for="product in rechargeCardProducts"
                  :key="`${product.name}-${product.url}`"
                  type="button"
                  class="flex min-h-24 items-center justify-between gap-3 rounded-lg border border-gray-300 px-4 py-3 text-left transition-colors hover:border-primary-400 hover:bg-primary-50/50 dark:border-dark-600 dark:hover:border-primary-500 dark:hover:bg-primary-950/20"
                  @click="openRechargeCardDialog(product)"
                >
                  <span class="min-w-0">
                    <span class="block truncate text-lg font-bold text-gray-900 dark:text-white">{{ product.name }}</span>
                    <span class="mt-1 block text-sm font-medium text-gray-500 dark:text-gray-400">
                      {{ rechargeCardProductMeta(product) }}
                    </span>
                  </span>
                  <Icon name="externalLink" size="sm" class="shrink-0 text-primary-500" />
                </button>
              </div>
              <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('payment.rechargeCardRedeemHint') }}</p>
            </section>
          </template>
          <!-- Subscribe Tab -->
          <template v-else-if="activeTab === 'subscription'">
            <!-- Subscription confirm (inline, replaces plan list) -->
            <template v-if="selectedPlan">
              <div class="card p-5">
                <!-- Header: platform badge + plan name -->
                <div class="mb-3 flex flex-wrap items-center gap-2">
                  <span :class="['rounded-md border px-2 py-0.5 text-xs font-medium', planBadgeClass]">
                    {{ platformLabel(selectedPlan.group_platform || '') }}
                  </span>
                  <h3 class="text-lg font-bold text-gray-900 dark:text-white">{{ selectedPlan.name }}</h3>
                </div>
                <!-- Price -->
                <div class="flex items-baseline gap-2">
                  <span v-if="selectedPlan.original_price" class="text-sm text-gray-400 line-through dark:text-gray-500">
                    {{ formatSelectedPaymentAmount(selectedPlan.original_price) }}
                  </span>
                  <span :class="['text-3xl font-bold', planTextClass]">{{ formatSelectedPaymentAmount(selectedPlan.price) }}</span>
                  <span class="text-sm text-gray-500 dark:text-gray-400">/ {{ planValiditySuffix }}</span>
                </div>
                <!-- Description -->
                <p v-if="selectedPlan.description" class="mt-2 text-sm leading-relaxed text-gray-500 dark:text-gray-400">
                  {{ selectedPlan.description }}
                </p>
                <!-- Rate + Limits grid -->
                <div class="mt-3 grid grid-cols-2 gap-3">
                  <div>
                    <span class="text-xs text-gray-400 dark:text-gray-500">{{ t('payment.planCard.rate') }}</span>
                    <div class="flex items-baseline">
                      <span :class="['text-lg font-bold', planTextClass]">×{{ selectedPlan.rate_multiplier ?? 1 }}</span>
                    </div>
                  </div>
                  <div v-if="selectedPlan.daily_limit_usd != null">
                    <span class="text-xs text-gray-400 dark:text-gray-500">{{ t('payment.planCard.dailyLimit') }}</span>
                    <div class="text-lg font-semibold text-gray-800 dark:text-gray-200">¥{{ selectedPlan.daily_limit_usd }}</div>
                  </div>
                  <div v-if="selectedPlan.weekly_limit_usd != null">
                    <span class="text-xs text-gray-400 dark:text-gray-500">{{ t('payment.planCard.weeklyLimit') }}</span>
                    <div class="text-lg font-semibold text-gray-800 dark:text-gray-200">¥{{ selectedPlan.weekly_limit_usd }}</div>
                  </div>
                  <div v-if="selectedPlan.monthly_limit_usd != null">
                    <span class="text-xs text-gray-400 dark:text-gray-500">{{ t('payment.planCard.monthlyLimit') }}</span>
                    <div class="text-lg font-semibold text-gray-800 dark:text-gray-200">¥{{ selectedPlan.monthly_limit_usd }}</div>
                  </div>
                  <div v-if="selectedPlan.daily_limit_usd == null && selectedPlan.weekly_limit_usd == null && selectedPlan.monthly_limit_usd == null">
                    <span class="text-xs text-gray-400 dark:text-gray-500">{{ t('payment.planCard.quota') }}</span>
                    <div class="text-lg font-semibold text-gray-800 dark:text-gray-200">{{ t('payment.planCard.unlimited') }}</div>
                  </div>
                </div>
              </div>
              <div v-if="enabledMethods.length >= 1" class="card p-6">
                <PaymentMethodSelector
                  :methods="subMethodOptions"
                  :selected="selectedMethod"
                  @select="selectedMethod = $event"
                />
              </div>
              <div v-if="feeRate > 0 && selectedPlan.price > 0" class="card p-6">
                <div class="space-y-2 text-sm">
                  <div class="flex justify-between">
                    <span class="text-gray-500 dark:text-gray-400">{{ t('payment.amountLabel') }}</span>
                    <span class="text-gray-900 dark:text-white">{{ formatSelectedPaymentAmount(selectedPlan.price) }}</span>
                  </div>
                  <div class="flex justify-between">
                    <span class="text-gray-500 dark:text-gray-400">{{ t('payment.fee') }} ({{ feeRate }}%)</span>
                    <span class="text-gray-900 dark:text-white">{{ formatSelectedPaymentAmount(subFeeAmount) }}</span>
                  </div>
                  <div class="flex justify-between border-t border-gray-200 pt-2 dark:border-dark-600">
                    <span class="font-medium text-gray-700 dark:text-gray-300">{{ t('payment.actualPay') }}</span>
                    <span class="text-lg font-bold text-primary-600 dark:text-primary-400">{{ formatSelectedPaymentAmount(subTotalAmount) }}</span>
                  </div>
                </div>
              </div>
              <button :class="['btn w-full py-3 text-base font-medium', paymentButtonClass]" :disabled="!canSubmitSubscription || submitting" @click="confirmSubscribe">
                <span v-if="submitting" class="flex items-center justify-center gap-2">
                  <span class="h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent"></span>
                  {{ t('common.processing') }}
                </span>
                <span v-else>{{ t('payment.createOrder') }} {{ formatSelectedPaymentAmount(feeRate > 0 ? subTotalAmount : selectedPlan.price) }}</span>
              </button>
              <button class="btn btn-secondary w-full" @click="selectedPlan = null">{{ t('common.cancel') }}</button>
            </template>
            <!-- Plan list -->
            <template v-else>
              <div v-if="checkout.plans.length === 0" class="card py-16 text-center">
                <Icon name="gift" size="xl" class="mx-auto mb-3 text-gray-300 dark:text-dark-600" />
                <p class="text-gray-500 dark:text-gray-400">{{ t('payment.noPlans') }}</p>
              </div>
              <div v-else :class="planGridClass">
                <SubscriptionPlanCard v-for="plan in checkout.plans" :key="plan.id" :plan="plan" :active-subscriptions="activeSubscriptions" @select="selectPlan" />
              </div>
              <!-- Active subscriptions (compact, below plan list) -->
              <div v-if="activeSubscriptions.length > 0">
                <p class="mb-2 text-xs font-medium text-gray-400 dark:text-gray-500">{{ t('payment.activeSubscription') }}</p>
                <div class="space-y-2">
                  <div v-for="sub in activeSubscriptions" :key="sub.id"
                    class="flex items-center gap-3 rounded-xl border border-gray-100 bg-white px-3 py-2 dark:border-dark-700 dark:bg-dark-800">
                    <div :class="['h-6 w-1 shrink-0 rounded-full', platformAccentBarClass(sub.group?.platform || '')]" />
                    <div class="min-w-0 flex-1">
                      <div class="flex items-center gap-1.5">
                        <span class="truncate text-xs font-semibold text-gray-900 dark:text-white">{{ sub.group?.name || t('payment.groupFallback', { id: sub.group_id }) }}</span>
                        <span :class="['shrink-0 rounded-full px-1.5 py-0.5 text-[9px] font-medium', platformBadgeLightClass(sub.group?.platform || '')]">{{ platformLabel(sub.group?.platform || '') }}</span>
                      </div>
                      <div class="flex flex-wrap gap-x-3 text-[11px] text-gray-400 dark:text-gray-500">
                        <span>{{ t('payment.planCard.rate') }}: ×{{ sub.group?.rate_multiplier ?? 1 }}</span>
                        <span v-if="sub.group?.daily_limit_usd == null && sub.group?.weekly_limit_usd == null && sub.group?.monthly_limit_usd == null">{{ t('payment.planCard.quota') }}: {{ t('payment.planCard.unlimited') }}</span>
                        <span v-if="sub.expires_at">{{ t('userSubscriptions.daysRemaining', { days: getDaysRemaining(sub.expires_at) }) }}</span>
                        <span v-else>{{ t('userSubscriptions.noExpiration') }}</span>
                      </div>
                    </div>
                    <span class="badge badge-success shrink-0 text-[10px]">{{ t('userSubscriptions.status.active') }}</span>
                  </div>
                </div>
              </div>
            </template>
          </template>
          <!-- Redeem Tab -->
          <template v-else-if="activeTab === 'redeem'">
            <RedeemView embedded />
          </template>
        </template>
        <div v-if="activeTab !== 'redeem' && (checkout.help_text || checkout.help_image_url) && paymentPhase === 'select' && !selectedPlan" class="card p-4">
          <div class="flex flex-col items-center gap-3">
            <img v-if="checkout.help_image_url" :src="checkout.help_image_url" alt=""
              class="h-40 max-w-full cursor-pointer rounded-lg object-contain transition-opacity hover:opacity-80"
              @click="previewImage = checkout.help_image_url" />
            <p v-if="checkout.help_text" class="text-center text-sm text-gray-500 dark:text-gray-400">{{ checkout.help_text }}</p>
          </div>
        </div>
      </template>
    </div>
    <!-- Renewal Plan Selection Modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showRenewalModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4" @click.self="closeRenewalModal">
          <div class="relative w-full max-w-lg rounded-2xl border border-gray-200 bg-white p-6 shadow-2xl dark:border-dark-700 dark:bg-dark-900">
            <!-- Close button -->
            <button class="absolute right-4 top-4 rounded-lg p-1 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-dark-700 dark:hover:text-gray-200" @click="closeRenewalModal">
              <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" /></svg>
            </button>
            <h3 class="mb-4 text-lg font-semibold text-gray-900 dark:text-white">{{ t('payment.selectPlan') }}</h3>
            <div class="space-y-4">
              <SubscriptionPlanCard v-for="plan in renewalPlans" :key="plan.id" :plan="plan" :active-subscriptions="activeSubscriptions" @select="selectPlanFromModal" />
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
    <!-- Recharge Card Purchase Modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div
          v-if="selectedRechargeCardProduct"
          class="fixed inset-0 z-[55] flex items-center justify-center bg-black/70 p-4 backdrop-blur-md"
          @click.self="closeRechargeCardDialog"
        >
          <div class="relative w-full max-w-3xl overflow-hidden rounded-2xl border border-gray-200 bg-white shadow-2xl dark:border-dark-700 dark:bg-dark-900">
            <button
              type="button"
              class="absolute right-4 top-4 z-10 rounded-lg p-1 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-dark-700 dark:hover:text-gray-200"
              :aria-label="t('common.close')"
              @click="closeRechargeCardDialog"
            >
              <Icon name="x" size="md" />
            </button>
            <div class="grid gap-0 md:grid-cols-[1fr_280px]">
              <section class="space-y-6 p-6 pr-14 md:p-8 md:pr-8">
                <div class="flex items-center gap-4">
                  <span class="flex h-14 w-14 shrink-0 items-center justify-center rounded-2xl bg-primary-500/10 text-primary-500">
                    <Icon name="gift" size="xl" />
                  </span>
                  <div class="min-w-0">
                    <h3 class="truncate text-2xl font-bold text-gray-900 dark:text-white">
                      {{ selectedRechargeCardProduct.name }}
                    </h3>
                    <p class="mt-1 text-sm font-medium text-gray-500 dark:text-gray-400">
                      {{ t('payment.rechargeCardDialogSubtitle') }}
                    </p>
                  </div>
                </div>

                <div class="grid gap-3 sm:grid-cols-2">
                  <div class="rounded-xl border border-gray-200 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-800/60">
                    <div class="text-sm text-gray-500 dark:text-gray-400">{{ t('payment.rechargeCardAmountLabel') }}</div>
                    <div class="mt-1 text-3xl font-bold text-gray-900 dark:text-white">
                      ¥{{ selectedRechargeCardProduct.amount || '-' }}
                    </div>
                  </div>
                  <div class="rounded-xl border border-primary-200 bg-primary-50 p-4 dark:border-primary-900/40 dark:bg-primary-950/30">
                    <div class="text-sm text-primary-600 dark:text-primary-300">{{ t('payment.rechargeCardPriceLabel') }}</div>
                    <div class="mt-1 text-3xl font-bold text-primary-600 dark:text-primary-300">
                      ¥{{ selectedRechargeCardProduct.price || '-' }}
                    </div>
                  </div>
                </div>

                <div class="rounded-xl border border-gray-200 p-4 dark:border-dark-700">
                  <div class="mb-2 flex items-center gap-2 text-sm font-semibold text-gray-900 dark:text-white">
                    <Icon name="document" size="sm" class="text-primary-500" />
                    {{ t('payment.rechargeCardHowToUse') }}
                  </div>
                  <p class="text-sm leading-6 text-gray-500 dark:text-gray-400">
                    {{ t('payment.rechargeCardRedeemHint') }}
                  </p>
                </div>

                <div class="flex flex-wrap gap-2">
                  <button type="button" class="btn btn-primary" @click="activeTab = 'redeem'; closeRechargeCardDialog()">
                    {{ t('payment.goRedeem') }}
                  </button>
                  <a
                    :href="selectedRechargeCardProduct.url"
                    target="_blank"
                    rel="noopener noreferrer"
                    class="btn btn-secondary"
                  >
                    {{ t('payment.openRechargeCardPurchase') }}
                  </a>
                </div>
              </section>

              <aside class="flex flex-col items-center justify-center gap-4 border-t border-gray-200 bg-gray-50 p-6 dark:border-dark-700 dark:bg-dark-950/50 md:border-l md:border-t-0">
                <div class="rounded-2xl border border-gray-200 bg-white p-4 shadow-sm dark:border-dark-700 dark:bg-white">
                  <canvas ref="rechargeCardQrCanvas" class="block h-52 w-52"></canvas>
                </div>
                <div class="space-y-1 text-center">
                  <div class="text-base font-bold text-gray-900 dark:text-white">{{ t('payment.scanRechargeCardQr') }}</div>
                  <p class="text-sm leading-5 text-gray-500 dark:text-gray-400">
                    {{ t('payment.scanRechargeCardQrHint') }}
                  </p>
                </div>
              </aside>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
    <!-- Image Preview Overlay -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="previewImage" class="fixed inset-0 z-[60] flex items-center justify-center bg-black/70 backdrop-blur-sm" @click="previewImage = ''">
          <img :src="previewImage" alt="" class="max-h-[85vh] max-w-[90vw] rounded-xl object-contain shadow-2xl" />
        </div>
      </Transition>
    </Teleport>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { usePaymentStore } from '@/stores/payment'
import { useSubscriptionStore } from '@/stores/subscriptions'
import { useAppStore } from '@/stores'
import { paymentAPI } from '@/api/payment'
import { extractApiErrorMessage, extractI18nErrorMessage } from '@/utils/apiError'
import { isMobileDevice } from '@/utils/device'
import type { SubscriptionPlan, CheckoutInfoResponse, CreateOrderResult, OrderType, RechargeCardProduct, RechargePackage } from '@/types/payment'
import AppLayout from '@/components/layout/AppLayout.vue'
import PaymentMethodSelector from '@/components/payment/PaymentMethodSelector.vue'
import { METHOD_ORDER, getPaymentPopupFeatures } from '@/components/payment/providerConfig'
import {
  PAYMENT_RECOVERY_STORAGE_KEY,
  buildCreateOrderPayload,
  clearPaymentRecoverySnapshot,
  decidePaymentLaunch,
  getVisibleMethods,
  normalizeVisibleMethod,
  readPaymentRecoverySnapshot,
  type PaymentRecoverySnapshot,
  writePaymentRecoverySnapshot,
} from '@/components/payment/paymentFlow'
import { platformAccentBarClass, platformBadgeLightClass, platformBadgeClass, platformTextClass, platformLabel } from '@/utils/platformColors'
import SubscriptionPlanCard from '@/components/payment/SubscriptionPlanCard.vue'
import PaymentStatusPanel from '@/components/payment/PaymentStatusPanel.vue'
import Icon from '@/components/icons/Icon.vue'
import RedeemView from './RedeemView.vue'
import { formatPaymentAmount, normalizePaymentCurrency } from '@/components/payment/currency'
import type { PaymentMethodOption } from '@/components/payment/PaymentMethodSelector.vue'
import { buildPaymentErrorToastMessage, describePaymentScenarioError } from './paymentUx'
import { hasWechatResumeQuery, parseWechatResumeRoute, stripWechatResumeQuery } from './paymentWechatResume'
import QRCode from 'qrcode'
import alipayIcon from '@/assets/icons/alipay.svg'
import wxpayIcon from '@/assets/icons/wxpay.svg'
import stripeIcon from '@/assets/icons/stripe.svg'
import airwallexIcon from '@/assets/icons/airwallex.svg'
import easypayIcon from '@/assets/icons/easypay.svg'

const i18n = useI18n()
const { t } = i18n
const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const paymentStore = usePaymentStore()
const subscriptionStore = useSubscriptionStore()
const appStore = useAppStore()

const activeSubscriptions = computed(() => subscriptionStore.activeSubscriptions)

function getDaysRemaining(expiresAt: string): number {
  const diff = new Date(expiresAt).getTime() - Date.now()
  return Math.max(0, Math.ceil(diff / (1000 * 60 * 60 * 24)))
}

const loading = ref(true)
const submitting = ref(false)
const errorMessage = ref('')
const errorHintMessage = ref('')
const activeTab = ref<'recharge' | 'rechargeCard' | 'subscription' | 'redeem'>('recharge')
const amount = ref<number | null>(null)
const selectedRechargePackageId = ref('')
const selectedMethod = ref('')
const selectedPlan = ref<SubscriptionPlan | null>(null)
const selectedRechargeCardProduct = ref<RechargeCardProduct | null>(null)
const rechargeCardQrCanvas = ref<HTMLCanvasElement | null>(null)
const previewImage = ref('')

const paymentPhase = ref<'select' | 'paying'>('select')

interface CreateOrderOptions {
  openid?: string
  wechatResumeToken?: string
  paymentType?: string
  rechargePackageId?: string
  isResume?: boolean
  mobileQrFallbackAttempted?: boolean
}

interface WeixinJSBridgeLike {
  invoke(
    action: string,
    payload: Record<string, unknown>,
    callback: (result: Record<string, unknown>) => void,
  ): void
}

function emptyPaymentState(): PaymentRecoverySnapshot {
  return {
    orderId: 0,
    amount: 0,
    qrCode: '',
    expiresAt: '',
    paymentType: '',
    payUrl: '',
    outTradeNo: '',
    clientSecret: '',
    intentId: '',
    currency: '',
    countryCode: '',
    paymentEnv: '',
    payAmount: 0,
    orderType: '',
    paymentMode: '',
    resumeToken: '',
    createdAt: 0,
  }
}

function getWeixinJSBridge(): WeixinJSBridgeLike | undefined {
  return (window as Window & { WeixinJSBridge?: WeixinJSBridgeLike }).WeixinJSBridge
}

function waitForWeixinJSBridge(timeoutMs = 4000): Promise<WeixinJSBridgeLike | null> {
  const existing = getWeixinJSBridge()
  if (existing) return Promise.resolve(existing)

  return new Promise((resolve) => {
    let settled = false
    const finish = (bridge: WeixinJSBridgeLike | null) => {
      if (settled) return
      settled = true
      document.removeEventListener('WeixinJSBridgeReady', handleReady)
      document.removeEventListener('onWeixinJSBridgeReady', handleReady)
      window.clearTimeout(timer)
      resolve(bridge)
    }
    const handleReady = () => finish(getWeixinJSBridge() ?? null)
    const timer = window.setTimeout(() => finish(getWeixinJSBridge() ?? null), timeoutMs)
    document.addEventListener('WeixinJSBridgeReady', handleReady, false)
    document.addEventListener('onWeixinJSBridgeReady', handleReady, false)
  })
}

async function invokeWechatJsapiPayment(payload: Record<string, unknown>): Promise<Record<string, unknown>> {
  const bridge = await waitForWeixinJSBridge()
  if (!bridge) {
    throw new Error('WECHAT_JSAPI_UNAVAILABLE')
  }
  return new Promise((resolve) => {
    bridge.invoke('getBrandWCPayRequest', payload, (result) => resolve(result || {}))
  })
}

const paymentState = ref<PaymentRecoverySnapshot>(emptyPaymentState())

function persistRecoverySnapshot(snapshot: PaymentRecoverySnapshot) {
  if (typeof window === 'undefined' || !snapshot.orderId) return
  writePaymentRecoverySnapshot(window.localStorage, snapshot, PAYMENT_RECOVERY_STORAGE_KEY)
}

function removeRecoverySnapshot() {
  if (typeof window === 'undefined') return
  clearPaymentRecoverySnapshot(window.localStorage, PAYMENT_RECOVERY_STORAGE_KEY)
}

function resetPayment() {
  paymentPhase.value = 'select'
  paymentState.value = emptyPaymentState()
  removeRecoverySnapshot()
}

async function redirectToPaymentResult(state: PaymentRecoverySnapshot): Promise<void> {
  const query: Record<string, string | undefined> = {}
  if (state.orderId > 0) {
    query.order_id = String(state.orderId)
  }
  if (state.outTradeNo) {
    query.out_trade_no = state.outTradeNo
  }
  if (state.resumeToken) {
    query.resume_token = state.resumeToken
  }
  await router.push({
    path: '/payment/result',
    query,
  })
}

function buildWechatOAuthAuthorizeUrl(
  authorizeUrl: string,
  context: { paymentType: string; orderType: OrderType; planId?: number; orderAmount: number; rechargePackageId?: string },
): string {
  const normalizedUrl = authorizeUrl.trim()
  if (!normalizedUrl || typeof window === 'undefined') {
    return normalizedUrl
  }

  try {
    const targetUrl = new URL(normalizedUrl, window.location.origin)
    const redirectPath = targetUrl.searchParams.get('redirect') || route.fullPath || '/recharge'
    const redirectUrl = new URL(redirectPath, window.location.origin)
    const paymentType = normalizeVisibleMethod(context.paymentType) || context.paymentType.trim() || 'wxpay'

    redirectUrl.searchParams.set('payment_type', paymentType)
    redirectUrl.searchParams.set('order_type', context.orderType)

    if (context.planId) {
      redirectUrl.searchParams.set('plan_id', String(context.planId))
    } else {
      redirectUrl.searchParams.delete('plan_id')
    }

    if (context.orderAmount > 0) {
      redirectUrl.searchParams.set('amount', String(context.orderAmount))
    } else {
      redirectUrl.searchParams.delete('amount')
    }
    if (context.rechargePackageId) {
      redirectUrl.searchParams.set('recharge_package_id', context.rechargePackageId)
    } else {
      redirectUrl.searchParams.delete('recharge_package_id')
    }

    targetUrl.searchParams.set('redirect', `${redirectUrl.pathname}${redirectUrl.search}`)
    return targetUrl.toString()
  } catch {
    return normalizedUrl
  }
}

function onPaymentDone() {
  const wasSubscription = paymentState.value.orderType === 'subscription'
  resetPayment()
  selectedPlan.value = null
  if (wasSubscription) {
    subscriptionStore.fetchActiveSubscriptions(true).catch(() => {})
  }
}

function onPaymentSuccess() {
  removeRecoverySnapshot()
  authStore.refreshUser()
  if (paymentState.value.orderType === 'subscription') {
    subscriptionStore.fetchActiveSubscriptions(true).catch(() => {})
  }
}

function onPaymentSettled() {
  removeRecoverySnapshot()
}

// All checkout data from single API call
const checkout = ref<CheckoutInfoResponse>({
  methods: {}, global_min: 0, global_max: 0,
  min_amount: 5, max_amount: 0,
  plans: [], balance_disabled: false, balance_recharge_multiplier: 1, recharge_fee_rate: 0, help_text: '', help_image_url: '', recharge_card_products: [],
  recharge_packages: [], stripe_publishable_key: '',
})

const tabs = computed(() => {
  const result: { key: 'recharge' | 'rechargeCard' | 'subscription' | 'redeem'; label: string }[] = []
  if (rechargeCardProducts.value.length > 0) result.push({ key: 'rechargeCard', label: t('payment.tabRechargeCard') })
  if (!checkout.value.balance_disabled) result.push({ key: 'recharge', label: t('payment.tabTopUp') })
  result.push({ key: 'subscription', label: t('payment.tabSubscribe') })
  result.push({ key: 'redeem', label: t('payment.tabRedeem') })
  return result
})

const visibleMethods = computed(() => getVisibleMethods(checkout.value.methods))
const enabledMethods = computed(() => Object.keys(visibleMethods.value))
const availableRechargePackages = computed(() => {
  const packages = checkout.value.recharge_packages || []
  return [...packages]
    .filter((item) => item.enabled !== false && item.amount > 0 && item.pay_amount > 0)
    .sort((a, b) => (a.sort_order ?? 0) - (b.sort_order ?? 0))
})
const selectedRechargePackage = computed(() =>
  availableRechargePackages.value.find((item) => item.id === selectedRechargePackageId.value) || null,
)
const hasRechargePackages = computed(() => availableRechargePackages.value.length > 0)
const validAmount = computed(() => selectedRechargePackage.value?.pay_amount ?? amount.value ?? 0)
const balanceRechargeMultiplier = computed(() => {
  const multiplier = checkout.value.balance_recharge_multiplier
  return multiplier > 0 ? multiplier : 1
})
const creditedAmount = computed(() => {
  if (selectedRechargePackage.value) {
    return Math.round(selectedRechargePackage.value.amount * 100) / 100
  }
  return Math.round((validAmount.value * balanceRechargeMultiplier.value) * 100) / 100
})

// Adaptive grid: center single card, 2-col for 2 plans, 3-col for 3+
const planGridClass = computed(() => {
  const n = checkout.value.plans.length
  if (n <= 2) return 'grid grid-cols-1 gap-5 sm:grid-cols-2'
  return 'grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3'
})

// Check if an amount fits a method's [min, max]. 0 = no limit.
function amountFitsMethod(amt: number, methodType: string): boolean {
  if (amt <= 0) return true
  const ml = visibleMethods.value[methodType]
  if (!ml) return false
  if (ml.single_min > 0 && amt < ml.single_min) return false
  if (ml.single_max > 0 && amt > ml.single_max) return false
  return true
}

// Selected method's limits (for validation and error messages)
const selectedLimit = computed(() => visibleMethods.value[selectedMethod.value])
const selectedCurrency = computed(() => normalizePaymentCurrency(selectedLimit.value?.currency))
const localeCode = computed(() => {
  const raw = i18n.locale as unknown
  if (typeof raw === 'string') return raw
  if (raw && typeof raw === 'object' && 'value' in raw) {
    return String((raw as { value?: string }).value || '')
  }
  return undefined
})

function formatSelectedPaymentAmount(value: number): string {
  return formatPaymentAmount(value, selectedCurrency.value, localeCode.value)
}

const methodOptions = computed<PaymentMethodOption[]>(() =>
  enabledMethods.value.map((type) => {
    const ml = visibleMethods.value[type]
    return {
      type,
      fee_rate: ml?.fee_rate ?? 0,
      available: ml?.available !== false && amountFitsMethod(validAmount.value, type),
    }
  }).sort((a, b) => {
    const order: readonly string[] = METHOD_ORDER
    const ai = order.indexOf(a.type)
    const bi = order.indexOf(b.type)
    return (ai === -1 ? 999 : ai) - (bi === -1 ? 999 : bi)
  })
)

const feeRate = computed(() => checkout.value?.recharge_fee_rate ?? 0)
const feeAmount = computed(() =>
  feeRate.value > 0 && validAmount.value > 0
    ? Math.ceil(((validAmount.value * feeRate.value) / 100) * 100) / 100
    : 0
)
const totalAmount = computed(() =>
  feeRate.value > 0 && validAmount.value > 0
    ? Math.round((validAmount.value + feeAmount.value) * 100) / 100
    : validAmount.value
)

const amountError = computed(() => {
  if (hasRechargePackages.value) return ''
  if (validAmount.value <= 0) return ''
  if (validAmount.value < rechargeMinAmount.value) return t('payment.amountTooLow', { min: formatSelectedPaymentAmount(rechargeMinAmount.value) })
  if (rechargeMaxAmount.value > 0 && validAmount.value > rechargeMaxAmount.value) return t('payment.amountTooHigh', { max: formatSelectedPaymentAmount(rechargeMaxAmount.value) })
  return ''
})

const minimumRechargeAmount = 5

const rechargeMinAmount = computed(() => Math.max(minimumRechargeAmount, checkout.value.min_amount || 0))
const rechargeMaxAmount = computed(() => checkout.value.max_amount || 0)

const quickRechargeAmounts = [5, 10, 50, 100, 200, 500, 1000]

const amountInputText = computed(() => amount.value == null ? '' : String(amount.value))

const amountPlaceholder = computed(() => {
  if (rechargeMinAmount.value > 0 && rechargeMaxAmount.value > 0) return `${t('payment.minimumAmountShort', { amount: rechargeMinAmount.value })}`
  if (rechargeMinAmount.value > 0) return t('payment.minimumAmountShort', { amount: rechargeMinAmount.value })
  if (rechargeMaxAmount.value > 0) return t('payment.maximumAmountShort', { amount: rechargeMaxAmount.value })
  return t('payment.enterAmount')
})

const quickRechargeOptions = computed(() =>
  quickRechargeAmounts
    .filter((value) =>
      value >= rechargeMinAmount.value &&
      (rechargeMaxAmount.value <= 0 || value <= rechargeMaxAmount.value),
    )
    .map((value) => ({ amount: value, badge: rechargeBonusBadge.value })),
)

const rechargeCardProducts = computed(() =>
  [...(checkout.value.recharge_card_products || [])]
    .filter((product) => product.enabled !== false && product.url)
    .sort((a, b) => (a.sort_order ?? 0) - (b.sort_order ?? 0)),
)

const rechargeBonusBadge = computed(() => {
  if (balanceRechargeMultiplier.value <= 1) return ''
  return t('payment.rechargeBonusBadge', { multiplier: balanceRechargeMultiplier.value.toFixed(2) })
})

const rechargeEstimateText = computed(() => {
  if (validAmount.value <= 0) return t('payment.rechargeAutoEstimate')
  if (feeRate.value > 0) {
    return t('payment.rechargePayEstimate', { amount: formatSelectedPaymentAmount(totalAmount.value) })
  }
  if (balanceRechargeMultiplier.value !== 1) {
    return t('payment.rechargeCreditEstimate', { amount: creditedAmount.value.toFixed(2) })
  }
  return t('payment.rechargeAutoEstimate')
})

function handleRechargeAmountInput(event: Event) {
  const raw = (event.target as HTMLInputElement).value
  if (raw === '') {
    amount.value = null
    return
  }
  const parsed = Number(raw)
  amount.value = Number.isFinite(parsed) && parsed > 0 ? Math.round(parsed * 100) / 100 : null
}

function selectRechargeAmount(value: number) {
  amount.value = value
  selectedRechargePackageId.value = ''
}

function selectRechargePackage(pkg: RechargePackage) {
  selectedRechargePackageId.value = pkg.id
  amount.value = pkg.pay_amount
}

function formatQuickAmount(value: number) {
  return formatSelectedPaymentAmount(value).replace(/\.00(?=\D*$)/, '')
}

function formatPackagePaymentAmount(value: number) {
  return formatPaymentAmount(value, selectedCurrency.value, localeCode.value).replace(/\.00(?=\D*$)/, '')
}

function chooseFirstAvailableMethod() {
  const next = methodOptions.value.find((method) => method.available)?.type
  if (next) {
    selectedMethod.value = next
  }
}

function rechargeCardProductMeta(product: { amount?: number; price?: number }) {
  const parts: string[] = []
  if (product.amount && product.amount > 0) {
    parts.push(t('payment.cardAmount', { amount: product.amount }))
  }
  if (product.price && product.price > 0) {
    parts.push(t('payment.cardPrice', { price: product.price }))
  }
  return parts.join(' · ') || t('payment.cardExternalPurchase')
}

function openRechargeCardDialog(product: RechargeCardProduct) {
  selectedRechargeCardProduct.value = product
}

function closeRechargeCardDialog() {
  selectedRechargeCardProduct.value = null
}

async function renderRechargeCardQr() {
  if (!selectedRechargeCardProduct.value?.url) return
  await nextTick()
  if (!rechargeCardQrCanvas.value) return
  await QRCode.toCanvas(rechargeCardQrCanvas.value, selectedRechargeCardProduct.value.url, {
    width: 208,
    margin: 2,
    color: {
      dark: '#111827',
      light: '#ffffff',
    },
  })
}

watch(
  () => selectedRechargeCardProduct.value?.url,
  () => {
    renderRechargeCardQr().catch(() => {})
  },
)

function paymentMethodIcon(type: string): string {
  if (type.includes('alipay')) return alipayIcon
  if (type.includes('wxpay')) return wxpayIcon
  if (type === 'stripe') return stripeIcon
  if (type === 'airwallex') return airwallexIcon
  if (type === 'creem') return stripeIcon
  if (type === 'easypay') return easypayIcon
  return alipayIcon
}

const canSubmit = computed(() =>
  validAmount.value > 0
    && amountError.value === ''
    && (!hasRechargePackages.value || !!selectedRechargePackage.value)
    && !!selectedMethod.value
    && !!selectedLimit.value
    && amountFitsMethod(validAmount.value, selectedMethod.value)
    && !checkout.value.balance_disabled
    && selectedLimit.value?.available !== false
)

// Subscription-specific: method options based on plan price
const subMethodOptions = computed<PaymentMethodOption[]>(() => {
  const planPrice = selectedPlan.value?.price ?? 0
  return enabledMethods.value.map((type) => {
    const ml = visibleMethods.value[type]
    return {
      type,
      fee_rate: ml?.fee_rate ?? 0,
      available: ml?.available !== false && amountFitsMethod(planPrice, type),
    }
  })
})

const subFeeAmount = computed(() => {
  const price = selectedPlan.value?.price ?? 0
  if (feeRate.value <= 0 || price <= 0) return 0
  return Math.ceil(((price * feeRate.value) / 100) * 100) / 100
})

const subTotalAmount = computed(() => {
  const price = selectedPlan.value?.price ?? 0
  if (feeRate.value <= 0 || price <= 0) return price
  return Math.round((price + subFeeAmount.value) * 100) / 100
})

const canSubmitSubscription = computed(() =>
  selectedPlan.value !== null
    && amountFitsMethod(selectedPlan.value.price, selectedMethod.value)
    && selectedLimit.value?.available !== false
)

// Auto-switch to first available method when current selection can't handle the amount
watch(() => [validAmount.value, selectedMethod.value] as const, ([amt, method]) => {
  if (amt <= 0) return
  const current = methodOptions.value.find((item) => item.type === method)
  if (current?.available) return
  chooseFirstAvailableMethod()
})

watch(availableRechargePackages, (packages) => {
  if (!packages.length) {
    selectedRechargePackageId.value = ''
    return
  }
  if (!packages.some((pkg) => pkg.id === selectedRechargePackageId.value)) {
    selectRechargePackage(packages[0])
  }
  chooseFirstAvailableMethod()
}, { immediate: true })

// Payment button class: follows selected payment method color
const paymentButtonClass = computed(() => {
  const m = selectedMethod.value
  if (!m) return 'btn-primary'
  if (m.includes('alipay')) return 'btn-alipay'
  if (m.includes('wxpay')) return 'btn-wxpay'
  if (m === 'stripe') return 'btn-stripe'
  if (m === 'airwallex') return 'btn-airwallex'
  if (m === 'creem') return 'btn-creem'
  return 'btn-primary'
})

// Subscription confirm: platform accent colors (clean card, no gradient)
const planBadgeClass = computed(() => platformBadgeClass(selectedPlan.value?.group_platform || ''))
const planTextClass = computed(() => platformTextClass(selectedPlan.value?.group_platform || ''))

// Renewal modal state
const showRenewalModal = ref(false)
const renewGroupId = ref<number | null>(null)
const renewalPlans = computed(() => {
  if (renewGroupId.value == null) return []
  return checkout.value.plans.filter(p => p.group_id === renewGroupId.value)
})

const planValiditySuffix = computed(() => {
  if (!selectedPlan.value) return ''
  const u = selectedPlan.value.validity_unit || 'day'
  if (u === 'month') return t('payment.perMonth')
  if (u === 'year') return t('payment.perYear')
  return `${selectedPlan.value.validity_days}${t('payment.days')}`
})

function selectPlan(plan: SubscriptionPlan) {
  selectedPlan.value = plan
  errorMessage.value = ''
}

function selectPlanFromModal(plan: SubscriptionPlan) {
  showRenewalModal.value = false
  renewGroupId.value = null
  selectedPlan.value = plan
  errorMessage.value = ''
}

function closeRenewalModal() {
  showRenewalModal.value = false
  renewGroupId.value = null
}

async function handleSubmitRecharge() {
  if (!canSubmit.value || submitting.value) return
  await createOrder(validAmount.value, 'balance', undefined, {
    rechargePackageId: selectedRechargePackage.value?.id,
  })
}

async function confirmSubscribe() {
  if (!selectedPlan.value || submitting.value) return
  await createOrder(selectedPlan.value.price, 'subscription', selectedPlan.value.id)
}

async function createOrder(orderAmount: number, orderType: OrderType, planId?: number, options: CreateOrderOptions = {}) {
  submitting.value = true
  errorMessage.value = ''
  errorHintMessage.value = ''
  const requestType = normalizeVisibleMethod(options.paymentType || selectedMethod.value) || options.paymentType || selectedMethod.value
  try {
    const payload = buildCreateOrderPayload({
      amount: orderAmount,
      rechargePackageId: options.rechargePackageId,
      paymentType: requestType,
      orderType,
      planId,
      origin: typeof window !== 'undefined' ? window.location.origin : '',
      isMobile: isMobileDevice(),
      isWechatBrowser: typeof window !== 'undefined' && /MicroMessenger/i.test(window.navigator.userAgent),
      forceQRCode: !!(checkout.value.alipay_force_qrcode && normalizeVisibleMethod(requestType) === 'alipay'),
    })
    if (options.openid) {
      payload.openid = options.openid
    }
    if (options.wechatResumeToken) {
      payload.wechat_resume_token = options.wechatResumeToken
    }

    const result = await paymentStore.createOrder(payload) as CreateOrderResult & { resume_token?: string }
    const openWindow = (url: string) => {
      const win = window.open(url, 'paymentPopup', getPaymentPopupFeatures())
      if (!win || win.closed) {
        window.location.href = url
      }
    }
    const visibleMethod = normalizeVisibleMethod(requestType) || requestType
    // When user clicks the dedicated Stripe button, leave method blank so the
    // landing page renders Stripe's full Payment Element (card/link/alipay/wxpay).
    const stripeMethod = visibleMethod === 'stripe'
      ? ''
      : visibleMethod === 'wxpay' ? 'wechat_pay' : 'alipay'
    const stripeRouteUrl = result.client_secret && visibleMethod !== 'airwallex'
      ? router.resolve({
        path: '/payment/stripe',
        query: {
          order_id: String(result.order_id),
          client_secret: result.client_secret,
          method: stripeMethod || undefined,
          resume_token: result.resume_token || undefined,
        },
      }).href
      : ''
    const airwallexRouteUrl = result.client_secret && result.intent_id
      ? router.resolve({
        path: '/payment/airwallex',
        query: {
          order_id: String(result.order_id),
          out_trade_no: result.out_trade_no || undefined,
          resume_token: result.resume_token || undefined,
        },
      }).href
      : ''
    const decision = decidePaymentLaunch(result, {
      visibleMethod,
      orderType,
      isMobile: isMobileDevice(),
      isWechatBrowser: typeof window !== 'undefined' && /MicroMessenger/i.test(window.navigator.userAgent),
      forceQRCode: !!(checkout.value.alipay_force_qrcode && visibleMethod === 'alipay'),
      stripePopupUrl: stripeRouteUrl,
      stripeRouteUrl,
      airwallexRouteUrl,
    })

    if (decision.kind === 'wechat_oauth' && decision.oauth?.authorize_url) {
      window.location.href = buildWechatOAuthAuthorizeUrl(decision.oauth.authorize_url, {
        paymentType: visibleMethod,
        orderType,
        planId,
        orderAmount,
        rechargePackageId: options.rechargePackageId,
      })
      return
    }

    if (decision.kind === 'unhandled') {
      applyScenarioError({ reason: 'UNHANDLED_PAYMENT_SCENARIO' }, visibleMethod)
      return
    }

    paymentState.value = decision.paymentState
    paymentPhase.value = 'paying'
    persistRecoverySnapshot(decision.recovery)

    if (decision.kind === 'stripe_popup') {
      openWindow(decision.paymentState.payUrl)
      return
    }
    if (decision.kind === 'stripe_route') {
      window.location.href = decision.paymentState.payUrl
      return
    }
    if (decision.kind === 'airwallex_route') {
      window.location.href = decision.paymentState.payUrl
      return
    }
    if (decision.kind === 'wechat_jsapi' && decision.jsapi) {
      try {
        const jsapiResult = await invokeWechatJsapiPayment(decision.jsapi as Record<string, unknown>)
        const errMsg = String(jsapiResult.err_msg || '').toLowerCase()
        if (errMsg.includes('cancel')) {
          appStore.showInfo(t('payment.qr.cancelled'))
          resetPayment()
        } else if (errMsg && !errMsg.includes('ok')) {
          resetPayment()
          const fallbackApplied = await attemptMobileQrFallback(
            { reason: 'WECHAT_JSAPI_FAILED', message: errMsg },
            {
              orderAmount,
              orderType,
              planId,
              rechargePackageId: options.rechargePackageId,
              paymentType: visibleMethod,
              attempted: options.mobileQrFallbackAttempted === true,
            },
          )
          if (!fallbackApplied) {
            applyScenarioError({ reason: 'WECHAT_JSAPI_FAILED', message: errMsg }, visibleMethod)
          }
        } else {
          const resultState = { ...decision.paymentState }
          resetPayment()
          await redirectToPaymentResult(resultState)
        }
      } catch (err: unknown) {
        resetPayment()
        const fallbackApplied = await attemptMobileQrFallback(err, {
          orderAmount,
          orderType,
          planId,
          rechargePackageId: options.rechargePackageId,
          paymentType: visibleMethod,
          attempted: options.mobileQrFallbackAttempted === true,
        })
        if (!fallbackApplied) {
          throw err
        }
      }
      return
    }
    if (decision.kind === 'redirect_waiting' && decision.paymentState.payUrl) {
      if (isMobileDevice()) {
        window.location.href = decision.paymentState.payUrl
        return
      }
      openWindow(decision.paymentState.payUrl)
    }
  } catch (err: unknown) {
    const apiErr = err as Record<string, unknown>
    if (apiErr.reason === 'TOO_MANY_PENDING') {
      const metadata = apiErr.metadata as Record<string, unknown> | undefined
      errorMessage.value = t('payment.errors.tooManyPending', { max: metadata?.max || '' })
      errorHintMessage.value = ''
    } else if (apiErr.reason === 'CANCEL_RATE_LIMITED') {
      errorMessage.value = t('payment.errors.cancelRateLimited')
      errorHintMessage.value = ''
    } else if (await attemptMobileQrFallback(err, {
      orderAmount,
      orderType,
      planId,
      rechargePackageId: options.rechargePackageId,
      paymentType: requestType,
      attempted: options.mobileQrFallbackAttempted === true,
    })) {
      return
    } else {
      const handled = applyScenarioError(
        err,
        normalizeVisibleMethod(options.paymentType || selectedMethod.value) || selectedMethod.value,
      )
      if (!handled) {
        errorMessage.value = extractI18nErrorMessage(err, t, 'payment.errors', extractApiErrorMessage(err, t('payment.result.failed')))
        errorHintMessage.value = ''
      }
      if (handled) {
        return
      }
    }
    appStore.showError(buildPaymentErrorToastMessage(errorMessage.value, errorHintMessage.value))
  } finally {
    submitting.value = false
  }
}

interface MobileQrFallbackContext {
  orderAmount: number
  orderType: OrderType
  planId?: number
  rechargePackageId?: string
  paymentType: string
  attempted: boolean
}

function shouldFallbackToDesktopQr(err: unknown, paymentMethod: string, attempted: boolean): boolean {
  if (attempted || !isMobileDevice()) {
    return false
  }

  const normalizedMethod = normalizeVisibleMethod(paymentMethod) || paymentMethod
  const reason = typeof err === 'object' && err && 'reason' in err && typeof err.reason === 'string'
    ? err.reason
    : ''
  const message = err instanceof Error
    ? err.message
    : (typeof err === 'object' && err && 'message' in err && typeof err.message === 'string'
      ? err.message
      : '')
  const normalizedMessage = message.toLowerCase()

  if (normalizedMethod === 'wxpay') {
    return reason === 'WECHAT_H5_NOT_AUTHORIZED'
      || reason === 'WECHAT_PAYMENT_MP_NOT_CONFIGURED'
      || reason === 'WECHAT_JSAPI_FAILED'
      || reason === 'PAYMENT_GATEWAY_ERROR'
      || reason === 'UNHANDLED_PAYMENT_SCENARIO'
      || normalizedMessage.includes('weixinjsbridge is unavailable')
      || normalizedMessage.includes('wechat_jsapi_unavailable')
  }

  if (normalizedMethod === 'alipay') {
    return reason === 'PAYMENT_GATEWAY_ERROR' || reason === 'UNHANDLED_PAYMENT_SCENARIO'
  }

  return false
}

async function attemptMobileQrFallback(err: unknown, context: MobileQrFallbackContext): Promise<boolean> {
  if (!shouldFallbackToDesktopQr(err, context.paymentType, context.attempted)) {
    return false
  }

  try {
    const visibleMethod = normalizeVisibleMethod(context.paymentType) || context.paymentType
    const payload = buildCreateOrderPayload({
      amount: context.orderAmount,
      rechargePackageId: context.rechargePackageId,
      paymentType: visibleMethod,
      orderType: context.orderType,
      planId: context.planId,
      origin: typeof window !== 'undefined' ? window.location.origin : '',
      isMobile: false,
      isWechatBrowser: false,
    })
    const result = await paymentStore.createOrder(payload) as CreateOrderResult & { resume_token?: string }
    const stripeMethod = visibleMethod === 'wxpay' ? 'wechat_pay' : 'alipay'
    const stripeRouteUrl = result.client_secret
      ? router.resolve({
        path: '/payment/stripe',
        query: {
          order_id: String(result.order_id),
          client_secret: result.client_secret,
          method: stripeMethod,
          resume_token: result.resume_token || undefined,
        },
      }).href
      : ''
    const decision = decidePaymentLaunch(result, {
      visibleMethod,
      orderType: context.orderType,
      isMobile: false,
      isWechatBrowser: false,
      stripePopupUrl: stripeRouteUrl,
      stripeRouteUrl,
    })

    if (decision.kind !== 'qr_waiting' || !decision.paymentState.qrCode) {
      return false
    }

    errorMessage.value = ''
    errorHintMessage.value = ''
    paymentState.value = decision.paymentState
    paymentPhase.value = 'paying'
    persistRecoverySnapshot(decision.recovery)
    appStore.showWarning(t('payment.errors.mobilePaymentFallbackToQr'))
    return true
  } catch {
    return false
  }
}

function applyScenarioError(err: unknown, paymentMethod: string): boolean {
  const descriptor = describePaymentScenarioError(err, {
    paymentMethod,
    isMobile: isMobileDevice(),
    isWechatBrowser: typeof window !== 'undefined' && /MicroMessenger/i.test(window.navigator.userAgent),
  })
  if (!descriptor) {
    errorMessage.value = ''
    errorHintMessage.value = ''
    return false
  }
  errorMessage.value = t(descriptor.messageKey)
  errorHintMessage.value = descriptor.hintKey ? t(descriptor.hintKey) : ''
  appStore.showError(buildPaymentErrorToastMessage(errorMessage.value, errorHintMessage.value))
  return true
}

async function resumeWechatPaymentFromQuery() {
  const resume = parseWechatResumeRoute(route.query, checkout.value.plans, validAmount.value)
  if (!resume) {
    return
  }

  selectedMethod.value = resume.paymentType
  if (resume.rechargePackageId) {
    selectedRechargePackageId.value = resume.rechargePackageId
  }
  if (resume.orderType === 'balance' && resume.orderAmount > 0) {
    amount.value = resume.orderAmount
  }
  if (resume.orderType === 'subscription' && resume.planId) {
    selectedPlan.value = checkout.value.plans.find(plan => plan.id === resume.planId) ?? null
  }

  await router.replace({ path: route.path, query: stripWechatResumeQuery(route.query) })

  if (resume.wechatResumeToken) {
    await createOrder(0, resume.orderType, resume.planId, {
      wechatResumeToken: resume.wechatResumeToken,
      paymentType: resume.paymentType,
      rechargePackageId: resume.rechargePackageId,
      isResume: true,
    })
    return
  }

  if (resume.orderAmount > 0 && resume.openid) {
    await createOrder(resume.orderAmount, resume.orderType, resume.planId, {
      openid: resume.openid,
      paymentType: resume.paymentType,
      rechargePackageId: resume.rechargePackageId,
      isResume: true,
    })
  }
}

onMounted(async () => {
  try {
    const res = await paymentAPI.getCheckoutInfo()
    checkout.value = res.data
    if (enabledMethods.value.length) {
      chooseFirstAvailableMethod()
    }
    if (typeof window !== 'undefined') {
      if (hasWechatResumeQuery(route.query)) {
        removeRecoverySnapshot()
      }
      const routeResumeToken = typeof route.query.resume_token === 'string'
        ? route.query.resume_token
        : typeof route.query.wechat_resume_token === 'string'
          ? route.query.wechat_resume_token
          : undefined
      const restored = readPaymentRecoverySnapshot(
        window.localStorage.getItem(PAYMENT_RECOVERY_STORAGE_KEY),
        { resumeToken: routeResumeToken },
      )
      if (restored) {
        paymentState.value = restored
        paymentPhase.value = 'paying'
        const restoredMethod = normalizeVisibleMethod(restored.paymentType)
        if (restoredMethod) {
          selectedMethod.value = restoredMethod
        }
      } else {
        removeRecoverySnapshot()
      }
    }
    await resumeWechatPaymentFromQuery()
    if (rechargeCardProducts.value.length > 0) {
      activeTab.value = 'rechargeCard'
    } else if (checkout.value.balance_disabled) {
      activeTab.value = 'subscription'
    }
    // Handle renewal navigation: ?tab=subscription&group=123
    if (route.query.tab === 'subscription') {
      activeTab.value = 'subscription'
      if (route.query.group) {
        const groupId = Number(route.query.group)
        const groupPlans = checkout.value.plans.filter(p => p.group_id === groupId)
        if (groupPlans.length === 1) {
          selectedPlan.value = groupPlans[0]
        } else if (groupPlans.length > 1) {
          renewGroupId.value = groupId
          showRenewalModal.value = true
        }
      }
    }
  } catch (err: unknown) { appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error'))) }
  finally { loading.value = false }
  // Fetch active subscriptions (uses cache, non-blocking)
  subscriptionStore.fetchActiveSubscriptions().catch(() => {})
})
</script>
