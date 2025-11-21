import { createI18n } from 'vue-i18n';
import enLocale from './locales/en.json';
import zhLocale from './locales/zh.json';
import ruLocale from './locales/ru.json';
// Import Element Plus locale files
import elementEnLocale from 'element-plus/es/locale/lang/en';
import elementZhLocale from 'element-plus/es/locale/lang/zh-cn';
import elementRuLocale from 'element-plus/es/locale/lang/ru';

// Define Element Plus locale mapping
export const elementPlusLocales = {
  en: elementEnLocale,
  zh: elementZhLocale,
  ru: elementRuLocale
};

// Messages object from JSON files
const messages = {
  en: {
    ...enLocale,
  },
  zh: {
    ...zhLocale,
  },
  ru: {
    ...ruLocale,
  }
};

const i18n = createI18n({
  locale: localStorage.getItem('locale') || 'ru',
  fallbackLocale: 'en',
  messages,
  legacy: false,
  runtimeOnly: false, 
  globalInjection: true,
  silentTranslationWarn: false, 
  missingWarn: true,
  fallbackWarn: true 
});

export default i18n;
