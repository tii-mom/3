<template>
  <AppLayout>
  <div class="space-y-6">
    <section class="rounded-xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-700 dark:bg-dark-900">
      <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
        <div>
          <h1 class="text-xl font-bold text-gray-950 dark:text-white">商城管理</h1>
          <p class="mt-1.5 text-sm text-gray-500 dark:text-gray-400">商品上架、品类配置、轮播展示与订单发货；用户付款后在订单里填写发货信息，推广返点进入推广钱包。</p>
        </div>
        <div class="flex flex-wrap gap-2">
          <button v-if="tab === 'products'" class="btn-primary rounded-lg px-4 py-2" @click="openProductDialog()">新增商品</button>
          <button v-else-if="tab === 'categories'" class="btn-primary rounded-lg px-4 py-2" @click="openCategoryDialog()">新增品类</button>
          <button v-else-if="tab === 'banners'" class="btn-primary rounded-lg px-4 py-2" @click="openBannerDialog()">新增轮播</button>
        </div>
      </div>
      <div class="mt-5 flex gap-1.5 overflow-x-auto rounded-lg bg-gray-100 p-1 dark:bg-dark-800">
        <button v-for="item in tabs" :key="item.key" class="whitespace-nowrap rounded-md px-3.5 py-1.5 text-sm font-medium transition" :class="tab === item.key ? 'bg-white text-gray-950 shadow-sm dark:bg-dark-700 dark:text-white' : 'text-gray-500 hover:text-gray-800 dark:text-gray-400 dark:hover:text-gray-200'" @click="tab = item.key">
          {{ item.label }}
        </button>
      </div>
    </section>

    <section v-if="tab === 'products'" class="rounded-xl border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-900">
      <div class="hidden overflow-x-auto lg:block">
        <table class="min-w-full divide-y divide-gray-100 text-sm dark:divide-dark-700">
          <thead class="bg-gray-50 text-left text-xs uppercase text-gray-500 dark:bg-dark-800 dark:text-gray-400">
            <tr>
              <th class="px-5 py-3">商品</th>
              <th class="px-5 py-3">价格</th>
              <th class="px-5 py-3">类型</th>
              <th class="px-5 py-3">佣金</th>
              <th class="px-5 py-3">状态</th>
              <th class="px-5 py-3 text-right">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
            <tr v-for="product in products" :key="product.id">
              <td class="px-5 py-4">
                <div class="flex items-center gap-3">
                  <img v-if="shopImage(product.image_url)" :src="shopImage(product.image_url)" class="h-12 w-12 rounded-lg object-cover" alt="">
                  <div v-else class="shop-thumb h-12 w-12 rounded-lg" aria-hidden="true">3</div>
                  <div class="min-w-0">
                    <div class="flex flex-wrap items-center gap-2">
                      <span class="font-semibold text-gray-950 dark:text-white">{{ product.name }}</span>
                      <span class="rounded-full bg-gray-100 px-2 py-0.5 text-[11px] font-medium text-gray-600 dark:bg-dark-800 dark:text-gray-300">{{ categoryLabel(product.category) }}</span>
                    </div>
                    <div class="line-clamp-1 text-xs text-gray-500">{{ product.description }}</div>
                  </div>
                </div>
              </td>
              <td class="px-5 py-4 font-semibold">¥{{ money(product.price_cny_minor) }}</td>
              <td class="px-5 py-4">{{ typeLabel(product.product_type) }}</td>
              <td class="px-5 py-4">{{ (product.commission_bps / 100).toFixed(0) }}%</td>
              <td class="px-5 py-4"><span :class="statusClass(product.status)" class="rounded-full px-2.5 py-1 text-xs font-semibold">{{ statusLabel(product.status) }}</span></td>
              <td class="px-5 py-4 text-right">
                <button class="btn-secondary mr-2 rounded-lg px-3 py-1.5 text-xs" @click="openProductDialog(product)">编辑</button>
                <button class="rounded-lg px-3 py-1.5 text-xs font-semibold text-red-600 hover:bg-red-50 dark:hover:bg-red-500/10" @click="removeProduct(product)">删除</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- 移动端：表格列太密，改为卡片列表 -->
      <div class="divide-y divide-gray-100 dark:divide-dark-700 lg:hidden">
        <article v-for="product in products" :key="product.id" class="p-4">
          <div class="flex gap-3">
            <img v-if="shopImage(product.image_url)" :src="shopImage(product.image_url)" class="h-14 w-14 shrink-0 rounded-lg object-cover" alt="">
            <div v-else class="shop-thumb h-14 w-14 shrink-0 rounded-lg" aria-hidden="true">3</div>
            <div class="min-w-0 flex-1">
              <div class="flex items-start justify-between gap-2">
                <div class="min-w-0">
                  <div class="truncate font-semibold text-gray-950 dark:text-white">{{ product.name }}</div>
                  <div class="mt-0.5 line-clamp-2 text-xs text-gray-500">{{ product.description }}</div>
                  <div class="mt-1 text-[11px] font-medium text-gray-500 dark:text-gray-400">{{ categoryLabel(product.category) }}</div>
                </div>
                <span :class="statusClass(product.status)" class="shrink-0 rounded-full px-2 py-0.5 text-xs font-semibold">{{ statusLabel(product.status) }}</span>
              </div>
              <div class="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-gray-500 dark:text-gray-400">
                <span class="text-sm font-semibold text-gray-950 tabular-nums dark:text-white">¥{{ money(product.price_cny_minor) }}</span>
                <span>{{ typeLabel(product.product_type) }}</span>
                <span>佣金 {{ (product.commission_bps / 100).toFixed(0) }}%</span>
              </div>
              <div class="mt-3 flex gap-2">
                <button class="btn-secondary rounded-lg px-3 py-1.5 text-xs" @click="openProductDialog(product)">编辑</button>
                <button class="rounded-lg px-3 py-1.5 text-xs font-semibold text-red-600 hover:bg-red-50 dark:hover:bg-red-500/10" @click="removeProduct(product)">删除</button>
              </div>
            </div>
          </div>
        </article>
      </div>
    </section>

    <section v-else-if="tab === 'categories'" class="rounded-xl border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-900">
      <div class="border-b border-gray-100 px-5 py-4 dark:border-dark-700">
        <h2 class="text-lg font-bold text-gray-950 dark:text-white">商品品类</h2>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
          首页「选择套餐」的横版分类导航就按这里的顺序展示。关闭某个品类，它会连同该品类下的商品一起从首页隐藏（商品本身不会被删除）。
        </p>
      </div>

      <div class="hidden overflow-x-auto lg:block">
        <table class="min-w-full divide-y divide-gray-100 text-sm dark:divide-dark-700">
          <thead class="bg-gray-50 text-left text-xs uppercase text-gray-500 dark:bg-dark-800 dark:text-gray-400">
            <tr>
              <th class="px-5 py-3">品类</th>
              <th class="px-5 py-3">标识</th>
              <th class="px-5 py-3">商品数</th>
              <th class="px-5 py-3">排序</th>
              <th class="px-5 py-3">状态</th>
              <th class="px-5 py-3 text-right">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
            <tr v-for="category in adminCategories" :key="category.id">
              <td class="px-5 py-4">
                <div class="font-semibold text-gray-950 dark:text-white">{{ category.label }}</div>
                <div class="line-clamp-1 text-xs text-gray-500">{{ category.blurb || '—' }}</div>
              </td>
              <td class="px-5 py-4"><code class="rounded bg-gray-100 px-1.5 py-0.5 text-xs text-gray-600 dark:bg-dark-800 dark:text-gray-300">{{ category.slug }}</code></td>
              <td class="px-5 py-4 tabular-nums">{{ category.product_count }}</td>
              <td class="px-5 py-4 tabular-nums">{{ category.sort_order }}</td>
              <td class="px-5 py-4">
                <span :class="category.enabled ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-300' : 'bg-gray-100 text-gray-500 dark:bg-dark-800 dark:text-gray-300'" class="rounded-full px-2.5 py-1 text-xs font-semibold">
                  {{ category.enabled ? '展示中' : '已隐藏' }}
                </span>
              </td>
              <td class="px-5 py-4 text-right">
                <button class="btn-secondary mr-2 rounded-lg px-3 py-1.5 text-xs" @click="openCategoryDialog(category)">编辑</button>
                <button class="rounded-lg px-3 py-1.5 text-xs font-semibold text-red-600 hover:bg-red-50 dark:hover:bg-red-500/10" @click="removeCategory(category)">删除</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- 移动端：表格列太窄，改为卡片列表 -->
      <div class="divide-y divide-gray-100 dark:divide-dark-700 lg:hidden">
        <article v-for="category in adminCategories" :key="category.id" class="p-4">
          <div class="flex items-start justify-between gap-2">
            <div class="min-w-0">
              <div class="truncate font-semibold text-gray-950 dark:text-white">{{ category.label }}</div>
              <code class="mt-1 inline-block rounded bg-gray-100 px-1.5 py-0.5 text-xs text-gray-600 dark:bg-dark-800 dark:text-gray-300">{{ category.slug }}</code>
              <div class="mt-1 line-clamp-2 text-xs text-gray-500">{{ category.blurb || '—' }}</div>
            </div>
            <span :class="category.enabled ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-300' : 'bg-gray-100 text-gray-500 dark:bg-dark-800 dark:text-gray-300'" class="shrink-0 rounded-full px-2 py-0.5 text-xs font-semibold">
              {{ category.enabled ? '展示中' : '已隐藏' }}
            </span>
          </div>
          <div class="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-gray-500 dark:text-gray-400">
            <span>商品 {{ category.product_count }}</span>
            <span>排序 {{ category.sort_order }}</span>
          </div>
          <div class="mt-3 flex gap-2">
            <button class="btn-secondary rounded-lg px-3 py-1.5 text-xs" @click="openCategoryDialog(category)">编辑</button>
            <button class="rounded-lg px-3 py-1.5 text-xs font-semibold text-red-600 hover:bg-red-50 dark:hover:bg-red-500/10" @click="removeCategory(category)">删除</button>
          </div>
        </article>
      </div>

      <div v-if="!adminCategories.length" class="px-5 py-14 text-center text-sm text-gray-500 dark:text-gray-400">
        还没有品类，点右上角「新增品类」创建第一个。
      </div>
    </section>

    <section v-else-if="tab === 'banners'" class="grid gap-4 lg:grid-cols-2">
      <article v-for="banner in banners" :key="banner.id" class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-900">
        <img v-if="shopImage(banner.image_url)" :src="shopImage(banner.image_url)" class="h-40 w-full object-cover" alt="">
        <div v-else class="shop-thumb h-40 w-full" aria-hidden="true">3</div>
        <div class="p-5">
          <div class="flex items-start justify-between gap-3">
            <div>
              <h3 class="font-bold text-gray-950 dark:text-white">{{ banner.title }}</h3>
              <p class="mt-1 text-sm text-gray-500">{{ banner.subtitle }}</p>
            </div>
            <span class="rounded-full px-2.5 py-1 text-xs font-semibold" :class="banner.enabled ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-300' : 'bg-gray-100 text-gray-500 dark:bg-dark-800'">
              {{ banner.enabled ? '展示中' : '已关闭' }}
            </span>
          </div>
          <div class="mt-4 flex justify-end gap-2">
            <button class="btn-secondary rounded-lg px-3 py-1.5 text-xs" @click="openBannerDialog(banner)">编辑</button>
            <button class="rounded-lg px-3 py-1.5 text-xs font-semibold text-red-600 hover:bg-red-50 dark:hover:bg-red-500/10" @click="removeBanner(banner)">删除</button>
          </div>
        </div>
      </article>
    </section>

    <section v-else class="rounded-xl border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-900">
      <div class="flex flex-wrap items-center justify-between gap-3 border-b border-gray-100 px-5 py-4 dark:border-dark-700">
        <div>
          <h2 class="text-lg font-bold text-gray-950 dark:text-white">商城订单</h2>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">付款后先进入待发货，管理员填写内容后再发货给用户。</p>
        </div>
        <div class="flex flex-wrap gap-2 text-xs font-semibold">
          <span class="rounded-full bg-amber-50 px-3 py-1.5 text-amber-700 dark:bg-amber-500/10 dark:text-amber-200">待发货 {{ pendingOrdersCount }}</span>
          <span class="rounded-full bg-emerald-50 px-3 py-1.5 text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-200">已发货 {{ fulfilledOrdersCount }}</span>
        </div>
      </div>
      <div class="hidden overflow-x-auto lg:block">
        <table class="min-w-full divide-y divide-gray-100 text-sm dark:divide-dark-700">
          <thead class="bg-gray-50 text-left text-xs uppercase text-gray-500 dark:bg-dark-800 dark:text-gray-400">
            <tr>
              <th class="px-5 py-3">订单</th>
              <th class="px-5 py-3">用户</th>
              <th class="px-5 py-3">金额</th>
              <th class="px-5 py-3">状态</th>
              <th class="px-5 py-3">佣金</th>
              <th class="px-5 py-3">时间</th>
              <th class="px-5 py-3 text-right">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
            <tr v-for="order in orders" :key="order.id">
              <td class="px-5 py-4">
                <div class="font-semibold text-gray-950 dark:text-white">#{{ order.id }} {{ order.snapshot_name }}</div>
                <div class="text-xs text-gray-500">支付订单：{{ order.payment_order_id || '-' }}</div>
              </td>
              <td class="px-5 py-4">
                <div>{{ order.user_email || `用户 #${order.user_id}` }}</div>
                <div v-if="order.guest_contact" class="mt-1 text-xs font-medium text-emerald-600 dark:text-emerald-300">
                  游客联系方式：{{ order.guest_contact }}
                </div>
                <div v-if="order.order_no" class="mt-0.5 text-xs text-gray-500">单号 {{ order.order_no }}</div>
              </td>
              <td class="px-5 py-4">
                <div class="font-semibold">¥{{ money(order.snapshot_price_cny_minor) }}</div>
                <div v-if="order.wallet_applied_cny_minor > 0" class="mt-0.5 text-xs text-orange-600 dark:text-orange-300">
                  返点抵扣 ¥{{ money(order.wallet_applied_cny_minor) }} · 实付 ¥{{ money(order.payable_cny_minor) }}
                </div>
              </td>
              <td class="px-5 py-4">
                <div class="font-semibold text-gray-900 dark:text-white">{{ orderStatusLabel(order) }}</div>
                <div v-if="order.fulfillment_note" class="mt-1 max-w-xs truncate text-xs text-amber-600 dark:text-amber-300">发货内容：{{ order.fulfillment_note }}</div>
              </td>
              <td class="px-5 py-4">{{ order.snapshot_commission_bps > 0 ? `${(order.snapshot_commission_bps / 100).toFixed(0)}%` : '无' }}</td>
              <td class="px-5 py-4 text-xs text-gray-500">{{ formatDate(order.created_at) }}</td>
              <td class="px-5 py-4 text-right">
                <button
                  v-if="canFulfillOrder(order)"
                  type="button"
                  class="btn-primary rounded-lg px-3 py-1.5 text-xs"
                  @click="openFulfillDialog(order)"
                >
                  手动发货
                </button>
                <span v-else class="text-xs text-gray-400">-</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- 移动端：订单卡片 -->
      <div class="divide-y divide-gray-100 dark:divide-dark-700 lg:hidden">
        <article v-for="order in orders" :key="order.id" class="p-4">
          <div class="flex items-start justify-between gap-2">
            <div class="min-w-0">
              <div class="truncate font-semibold text-gray-950 dark:text-white">#{{ order.id }} {{ order.snapshot_name }}</div>
              <div class="mt-0.5 truncate text-xs text-gray-500">{{ order.user_email || `用户 #${order.user_id}` }}</div>
              <div v-if="order.guest_contact" class="mt-0.5 truncate text-xs font-medium text-emerald-600 dark:text-emerald-300">游客联系：{{ order.guest_contact }}</div>
            </div>
            <div class="shrink-0 text-right">
              <div class="text-sm font-semibold tabular-nums text-gray-950 dark:text-white">¥{{ money(order.snapshot_price_cny_minor) }}</div>
              <div v-if="order.wallet_applied_cny_minor > 0" class="mt-0.5 text-xs text-orange-600 dark:text-orange-300">
                抵扣 ¥{{ money(order.wallet_applied_cny_minor) }} · 实付 ¥{{ money(order.payable_cny_minor) }}
              </div>
              <div class="mt-1 text-xs text-gray-500">{{ orderStatusLabel(order) }}</div>
            </div>
          </div>
          <div v-if="order.fulfillment_note" class="mt-2 line-clamp-2 rounded-lg bg-amber-50 px-3 py-2 text-xs text-amber-700 dark:bg-amber-500/10 dark:text-amber-200">
            发货内容：{{ order.fulfillment_note }}
          </div>
          <div class="mt-3 flex items-center justify-between gap-2">
            <span class="text-xs text-gray-500">{{ formatDate(order.created_at) }}</span>
            <button
              v-if="canFulfillOrder(order)"
              type="button"
              class="btn-primary rounded-lg px-3 py-1.5 text-xs"
              @click="openFulfillDialog(order)"
            >
              手动发货
            </button>
          </div>
        </article>
      </div>
    </section>

    <div v-if="productDialog.open" class="fixed inset-0 z-50 flex items-center justify-center bg-gray-950/55 p-4 backdrop-blur-sm">
      <form class="shop-modal max-h-[92vh] w-full max-w-4xl overflow-y-auto rounded-xl bg-white p-5 shadow-2xl dark:bg-dark-900" @submit.prevent="saveProduct">
        <div class="flex items-start justify-between gap-4">
          <div>
            <p class="text-sm font-semibold text-primary-600 dark:text-primary-300">商品管理</p>
            <h3 class="mt-1 text-xl font-bold text-gray-950 dark:text-white">{{ productDialog.id ? '编辑商品' : '新增商品' }}</h3>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">填写商品展示、价格、库存和推广佣金，保存后即可在商城展示。</p>
          </div>
          <button type="button" class="rounded-full p-2 text-gray-400 hover:bg-gray-100 hover:text-gray-700 dark:hover:bg-dark-800 dark:hover:text-white" aria-label="关闭" @click="productDialog.open = false">✕</button>
        </div>

        <div class="mt-6 grid gap-5 lg:grid-cols-[1fr_18rem]">
          <div class="space-y-4">
            <label class="shop-field">
              <span>商品名称</span>
              <input v-model="productForm.name" class="input-field w-full" placeholder="例如：ChatGPT 桌面端安装服务" required>
            </label>
            <label class="shop-field">
              <span>商品描述</span>
              <textarea v-model="productForm.description" class="input-field min-h-28 w-full" placeholder="写清楚用户购买后能获得什么服务或商品。"></textarea>
            </label>
            <div class="grid gap-4 sm:grid-cols-2">
              <label class="shop-field">
                <span>售卖金额（元）</span>
                <input v-model.number="productPrice" type="number" min="0.01" step="0.01" class="input-field w-full" required>
              </label>
              <label class="shop-field">
                <span>划线原价（元）</span>
                <input v-model.number="productOriginalPrice" type="number" min="0" step="0.01" class="input-field w-full" placeholder="可不填">
              </label>
              <label class="shop-field">
                <span>推广佣金（%）</span>
                <input v-model.number="productCommissionPercent" type="number" min="0" max="100" step="1" class="input-field w-full">
              </label>
              <label class="shop-field">
                <span>库存（留空不限）</span>
                <input v-model.number="productForm.stock_quantity" type="number" min="0" class="input-field w-full" placeholder="不限">
              </label>
              <label class="shop-field">
                <span>状态</span>
                <select v-model="productForm.status" class="input-field w-full">
                  <option value="draft">草稿</option>
                  <option value="published">上架</option>
                  <option value="archived">归档</option>
                </select>
              </label>
            </div>

            <div class="mt-4 grid gap-4 md:grid-cols-2">
              <!-- 这里刻意不用 .shop-field：它的 `> span { display: block }` 会盖掉 Tailwind 的
                   flex 工具类（两处都是单类选择器，组件样式在 utilities 之后加载），
                   标题与「管理品类」会挤成一行。 -->
              <div>
                <div class="mb-1.5 flex items-center justify-between gap-2">
                  <span class="text-sm font-bold text-gray-700 dark:text-gray-200">商品品类</span>
                  <button
                    type="button"
                    class="text-xs font-semibold text-primary-600 hover:underline dark:text-primary-300"
                    @click="goToCategoriesTab"
                  >
                    管理品类
                  </button>
                </div>
                <select v-model="productForm.category" class="input-field w-full">
                  <option v-for="option in categorySelectOptions" :key="option.value" :value="option.value">
                    {{ option.label }}{{ option.count >= 0 ? `（${option.count}）` : '' }}
                  </option>
                </select>
              </div>
              <label class="shop-field">
                <span>交付模式</span>
                <select v-model="productForm.fulfillment_mode" class="input-field w-full">
                  <option value="manual">人工处理</option>
                  <option value="session_topup">代充值（需 Session）</option>
                  <option value="account_delivery">成品号（发账号）</option>
                  <option value="rental">租号（按时长）</option>
                </select>
              </label>
              <label class="shop-field">
                <span>角标文字</span>
                <input v-model="productForm.badge_text" type="text" class="input-field w-full" placeholder="如：最热门，留空不显示">
              </label>
              <label class="shop-field">
                <span>交付提示</span>
                <input v-model="productForm.delivery_form_hint" type="text" class="input-field w-full" placeholder="如：提交 Session 后 1-3 分钟到账">
              </label>
              <label class="shop-field">
                <span>规格说明</span>
                <input v-model="productForm.spec_label" type="text" class="input-field w-full" placeholder="如：独享 · 30 天质保">
              </label>
              <label class="flex cursor-pointer flex-row items-center gap-2 md:col-span-2">
                <input v-model="productForm.highlight" type="checkbox" class="h-4 w-4 rounded border-gray-300">
                <span class="text-sm font-medium text-gray-700 dark:text-gray-200">在官网首页高亮推荐</span>
              </label>
            </div>
          </div>

          <aside class="rounded-xl border border-gray-200 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-800/70">
            <p class="text-sm font-semibold text-gray-900 dark:text-white">商品图片</p>
            <div class="mt-3 overflow-hidden rounded-2xl border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-900">
              <img v-if="shopImage(productForm.image_url)" :src="shopImage(productForm.image_url)" alt="商品预览图" class="h-44 w-full object-cover">
              <div v-else class="shop-thumb h-44 w-full" aria-hidden="true">3</div>
            </div>
            <input ref="productFileInput" type="file" accept="image/jpeg,image/png,image/webp,image/gif" class="hidden" @change="handleProductImageChange">
            <button type="button" class="btn-secondary mt-3 w-full justify-center rounded-2xl px-4 py-2.5" :disabled="uploadingProductImage" @click="productFileInput?.click()">
              {{ uploadingProductImage ? '上传中...' : '本地上传图片' }}
            </button>
            <label class="shop-field mt-3">
              <span>或粘贴图片地址</span>
              <input v-model="productForm.image_url" class="input-field w-full" placeholder="https://...">
            </label>
            <p class="mt-3 text-xs leading-5 text-gray-500 dark:text-gray-400">首页把商品图当作 46px 的方形图标展示，所以本地上传会自动居中裁成正方形并压缩成 WebP（{{ PRODUCT_IMAGE_SIZE }}×{{ PRODUCT_IMAGE_SIZE }}）。支持 JPG、PNG、WebP、GIF，原图不超过 12MB。</p>

            <div class="mt-5 border-t border-gray-200 pt-4 dark:border-dark-700">
              <div class="flex items-center justify-between gap-2">
                <p class="text-sm font-semibold text-gray-900 dark:text-white">商品图廊</p>
                <span class="text-xs text-gray-500 dark:text-gray-400">{{ productForm.gallery?.length || 0 }}/6</span>
              </div>
              <p class="mt-1 text-xs leading-5 text-gray-500 dark:text-gray-400">下单抽屉里展示的多张细节图，最多 6 张。</p>
              <div v-if="(productForm.gallery?.length || 0) > 0" class="mt-3 space-y-2">
                <div v-for="(url, index) in productForm.gallery" :key="`gallery-${index}`" class="flex items-center gap-2">
                  <img v-if="shopImage(url)" :src="shopImage(url)" alt="图廊预览" class="h-9 w-9 shrink-0 rounded-lg object-cover">
                  <div v-else class="shop-thumb h-9 w-9 shrink-0 rounded-lg text-xs" aria-hidden="true">图</div>
                  <input v-model="productForm.gallery[index]" class="input-field min-w-0 flex-1" placeholder="https://...">
                  <button type="button" class="shrink-0 rounded-lg px-2 py-1 text-sm text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-500/10" aria-label="移除该图" @click="removeGalleryItem(index)">移除</button>
                </div>
              </div>
              <button type="button" class="btn-secondary mt-3 w-full justify-center rounded-2xl px-4 py-2" :disabled="(productForm.gallery?.length || 0) >= 6" @click="addGalleryItem">
                添加图片地址
              </button>
            </div>
          </aside>
        </div>
        <div class="mt-6 flex justify-end gap-3">
          <button type="button" class="btn-secondary rounded-2xl px-4 py-2.5" @click="productDialog.open = false">取消</button>
          <button class="btn-primary rounded-2xl px-4 py-2.5">保存</button>
        </div>
      </form>
    </div>

    <div v-if="categoryDialog.open" class="fixed inset-0 z-50 flex items-center justify-center bg-gray-950/55 p-4 backdrop-blur-sm">
      <form class="shop-modal max-h-[92vh] w-full max-w-xl overflow-y-auto rounded-xl bg-white p-5 shadow-2xl dark:bg-dark-900" @submit.prevent="saveCategory">
        <div class="flex items-start justify-between gap-4">
          <div>
            <p class="text-sm font-semibold text-primary-600 dark:text-primary-300">品类管理</p>
            <h3 class="mt-1 text-xl font-bold text-gray-950 dark:text-white">{{ categoryDialog.id ? '编辑品类' : '新增品类' }}</h3>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">品类决定首页「选择套餐」的横版分类导航，也决定商品卡片的品牌图标与权益文案。</p>
          </div>
          <button type="button" class="rounded-full p-2 text-gray-400 hover:bg-gray-100 hover:text-gray-700 dark:hover:bg-dark-800 dark:hover:text-white" aria-label="关闭" @click="categoryDialog.open = false">✕</button>
        </div>

        <div class="mt-6 space-y-4">
          <label class="shop-field">
            <span>品类名称</span>
            <input v-model="categoryForm.label" class="input-field w-full" placeholder="例如：GPT 官方充值" required>
          </label>
          <label class="shop-field">
            <span>品类标识（slug）</span>
            <input
              v-model="categoryForm.slug"
              class="input-field w-full disabled:cursor-not-allowed disabled:opacity-60"
              :disabled="!!categoryDialog.id"
              placeholder="例如：gpt_topup"
              autocomplete="off"
            >
            <small class="mt-1 block text-xs font-normal leading-5 text-gray-500 dark:text-gray-400">
              {{ categoryDialog.id ? '标识创建后不可修改，避免打乱已上架商品的归属。' : '小写字母开头，只能用小写字母、数字与下划线，如 gpt_topup / x_premium。' }}
            </small>
          </label>
          <label class="shop-field">
            <span>一句话说明</span>
            <input v-model="categoryForm.blurb" class="input-field w-full" placeholder="展示在首页分组标题下，建议 20 字以内">
          </label>
          <div class="grid gap-4 sm:grid-cols-2">
            <label class="shop-field">
              <span>排序（越小越靠前）</span>
              <input v-model.number="categoryForm.sort_order" type="number" class="input-field w-full">
            </label>
            <label class="flex cursor-pointer flex-row items-center gap-2 sm:pt-7">
              <input v-model="categoryForm.enabled" type="checkbox" class="h-4 w-4 rounded border-gray-300">
              <span class="text-sm font-medium text-gray-700 dark:text-gray-200">在首页展示该品类</span>
            </label>
          </div>
        </div>

        <div class="mt-6 flex justify-end gap-3">
          <button type="button" class="btn-secondary rounded-2xl px-4 py-2.5" @click="categoryDialog.open = false">取消</button>
          <button class="btn-primary rounded-2xl px-4 py-2.5" :disabled="savingCategory">{{ savingCategory ? '保存中...' : '保存' }}</button>
        </div>
      </form>
    </div>

    <div v-if="bannerDialog.open" class="fixed inset-0 z-50 flex items-center justify-center bg-gray-950/55 p-4 backdrop-blur-sm">
      <form class="shop-modal w-full max-w-3xl rounded-xl bg-white p-5 shadow-2xl dark:bg-dark-900" @submit.prevent="saveBanner">
        <div class="flex items-start justify-between gap-4">
          <div>
            <p class="text-sm font-semibold text-primary-600 dark:text-primary-300">轮播管理</p>
            <h3 class="mt-1 text-xl font-bold text-gray-950 dark:text-white">{{ bannerDialog.id ? '编辑轮播' : '新增轮播' }}</h3>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">设置商城顶部展示图，可关联某个商品。</p>
          </div>
          <button type="button" class="rounded-full p-2 text-gray-400 hover:bg-gray-100 hover:text-gray-700 dark:hover:bg-dark-800 dark:hover:text-white" aria-label="关闭" @click="bannerDialog.open = false">✕</button>
        </div>
        <div class="mt-6 grid gap-5 md:grid-cols-[1fr_16rem]">
          <div class="space-y-4">
            <label class="shop-field"><span>标题</span><input v-model="bannerForm.title" class="input-field w-full" required></label>
            <label class="shop-field"><span>副标题</span><textarea v-model="bannerForm.subtitle" class="input-field min-h-20 w-full"></textarea></label>
            <label class="shop-field"><span>按钮文字</span><input v-model="bannerForm.button_text" class="input-field w-full"></label>
            <label class="shop-field"><span>关联商品</span><select v-model.number="bannerForm.product_id" class="input-field w-full"><option :value="null">不关联</option><option v-for="product in products" :key="product.id" :value="product.id">{{ product.name }}</option></select></label>
            <label class="flex items-center gap-2 text-sm font-medium text-gray-700 dark:text-gray-200"><input v-model="bannerForm.enabled" type="checkbox" class="rounded"> 展示轮播</label>
          </div>
          <aside class="rounded-xl border border-gray-200 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-800/70">
            <p class="text-sm font-semibold text-gray-900 dark:text-white">轮播图片</p>
            <div class="mt-3 overflow-hidden rounded-2xl border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-900">
              <img v-if="shopImage(bannerForm.image_url)" :src="shopImage(bannerForm.image_url)" alt="轮播预览图" class="h-36 w-full object-cover">
              <div v-else class="shop-thumb h-36 w-full" aria-hidden="true">3</div>
            </div>
            <input ref="bannerFileInput" type="file" accept="image/jpeg,image/png,image/webp,image/gif" class="hidden" @change="handleBannerImageChange">
            <button type="button" class="btn-secondary mt-3 w-full justify-center rounded-2xl px-4 py-2.5" :disabled="uploadingBannerImage" @click="bannerFileInput?.click()">
              {{ uploadingBannerImage ? '上传中...' : '本地上传图片' }}
            </button>
            <label class="shop-field mt-3"><span>或粘贴图片地址</span><input v-model="bannerForm.image_url" class="input-field w-full" placeholder="https://..."></label>
            <p class="mt-3 text-xs leading-5 text-gray-500 dark:text-gray-400">轮播图铺满卡片宽度展示，本地上传会自动等比压缩到 {{ BANNER_IMAGE_MAX_WIDTH }}px 宽以内。支持 JPG、PNG、WebP、GIF。</p>
          </aside>
        </div>
        <div class="mt-6 flex justify-end gap-3">
          <button type="button" class="btn-secondary rounded-2xl px-4 py-2.5" @click="bannerDialog.open = false">取消</button>
          <button class="btn-primary rounded-2xl px-4 py-2.5">保存</button>
        </div>
      </form>
    </div>

    <div v-if="fulfillDialog.open" class="fixed inset-0 z-50 flex items-center justify-center bg-gray-950/55 p-4 backdrop-blur-sm">
      <form class="shop-modal w-full max-w-xl rounded-xl bg-white p-5 shadow-2xl dark:bg-dark-900" @submit.prevent="confirmFulfillOrder">
        <div class="flex items-start justify-between gap-4">
          <div>
            <p class="text-sm font-semibold text-primary-600 dark:text-primary-300">商城订单</p>
            <h3 class="mt-1 text-xl font-bold text-gray-950 dark:text-white">手动发货</h3>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">用户会在商城订单里看到这段发货信息。</p>
          </div>
          <button type="button" class="rounded-full p-2 text-gray-400 hover:bg-gray-100 hover:text-gray-700 dark:hover:bg-dark-800 dark:hover:text-white" aria-label="关闭" @click="fulfillDialog.open = false">✕</button>
        </div>
        <div class="mt-5 rounded-2xl border border-gray-200 bg-gray-50 p-4 text-sm dark:border-dark-700 dark:bg-dark-800/70">
          <div class="font-semibold text-gray-950 dark:text-white">#{{ fulfillDialog.orderId }} {{ fulfillDialog.orderName }}</div>
          <div class="mt-1 text-gray-500 dark:text-gray-400">请填写交付内容、领取方式、卡密、下载链接或后续联系方式，用户会在订单里直接看到。</div>
        </div>
        <div v-if="fulfillDialog.deliveryLoading" class="mt-4 rounded-2xl border border-gray-200 bg-gray-50 p-4 text-sm text-gray-400 dark:border-dark-700 dark:bg-dark-800/70">
          正在解密用户提交的交付资料...
        </div>
        <div v-else-if="fulfillDialog.deliveryPayload" class="mt-4 rounded-2xl border border-primary-200 bg-primary-50/60 p-4 text-sm dark:border-primary-500/30 dark:bg-primary-500/10">
          <div class="flex items-center justify-between gap-3">
            <span class="font-semibold text-primary-700 dark:text-primary-300">用户提交的交付资料（已解密）</span>
            <button type="button" class="btn-secondary rounded-xl px-3 py-1 text-xs" @click="copyDeliveryPayload">复制</button>
          </div>
          <pre class="mt-2 max-h-56 overflow-auto whitespace-pre-wrap break-all font-mono text-xs text-gray-700 dark:text-gray-200">{{ fulfillDialog.deliveryPayload }}</pre>
          <p class="mt-2 text-xs text-gray-400">仅用于本次交付，请勿外传；充值完成后建议提醒用户登出所有设备。</p>
        </div>
        <label class="shop-field mt-5">
          <span>发货内容</span>
          <textarea v-model="fulfillDialog.note" class="input-field min-h-32 w-full" placeholder="例如：安装包下载地址、账号信息、服务联系说明等。" required></textarea>
        </label>
        <div class="mt-6 flex justify-end gap-3">
          <button type="button" class="btn-secondary rounded-2xl px-4 py-2.5" @click="fulfillDialog.open = false">取消</button>
          <button class="btn-primary rounded-2xl px-4 py-2.5" :disabled="fulfillingOrder">{{ fulfillingOrder ? '发送中...' : '发送给用户' }}</button>
        </div>
      </form>
    </div>
  </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import {
  adminShopAPI,
  resolveShopAssetUrl,
  type AdminShopCategory,
  type ShopAssetPurpose,
  type ShopBanner,
  type ShopCategoryPayload,
  type ShopOrder,
  type ShopProduct,
  type ShopProductPayload
} from '@/api/shop'
import { SHOP_CATEGORIES } from '@/constants/shop'
import { resizeImageForUpload } from '@/utils/imageResize'
import { useAppStore } from '@/stores'
import AppLayout from '@/components/layout/AppLayout.vue'

