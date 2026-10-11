# SimpleVoiceChanger

Быстрый и лёгкий стрим-войсчейнжер для Windows и Android.

## Возможности

- **15 аудиоэффектов** с плавной настройкой силы от 0% до 100%
- **2-слойная система обработки**: шумоподавление (AntiNoise + Noise Suppressor) → все остальные эффекты
- **Выбор устройств ввода/вывода** (микрофон, виртуальный кабель и т.п.)
- **Стерео-совместимость**: моно-сигнал дублируется в оба канала устройства вывода
- **Адаптивные настройки**: 1-2 столбца в зависимости от ширины окна
- **Пресеты**: сохранение, загрузка, импорт и экспорт настроек в JSON
- **Тёмная и светлая темы**
- **Мультиязычность**: 7 языков переводов (en, ru, uk, de, fr, es, zh)
- **Кроссплатформенность**: Windows 10/11 (.exe), Android armv8 10+ (.apk)

## Эффекты

| Эффект | Категория | Слой |
|--------|-----------|------|
| AntiNoise | noise | 1 (шум) |
| Noise Suppressor | noise | 1 (шум) |
| Low Cut | filter | 2 |
| High Cut | filter | 2 |
| Formant | modulation | 2 |
| Compressor | dynamics | 2 |
| Radio Effect | modulation | 2 |
| Reverb | spatial | 2 |
| Distortion | modulation | 2 |
| Saturation | modulation | 2 |
| Harmony Engine | voices | 2 |
| Chorus | voices | 2 |
| Flanger | voices | 2 |
| Phaser | voices | 2 |
| Doubler | voices | 2 |

## Статус

| Компонент | Статус |
|-----------|--------|
| Аудиодвижок | Реализован (реальный звук: Windows - PortAudio, Android - OpenSL ES) |
| Эффекты (15/15) | Реализованы и протестированы |
| GUI | Реализован (Fyne, вкладки, адаптивные настройки, выбор устройств) |
| Пресеты | Реализованы (JSON, с валидацией имён) |
| Переводы (7) | Реализованы |
| Платформенная абстракция | Реализована |
| Тесты | 22 теста, все проходят (`go test -race ./tests/`) |
| Windows сборка | Готова (`BUILD/SimpleVoiceChanger.exe` + `libportaudio.dll`) |
| Android сборка | Готова (`BUILD/SimpleVoiceChanger.apk`, 4 ABI, реальный захват/воспроизведение) |

## Быстрый старт

```bash
go run main.go
```

## Сборка

```bat
BUILD.bat
```

Результат: `BUILD/SimpleVoiceChanger.exe` (Windows), `BUILD/SimpleVoiceChanger.apk` (Android).

### Реальный звук (Windows)

Для реального захвата/воспроизведения на Windows установите `pkg-config` и PortAudio,
затем соберите с тегом:

```bash
go build -tags portaudio -o SimpleVoiceChanger.exe .
```

`BUILD.bat` включает тег автоматически, если `pkg-config --exists portaudio-2.0` успешен,
и копирует `libportaudio.dll` рядом с `.exe`. Без PortAudio собирается mock-бэкенд (без микрофона).

### Реальный звук (Android)

Android-бэкенд реализован через OpenSL ES (cgo/NDK) и включается автоматически в APK -
отдельный тег не нужен. Достаточно собрать APK через `fyne package -os android`
(шаг 8 в `BUILD.bat`) при наличии `ANDROID_HOME`/`ANDROID_NDK_HOME`.

## Тесты

```bash
go test -v ./tests/
```

## Лицензия

MIT License.