# SimpleVoiceChanger

Быстрый и лёгкий стрим-войсчейнжер для Windows и Android.

![GitHub stars](https://img.shields.io/github/stars/Major-Woolfi/SimpleVoiceChanger?style=social)

---

## 📖 Описание проекта

**SimpleVoiceChanger** — это приложение, перехватывающее звук с микрофона и отдающее уже обработанный голос с минимальной задержкой. Работает в фоне без прерываний.

### Возможности

- **15 аудиоэффектов** с плавной настройкой силы от 0% до 100%
- **2-слойная система обработки**: шумоподавление (AntiNoise + Noise Suppressor) → все остальные эффекты
- **Пресеты**: сохранение, загрузка, импорт и экспорт настроек в JSON
- **Тёмная и светлая темы**
- **Работа в фоне** без прерываний
- **Перекрёстная платформа**: Windows 10/11 (.exe), Android armv8 10+ (.apk)
- **Мультиязычность**: 7 языков переводов

### Эффекты

| Эффект | Описание | Слой |
|--------|----------|------|
| AntiNoise | Умное активное шумоподавление | 1 (шум) |
| Noise Suppressor | Удаление фонового шума | 1 (шум) |
| Low Cut | Удаление низких частот и гула | 2 |
| High Cut | Удаление высоких частот | 2 |
| Formant | Изменение тембра без изменения ноты | 2 |
| Compressor | Выравнивание громкости | 2 |
| Radio Effect | Звук как из радио | 2 |
| Reverb | Эхо и ощущение помещения | 2 |
| Distortion | Грязное и жёсткое звучание | 2 |
| Saturation | Плотное искажение звука | 2 |
| Harmony Engine | Дополнительные голоса поют вместе | 2 |
| Chorus | Несколько голосов одновременно | 2 |
| Flanger | Плавающее металлическое звучание | 2 |
| Phaser | Движущийся космический эффект | 2 |
| Doubler | Один голос записан дважды | 2 |

---

## 🚀 Быстрый старт

### Предварительные требования

- **Go 1.23** или новее
- **Windows 10/11** (для .exe)
- **Android SDK** (для .apk)

### Сборка (Windows)

```bat
BUILD.bat
```

Результат: файлы в `BUILD/`:
- `BUILD/SimpleVoiceChanger.exe`
- `BUILD/SimpleVoiceChanger.apk`

### Запуск

```bash
go run main.go
```

---

## 🗂 Структура проекта

```
SimpleVoiceChanger/
├── main.go                    # Точка входа
├── go.mod                     # Зависимости
├── BUILD.bat                  # Скрипт сборки
├── audio/                     # Ядро обработки звука
│   ├── engine.go              # Аудиодвижок (захват → обработка → воспроизведение)
│   ├── types.go               # Типы аудио
│   ├── buffer.go              # Кольцевой буфер
│   └── config.go              # Аудиоконфигурация
├── effects/                   # Все 15 эффектов
│   ├── effect.go              # Интерфейс эффекта
│   ├── registry.go            # Реестр эффектов
│   ├── antinoise.go           # AntiNoise (шумоподавление)
│   ├── noise_suppressor.go    # Noise Suppressor
│   ├── low_cut.go             # Low Cut
│   ├── high_cut.go            # High Cut
│   ├── formant.go             # Formant
│   ├── compressor.go          # Compressor
│   ├── radio.go               # Radio Effect
│   ├── reverb.go              # Reverb
│   ├── distortion.go          # Distortion
│   ├── saturation.go          # Saturation
│   ├── harmony.go             # Harmony Engine
│   ├── chorus.go              # Chorus
│   ├── flanger.go             # Flanger
│   ├── phaser.go              # Phaser
│   └── doubler.go             # Doubler
├── gui/                       # GUI (Fyne)
│   ├── app.go                 # Главное приложение
│   ├── mainmenu.go            # Главное меню
│   ├── effects.go             # Панель эффектов
│   ├── preset.go              # Управление пресетами
│   ├── settings.go            # Настройки приложения
│   └── theme.go               # Управление темами
├── core/                      # Основная логика
│   ├── config.go              # Конфигурация приложения
│   ├── preset.go              # Управление пресетами
│   └── voice_test.go          # Тестирование голоса
├── i18n/                      # Переводы (JSON)
│   ├── en.json                # English
│   ├── ru.json                # Русский
│   ├── uk.json                # Українська
│   ├── de.json                # Deutsch
│   ├── fr.json                # Français
│   ├── es.json                # Español
│   └── zh.json                # 中文
├── internal/
│   ├── platform/              # Платформенный код
│   └── service/               # Сервис фонового запуска
├── tests/                     # Тесты
└── BUILD/                     # Артефакты сборки
```

---

## ⚙️ Конфигурация

Настройки хранятся в `core/config.go` и сохраняются в JSON. Доступны через GUI:

- Выбор темы (тёмная/светлая)
- Частота дискретизации
- Размер буфера
- Автозапуск
- Сворачивание в трей

### Импорт/Экспорт пресетов

Пресеты сохраняются как `.json` файлы с полным набором настроек эффектов.

---

## 🏗 Архитектура

### Аудиопайплайн

```
Микрофон → RingBuffer → [Noise Layer] → [Effect Layer] → Динамик
```

**Слой 1 (Noise):** AntiNoise → Noise Suppressor
**Слой 2 (Effects):** Low Cut → High Cut → Formant → Compressor → Radio → Reverb → Distortion → Saturation → Harmony → Chorus → Flanger → Phaser → Doubler

Каждый эффект пропускается если сила = 0%.

### Платформы

- **Windows**: WASAPI через `go-wasapi`, GUI на Fyne
- **Android**: Gomobile bind, нативный Android UI с Go ядром

---

## 🌐 Локализация

Переводы в `i18n/` — отдельные `.json` файлы. Текущая поддержка:

| Язык | Код |
|------|-----|
| Русский | ru |
| English | en |
| Українська | uk |
| Deutsch | de |
| Français | fr |
| Español | es |
| 中文 | zh |

---

## 📊 Статус

| Компонент | Статус |
|-----------|--------|
| Аудиодвижок | Реализован |
| Эффекты (15/15) | Реализованы |
| GUI | Реализован |
| Пресеты | Реализованы |
| Переводы (7) | Реализованы |
| Windows сборка | Скрипт готов |
| Android сборка | Скрипт готов |
| Тесты | Базовые |

---

## 📄 Лицензия

MIT License — см. [LICENSE](LICENSE)