type TabKey = 'products' | 'categories' | 'banners' | 'orders'

const appStore = useAppStore()
const tab = ref<TabKey>('products')
const tabs: Array<{ key: TabKey; label: string }> = [
  { key: 'products', label: '商品管理' },
  { key: 'categories', label: '品类管理' },
  { key: 'banners', label: '轮播管理' },
  { key: 'orders', label: '商城订单' },
]
const products = ref<ShopProduct[]>([])
const banners = ref<ShopBanner[]>([])
const orders = ref<ShopOrder[]>([])
const adminCategories = ref<AdminShopCategory[]>([])
const fulfillingOrder = ref(false)
const savingCategory = ref(false)
const productFileInput = ref<HTMLInputElement | null>(null)
const bannerFileInput = ref<HTMLInputElement | null>(null)
const uploadingProductImage = ref(false)
const uploadingBannerImage = ref(false)

/**
 * 图片处理规格，直接对齐首页展示位：
 * - 商品图在首页只是 46px 方块（PlanCard → BrandLogo），裁成 512×512 足够 4x 屏；
 * - 轮播图铺满卡片宽度（约 540px），等比压到 1600px 宽即可。
 * 压缩后通常 30–150KB，后端上限只是为了兜住直连接口上传。
 */
const PRODUCT_IMAGE_SIZE = 512
const PRODUCT_IMAGE_MAX_BYTES = 180 * 1024
const BANNER_IMAGE_MAX_WIDTH = 1600
const BANNER_IMAGE_MAX_BYTES = 600 * 1024
/** 原图上限：只做「拒绝明显不合理」的兜底，真正的裁剪压缩交给 canvas */
const RAW_IMAGE_MAX_BYTES = 12 * 1024 * 1024

