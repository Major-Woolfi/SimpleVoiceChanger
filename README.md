# SimpleVoiceChanger

Быстрый и лёгкий стрим-войсчейнжер для Windows и Android.

## Возможности

- **15 аудиоэффектов** с плавной настройкой силы от 0% до 100%
- **2-слойная система обработки**: шумоподавление (AntiNoise + Noise Suppressor) → все остальные эффекты
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
| Аудиодвижок | Реализован (mock-захват/воспроизведение) |
| Эффекты (15/15) | Реализованы и протестированы |
| GUI | Реализован (Fyne, вкладки) |
| Пресеты | Реализованы (JSON) |
| Переводы (7) | Реализованы |
| Платформенная абстракция | Реализована |
| Тесты | 11 тестов, все проходят |
| Windows сборка | Скрипт готов |

## Быстрый старт

```bash
go run main.go
```

## Сборка

```bat
BUILD.bat
```

Результат: `BUILD/SimpleVoiceChanger.exe`

## Тесты

```bash
go test -v ./tests/
```

## Лицензия

MIT License.