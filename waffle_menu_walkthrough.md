# آلية عمل قائمة التطبيقات (Waffle Menu)

يشرح هذا المستند كيفية عمل زر "قائمة التطبيقات" (Waffle Menu) الموجود في أعلى يسار الشاشة (أو اليمين في التخطيط العربي) في ملف `DashboardLayout.vue`.

## 1. هيكل الزر (The Button)

الزر عبارة عن أيقونة مكونة من 9 نقاط مرتبة في شبكة 3x3، ويتم إنشاؤه باستخدام HTML و CSS بسيط:

```html
<button
  ref="triggerRef"
  @click.stop="togglePopover"
  class="app-launcher-btn"
  :class="{ 'active': showPopover }"
  title="القائمة"
>
  <div class="grid grid-cols-3 gap-1">
    <span v-for="i in 9" :key="i" class="dot"></span>
  </div>
</button>
```

- **التصميم (CSS)**:
  - الفئة `.app-launcher-btn`: تحدد شكل الزر الدائري.
  - الفئة `.dot`: ترسم النقاط الصغيرة داخل الشبكة.
  - عند التحويم (Hover) أو التفعيل (`active`)، يتغير لون الخلفية والنقاط.

## 2. القائمة المنبثقة (The Popover)

القائمة تظهر عند الضغط على الزر، وهي محاطة بـ `<Transition>` لإضافة تأثيرات حركية عند الظهور والاختفاء.

```html
<Transition name="popover">
  <div v-if="showPopover" ref="popoverRef" class="popover-menu" @click.stop>
    <!-- رأس القائمة (العنوان وزر العودة) -->
    <div class="popover-header">...</div>

    <!-- شبكة القائمة الرئيسية -->
    <div v-if="currentMenu === 'main'" class="menu-grid">...</div>

    <!-- شبكة القوائم الفرعية -->
    <div v-else class="menu-grid submenu">...</div>
  </div>
</Transition>
```

## 3. المنطق البرمجي (Script Logic)

يعتمد عمل القائمة على عدة متغيرات ومتغيرات محسوبة (Computed Properties) في Vue.js.

### أ- حالة العرض (State)
```javascript
const showPopover = ref(false)      // للتحكم في ظهور القائمة
const currentMenu = ref('main')     // لتحديد المستوى الحالي (الرئيسية أو فرعية)
```

### ب- تعريف العناصر (Menu Configuration)
يتم تعريف العناصر في `menuItems`، حيث يحتوي كل عنصر على:
- `title`: العنوان.
- `icon`: الأيقونة (إيموجي).
- `path`: المسار (إذا كان رابطاً مباشراً).
- `children`: قائمة فرعية (إذا كان قسماً رئيسياً).
- `roles`: مصفوفة الصلاحيات (من يمكنه رؤية هذا العنصر).
- `color`: لون الخلفية للأيقونة.

### ج- التصفية حسب الصلاحيات
يتم تصفية القائمة تلقائياً بناءً على دور المستخدم المسجل:
```javascript
return items.filter(item => item.roles.includes(authStore.userRole))
```

### د- التنقل (Navigation Handling)
عند النقر على عنصر (`handleItemClick`):
1. **إذا كان رابطاً (`path`)**:
   - يتم التوجيه باستخدام `router.push`.
   - تغلق القائمة (`showPopover = false`).
   - تعود القائمة للحالة الرئيسية (`currentMenu = 'main'`).
2. **إذا كان له أبناء (`children`)**:
   - يتغير `currentMenu` إلى عنوان العنصر (مثلاً "الإدارة").
   - تتغير القائمة لتعرض العناصر الفرعية لهذا القسم.

### هـ- العودة للخلف
زر العودة يظهر فقط داخل القوائم الفرعية، ووظيفته إعادة `currentMenu` إلى `'main'`.

### و- الإغلاق (Closing Behavior)
يتم إغلاق القائمة في الحالات التالية:
1. النقر على زر القائمة مرة أخرى.
2. النقر خارج القائمة (`handleClickOutside`).
3. ضغط زر `Escape`.
4. اختيار رابط نهائي.

## 4. التصميم والتنسيق (Styling)

- **Menu Grid**: تستخدم `display: grid` لترتيب العناصر في 3 أعمدة للقائمة الرئيسية، وعمودين للقوائم الفرعية.
- **Micro-interactions**: تأثيرات تكبير (Scale) بسيطة عند تحويم الماوس فوق الأيقونات.
- **Scrollbar**: شريط تمرير مخصص للقوائم الطويلة.

## ملخص التدفق (Flow Summary)
1. يضغط المستخدم على زر الـ Waffle.
2. تظهر القائمة الرئيسية (`main`) بتأثير حركي.
3. يختار المستخدم "الإدارة" (مثلاً).
4. تختفي عناصر القائمة الرئيسية وتظهر عناصر الإدارة (بفضل `v-if/v-else`).
5. يختار المستخدم "المستخدمين".
6. يتم توجيه المتصفح إلى `/admin/users` وتختفي القائمة تماماً.