const productDialog = reactive({ open: false, id: 0 })
const bannerDialog = reactive({ open: false, id: 0 })
// gallery 显式声明为非可选，模板里可以直接按下标读写，无需到处判空
const productForm = reactive<ShopProductPayload & { gallery: string[] }>({
  name: '',
  description: '',
  image_url: '',
  product_type: 'virtual',
  price_cny_minor: 1000,
  original_price_cny_minor: 0,
  grant_usd_amount: '0',
  stock_quantity: null,
  commission_bps: 1000,
  status: 'draft',
  sort_order: 0,
  fulfillment_mode: 'manual',
  delivery_form_hint: '',
  badge_text: '',
  spec_label: '',
  highlight: false,
  category: 'other',
  gallery: [],
})
const bannerForm = reactive({
  title: '',
  subtitle: '',
  image_url: '',
  button_text: '立即查看',
  product_id: null as number | null,
  enabled: true,
  sort_order: 0,
})
const fulfillDialog = reactive({
  open: false,
  orderId: 0,
  orderName: '',
  note: '',
  deliveryPayload: '',
  deliveryLoading: false,
})
const categoryDialog = reactive({ open: false, id: 0 })
const categoryForm = reactive<ShopCategoryPayload>({
  slug: '',
  label: '',
  blurb: '',
  sort_order: 0,
  enabled: true,
})

/**
 * 商品表单的品类下拉数据源。
 * 优先用后台已配置的品类（含每类商品数），接口还没回来时退化为内置兜底，
 * 保证列表为空也能编辑已有商品、不会把品类选择框留空。
 */
const categoryOptions = computed(() => {
  if (adminCategories.value.length) {
    return adminCategories.value.map((item) => ({
      value: item.slug,
      label: item.label,
      count: item.product_count,
      enabled: item.enabled
    }))
  }
  return SHOP_CATEGORIES.map((item) => ({ value: item.value, label: item.label, count: -1, enabled: true }))
})

/** 商品表单里选中的品类若已被后台删除，补一个选项，避免 select 显示为空 */
const categorySelectOptions = computed(() => {
  const options = categoryOptions.value
  const current = productForm.category
  if (current && !options.some((item) => item.value === current)) {
    return [{ value: current, label: `${current}（已删除）`, count: -1, enabled: false }, ...options]
  }
  return options
})

function categoryLabel(slug: string) {
  return adminCategories.value.find((item) => item.slug === slug)?.label || slug
}

/** 解密并展示用户提交的交付资料（Session 或收货信息） */
async function loadOrderDelivery(orderId: number) {
  fulfillDialog.deliveryLoading = true
  try {
    const res = await adminShopAPI.getOrderDelivery(orderId)
    if (fulfillDialog.orderId === orderId) {
      fulfillDialog.deliveryPayload = (res.data as { payload?: string } | undefined)?.payload || ''
    }
  } catch {
    // 解密失败不阻塞发货流程
  } finally {
    if (fulfillDialog.orderId === orderId) {
      fulfillDialog.deliveryLoading = false
    }
  }
}

async function copyDeliveryPayload() {
  if (!fulfillDialog.deliveryPayload) return
  try {
    await navigator.clipboard.writeText(fulfillDialog.deliveryPayload)
    appStore.showToast('success', '已复制到剪贴板', 2000)
  } catch {
    appStore.showToast('error', '复制失败，请手动选择复制', 2000)
  }
}

const productPrice = computed({
  get: () => productForm.price_cny_minor / 100,
  set: (value: number) => { productForm.price_cny_minor = Math.round(Number(value || 0) * 100) }
})
const productOriginalPrice = computed({
  get: () => (productForm.original_price_cny_minor || 0) / 100,
  set: (value: number) => { productForm.original_price_cny_minor = Math.round(Number(value || 0) * 100) }
})
const productCommissionPercent = computed({
  get: () => (productForm.commission_bps || 0) / 100,
  set: (value: number) => { productForm.commission_bps = Math.round(Number(value || 0) * 100) }
})
const pendingOrdersCount = computed(() => orders.value.filter(order => order.status === 'paid' && order.fulfillment_status === 'pending').length)
const fulfilledOrdersCount = computed(() => orders.value.filter(order => order.fulfillment_status === 'fulfilled' || order.status === 'fulfilled').length)

watch(tab, () => {
  if (tab.value === 'products') void loadProducts()
  if (tab.value === 'categories') void loadCategories()
  if (tab.value === 'banners') void loadBanners()
  if (tab.value === 'orders') void loadOrders()
})

function money(minor: number) {
  return (minor / 100).toFixed(2)
}

function shopImage(url?: string | null) {
  return resolveShopAssetUrl(url)
}

function typeLabel(type: string) {
  return type === 'platform_usd_balance' ? '额度商品' : '平台商品'
}

function statusLabel(status: string) {
  return ({ draft: '草稿', published: '上架中', archived: '已归档' } as Record<string, string>)[status] || status
}

function statusClass(status: string) {
  if (status === 'published') return 'bg-emerald-50 text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-300'
  if (status === 'draft') return 'bg-amber-50 text-amber-700 dark:bg-amber-500/10 dark:text-amber-300'
  return 'bg-gray-100 text-gray-500 dark:bg-dark-800 dark:text-gray-300'
}

function orderStatusLabel(order: ShopOrder) {
  if (order.fulfillment_status === 'fulfilled' || order.status === 'fulfilled') return '已发货'
  if (order.status === 'paid' && order.fulfillment_status === 'pending') return '待发货'
  return ({ pending: '待支付', paid: '已支付', cancelled: '已取消', refunded: '已退款', failed: '失败' } as Record<string, string>)[order.status] || order.status
}

function formatDate(value: string) {
  return value ? new Date(value).toLocaleString('zh-CN') : '-'
}

async function loadProducts() {
  const res = await adminShopAPI.listProducts()
  products.value = res.data || []
}

async function loadCategories() {
  const res = await adminShopAPI.listCategories()
  adminCategories.value = res.data || []
}

async function loadBanners() {
  const res = await adminShopAPI.listBanners()
  banners.value = res.data || []
}

async function loadOrders() {
  const res = await adminShopAPI.listOrders({ page: 1, page_size: 50 })
  orders.value = res.data.items || []
}

function canFulfillOrder(order: ShopOrder) {
  return order.status === 'paid' && order.fulfillment_status === 'pending'
}

function openFulfillDialog(order: ShopOrder) {
  fulfillDialog.open = true
  fulfillDialog.orderId = order.id
  fulfillDialog.orderName = order.snapshot_name
  fulfillDialog.note = order.fulfillment_note || ''
  fulfillDialog.deliveryPayload = ''
  fulfillDialog.deliveryLoading = false
  if (order.delivery_submitted_at) {
    void loadOrderDelivery(order.id)
  }
}

async function confirmFulfillOrder() {
  if (!fulfillDialog.orderId || !fulfillDialog.note.trim()) return
  fulfillingOrder.value = true
  try {
    await adminShopAPI.fulfillOrder(fulfillDialog.orderId, { fulfillment_note: fulfillDialog.note.trim() })
    appStore.showToast('success', '订单已发货', 2500)
    fulfillDialog.open = false
    fulfillDialog.orderId = 0
    fulfillDialog.orderName = ''
    fulfillDialog.note = ''
    await loadOrders()
  } catch (error: any) {
    appStore.showToast('error', error?.message || '发货失败', 3000)
  } finally {
    fulfillingOrder.value = false
  }
}

function addGalleryItem() {
  if (!productForm.gallery) {
    productForm.gallery = []
  }
  if (productForm.gallery.length >= 6) {
    return
  }
  productForm.gallery.push('')
}

function removeGalleryItem(index: number) {
  productForm.gallery?.splice(index, 1)
}

function openProductDialog(product?: ShopProduct) {
  productDialog.open = true
  productDialog.id = product?.id || 0
  Object.assign(productForm, product ? {
    name: product.name,
    description: product.description,
    image_url: product.image_url,
    product_type: product.product_type,
    price_cny_minor: product.price_cny_minor,
    original_price_cny_minor: product.original_price_cny_minor,
    grant_usd_amount: product.grant_usd_amount,
    stock_quantity: product.stock_quantity ?? null,
    commission_bps: product.commission_bps,
    status: product.status,
    sort_order: product.sort_order,
    fulfillment_mode: product.fulfillment_mode || 'manual',
    delivery_form_hint: product.delivery_form_hint || '',
    badge_text: product.badge_text || '',
    spec_label: product.spec_label || '',
    highlight: product.highlight || false,
    category: product.category || 'other',
    gallery: [...(product.gallery || [])],
  } : {
    name: '',
    description: '',
    image_url: '',
    product_type: 'virtual',
    price_cny_minor: 1000,
    original_price_cny_minor: 0,
    grant_usd_amount: '0',
    stock_quantity: null,
    commission_bps: 1000,
    status: 'draft',
    sort_order: 0,
    fulfillment_mode: 'manual',
    delivery_form_hint: '',
    badge_text: '',
    spec_label: '',
    highlight: false,
    category: 'other',
    gallery: [],
  })
}

async function saveProduct() {
  try {
    const payload: ShopProductPayload = {
      ...productForm,
      product_type: 'virtual',
      grant_usd_amount: '0',
      // 去掉用户没填完的空输入框，避免把空串存进图廊
      gallery: (productForm.gallery || []).map((url) => url.trim()).filter(Boolean),
    }
    if (productDialog.id) await adminShopAPI.updateProduct(productDialog.id, payload)
    else await adminShopAPI.createProduct(payload)
    productDialog.open = false
    appStore.showToast('success', '商品已保存', 2500)
    await loadProducts()
  } catch (error: any) {
    appStore.showToast('error', error?.message || '保存商品失败', 3000)
  }
}

/**
 * 上传前的图片处理：把图裁到首页实际需要的尺寸再上传。
 * 后端没有图像处理能力，所以「符合展示规格」这件事在客户端一次做完。
 */
async function prepareShopImage(file: File, target: ShopAssetPurpose): Promise<File> {
  if (file.type === 'image/gif') {
    // GIF 走 canvas 会丢帧。体积本来就合规时原样上传，超出才退化为静态图
    const cap = target === 'banner' ? BANNER_IMAGE_MAX_BYTES : PRODUCT_IMAGE_MAX_BYTES
    if (file.size <= cap) return file
  }
  return resizeImageForUpload(
    file,
    target === 'product'
      ? { square: PRODUCT_IMAGE_SIZE, maxBytes: PRODUCT_IMAGE_MAX_BYTES }
      : { maxWidth: BANNER_IMAGE_MAX_WIDTH, maxBytes: BANNER_IMAGE_MAX_BYTES }
  )
}

async function uploadImage(file: File, target: ShopAssetPurpose) {
  if (!file.type.startsWith('image/')) {
    appStore.showToast('warning', '请选择图片文件', 2500)
    return
  }
  if (file.size > RAW_IMAGE_MAX_BYTES) {
    appStore.showToast('warning', '原图不能超过 12MB，请先压缩再上传', 3000)
    return
  }
  if (target === 'product') uploadingProductImage.value = true
  else uploadingBannerImage.value = true
  try {
    const prepared = await prepareShopImage(file, target)
    const res = await adminShopAPI.uploadAsset(prepared, target)
    if (target === 'product') productForm.image_url = res.data.url
    else bannerForm.image_url = res.data.url
    appStore.showToast('success', '图片已上传', 2200)
  } catch (error: any) {
    appStore.showToast('error', error?.message || '图片上传失败', 3000)
  } finally {
    if (target === 'product') uploadingProductImage.value = false
    else uploadingBannerImage.value = false
  }
}

function handleProductImageChange(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (file) void uploadImage(file, 'product')
  input.value = ''
}

function handleBannerImageChange(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (file) void uploadImage(file, 'banner')
  input.value = ''
}

async function removeProduct(product: ShopProduct) {
  if (!confirm(`确认删除商品「${product.name}」？`)) return
  await adminShopAPI.deleteProduct(product.id)
  appStore.showToast('success', '商品已删除', 2500)
  await loadProducts()
}

/** 品类标识规则与后端 shopCategorySlugPattern 保持一致 */
const CATEGORY_SLUG_PATTERN = /^[a-z][a-z0-9_]{0,31}$/

/**
 * 从商品弹窗跳到品类管理。
 * 弹窗是全屏遮罩，不关掉就看不到下面的 tab，因此这里先关闭编辑态——
 * 入口文案是「管理品类」，属于用户主动发起的跳转。
 */
function goToCategoriesTab() {
  productDialog.open = false
  tab.value = 'categories'
}

function openCategoryDialog(category?: AdminShopCategory) {
  categoryDialog.open = true
  categoryDialog.id = category?.id || 0
  Object.assign(categoryForm, category ? {
    slug: category.slug,
    label: category.label,
    blurb: category.blurb || '',
    sort_order: category.sort_order,
    enabled: category.enabled,
  } : {
    slug: '',
    label: '',
    blurb: '',
    // 默认排到最后：新品类通常是在已有品类之外补充的
    sort_order: adminCategories.value.length
      ? Math.max(...adminCategories.value.map((item) => item.sort_order)) + 10
      : 0,
    enabled: true,
  })
}

async function saveCategory() {
  const slug = categoryForm.slug.trim()
  const label = categoryForm.label.trim()
  if (!categoryDialog.id && !CATEGORY_SLUG_PATTERN.test(slug)) {
    appStore.showToast('warning', '标识需以小写字母开头，只能用小写字母、数字与下划线，最长 32 位', 3500)
    return
  }
  if (!label) {
    appStore.showToast('warning', '请填写品类名称', 2500)
    return
  }
  savingCategory.value = true
  try {
    const payload: ShopCategoryPayload = {
      slug,
      label,
      blurb: categoryForm.blurb?.trim() || '',
      sort_order: Number(categoryForm.sort_order) || 0,
      enabled: categoryForm.enabled !== false,
    }
    if (categoryDialog.id) await adminShopAPI.updateCategory(categoryDialog.id, payload)
    else await adminShopAPI.createCategory(payload)
    categoryDialog.open = false
    appStore.showToast('success', '品类已保存', 2500)
    // 品类名称会出现在商品列表的分类标签上，一起刷新
    await Promise.all([loadCategories(), loadProducts()])
  } catch (error: any) {
    appStore.showToast('error', error?.message || '保存品类失败', 3000)
  } finally {
    savingCategory.value = false
  }
}

async function removeCategory(category: AdminShopCategory) {
  if (category.product_count > 0) {
    appStore.showToast('warning', `该品类下还有 ${category.product_count} 个商品，请先改到其它品类再删除`, 3500)
    return
  }
  if (!confirm(`确认删除品类「${category.label}」？`)) return
  try {
    await adminShopAPI.deleteCategory(category.id)
    appStore.showToast('success', '品类已删除', 2500)
    await loadCategories()
  } catch (error: any) {
    appStore.showToast('error', error?.message || '删除品类失败', 3000)
  }
}

function openBannerDialog(banner?: ShopBanner) {
  bannerDialog.open = true
  bannerDialog.id = banner?.id || 0
  Object.assign(bannerForm, banner ? {
    title: banner.title,
    subtitle: banner.subtitle,
    image_url: banner.image_url,
    button_text: banner.button_text,
    product_id: banner.product_id ?? null,
    enabled: banner.enabled,
    sort_order: banner.sort_order,
  } : {
    title: '',
    subtitle: '',
    image_url: '',
    button_text: '立即查看',
    product_id: null,
    enabled: true,
    sort_order: 0,
  })
}

async function saveBanner() {
  try {
    const payload = { ...bannerForm }
    if (bannerDialog.id) await adminShopAPI.updateBanner(bannerDialog.id, payload)
    else await adminShopAPI.createBanner(payload)
    bannerDialog.open = false
    appStore.showToast('success', '轮播已保存', 2500)
    await loadBanners()
  } catch (error: any) {
    appStore.showToast('error', error?.message || '保存轮播失败', 3000)
  }
}

async function removeBanner(banner: ShopBanner) {
  if (!confirm(`确认删除轮播「${banner.title}」？`)) return
  await adminShopAPI.deleteBanner(banner.id)
  appStore.showToast('success', '轮播已删除', 2500)
  await loadBanners()
}

onMounted(async () => {
  await Promise.all([loadProducts(), loadCategories(), loadBanners(), loadOrders()])
})
</script>

<style scoped>
/* 无图占位：跟随明暗主题，避免深色模式出现刺眼亮块（与控制台商城一致） */
.shop-thumb {
  display: grid;
  place-items: center;
  background: linear-gradient(135deg, #f4f4f5 0%, #e8e8ea 100%);
  color: rgba(9, 9, 11, 0.34);
  font-size: 0.875rem;
  font-weight: 700;
}

.shop-modal {
  color: rgb(17 24 39);
}

.shop-field {
  display: block;
}

.shop-field > span {
  margin-bottom: 0.35rem;
  display: block;
  font-size: 0.875rem;
  font-weight: 700;
  color: rgb(55 65 81);
}

.shop-modal :deep(.input-field) {
  min-height: 2.75rem;
  border-color: rgb(209 213 219);
  background-color: white;
  color: rgb(17 24 39);
}

.shop-modal :deep(.input-field::placeholder) {
  color: rgb(156 163 175);
}
</style>

<style>
/* 深色覆盖必须放在**非 scoped** 块里：Vue 的 scoped-CSS 编译器会把
   `:global(.dark) X` 编译成只剩 `.dark`，X 被丢弃，导致生产构建深色规则失效。
   这里的选择器都带页面独有类名（.shop-modal / .tool-modal），不会外泄影响其它页面。 */

.dark .shop-modal {
  color: rgb(243 244 246);
}

.dark .shop-field > span {
  color: rgb(229 231 235);
}

.dark .shop-thumb {
  background: linear-gradient(135deg, #16171b 0%, #0f1013 100%);
  color: rgba(255, 255, 255, 0.34);
}

.dark .shop-modal .input-field {
  border-color: rgb(55 65 81) !important;
  background-color: rgb(17 24 39) !important;
  color: #fff !important;
}
</style>
