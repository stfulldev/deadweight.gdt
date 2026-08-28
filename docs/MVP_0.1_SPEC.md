# deadweight.gdt MVP 0.1 — техническая и продуктовая спецификация

> Статус: implementation-ready draft  
> Целевая версия: `0.1.0`  
> Основной язык: Go  
> Интерфейс: standalone CLI  
> Целевой формат: Godot 4.x text scenes (`.tscn`, `format=3`)  
> Зависимость от Godot: отсутствует

## 1. Краткое определение продукта

**deadweight.gdt** — standalone Go CLI для статического анализа сложности сцен Godot 4.x. Он читает текстовые `.tscn`, рекурсивно раскрывает вложенные `.tscn`-сцены, считает структурные и рендеринговые метрики и сравнивает их с budget-профилем.

Основной первый опыт:

```bash
deadweight.gdt inspect res://levels/city.tscn
deadweight.gdt check res://levels/city.tscn --preset steam-deck
```

Инструмент не требует установленного Godot, не запускает проект и не пытается предсказывать FPS. Он даёт быстрый статический guardrail для локальной разработки и CI.

Ключевая продуктовая формула:

```text
Godot 4 .tscn
      ↓
subset parser
      ↓
recursive scene graph
      ↓
effective metrics
      ↓
heuristic/custom budget
      ↓
human-readable result + exit code
```

## 2. Product vision

deadweight.gdt должен отвечать на вопрос:

> «Не стала ли эта сцена структурно заметно тяжелее выбранного нами budget-профиля?»

Он должен быть полезен в трёх режимах:

1. **Исследование сцены.** Разработчик быстро получает понятную сводку без запуска Godot.
2. **Project guardrail.** Команда хранит свои лимиты в `.deadweight.gdt.json` и ловит случайный рост сцены.
3. **Первый ориентир.** Новый пользователь выбирает встроенный эвристический preset `mobile`, `steam-deck` или `desktop`, а затем адаптирует его к своей игре.

deadweight.gdt позиционируется как статический анализатор scene complexity, а не profiler. Его ценность — скорость, воспроизводимость, один переносимый бинарник и честное обозначение границ анализа.

## 3. Принципы продукта

1. **Go-first и standalone.** Основной путь не требует Godot, GDScript, editor plugin или import pipeline.
2. **Рекурсивность обязательна.** Метрики одной `.tscn` без раскрытия вложенных сцен могут быть обманчивыми.
3. **Никакой молчаливой неточности.** Неразрешённые инстансы и неподдерживаемое наследование переводят анализ в `PARTIAL`.
4. **Preset — ориентир, не обещание.** Встроенные значения всегда помечаются как `heuristic` и `experimental`.
5. **Детерминированность.** Один набор файлов и конфигурация должны давать одинаковый результат и одинаковый порядок вывода.
6. **Маленький поддерживаемый parser.** Реализуется документированный subset TSCN, нужный текущим метрикам; полный Godot parser не является целью.
7. **Без скрытой магии.** Все определения метрик, правила merge конфигурации и причины partial-анализа документируются.
8. **Без преждевременного расширения scope.** `.tres`, импортированные 3D-сцены, runtime profiling и GUI остаются за пределами 0.1.

## 4. Цели MVP 0.1

Версия 0.1 должна:

- анализировать Godot 4.x `.tscn` с заголовком `format=3`;
- находить корень проекта по `project.godot`;
- принимать filesystem-пути и `res://`;
- корректно разрешать `res://` и относительные resource paths;
- разбирать необходимый subset TSCN без regexp-only подхода;
- находить `ExtResource(...)`, используемые как `PackedScene` instances;
- рекурсивно раскрывать вложенные `.tscn`;
- правильно учитывать одну сцену, инстанцированную много раз;
- строить dependency graph и обнаруживать циклы;
- использовать in-memory parse/summary cache;
- считать восемь зафиксированных метрик;
- различать `COMPLETE` и `PARTIAL`;
- группировать unresolved instances/resources и показывать multiplicity;
- проверять метрики по встроенному preset, custom profile и overrides;
- поддерживать JSON-конфиг `.deadweight.gdt.json` версии 1;
- предоставлять команды `inspect`, `check`, `presets`, `presets show`;
- возвращать документированные exit codes;
- иметь достаточные unit, integration, golden и testdata tests;
- работать без запуска или наличия Godot.

## 5. Нецели MVP 0.1

В 0.1 намеренно не входят:

- предсказание FPS, frametime или фактической производительности;
- заявления о соответствии официальным требованиям Steam Deck или других устройств;
- запуск Godot headless/editor/import pipeline;
- анализ runtime-generated nodes и сцен, загружаемых из скриптов;
- полный parser всех Variant-типов Godot;
- deep parsing `.tres`;
- раскрытие `.scn`, `.glb`, `.gltf`, `.blend` и других импортируемых сцен;
- полноценный merge inherited scenes и их property overrides;
- анализ GDScript/C# inheritance и пользовательских типов узлов;
- vertices, triangles, draw calls, materials, shaders, textures, VRAM и audio memory;
- 2D light metrics;
- GUI, web UI или Godot editor plugin;
- JSON/SARIF/JUnit/HTML reports;
- GitHub Action и PR comments;
- persistent/on-disk cache;
- watch mode;
- анализ целой директории или всех сцен проекта одной командой;
- автоматическое определение целевого hardware или renderer из `project.godot`.

Эти пункты нельзя «заодно» добавлять при реализации 0.1 без отдельного изменения спецификации.

## 6. Целевые пользовательские сценарии

### 6.1. Быстрая инспекция

```bash
deadweight.gdt inspect scenes/city.tscn
```

Пользователь получает effective metrics с учётом вложенных `.tscn`, статус достоверности и предупреждения.

### 6.2. Проверка по встроенному preset

```bash
deadweight.gdt check res://levels/city.tscn --preset steam-deck
```

Пользователь получает таблицу `Actual / Budget / Result`. Превышение любого настроенного лимита возвращает ненулевой exit code.

### 6.3. Проверка по project profile

```bash
deadweight.gdt check res://levels/city.tscn --profile shipping
```

`shipping` берётся из `${project_root}/.deadweight.gdt.json`, может наследовать встроенный preset или другой custom profile.

### 6.4. Локальный разовый override

```bash
deadweight.gdt check city.tscn \
  --preset steam-deck \
  --budget mesh_instances=1600 \
  --budget shadow_lights=6
```

CLI overrides имеют максимальный приоритет и не изменяют файл конфигурации.

### 6.5. Консервативный CI

```bash
deadweight.gdt check res://levels/city.tscn \
  --profile shipping \
  --fail-on-partial
```

Если хотя бы часть effective scene не может быть статически раскрыта, команда завершится exit code `3`, даже если известные значения не превысили budget.

## 7. Термины

| Термин | Определение |
|---|---|
| Root scene | `.tscn`, переданная пользователем в `inspect` или `check` |
| Project root | Ближайшая подходящая директория с `project.godot` |
| Local node | `[node ...]`, физически объявленный в анализируемом файле |
| Ordinary node | Local node с `type=...`, не являющийся scene instance |
| Scene instance | `[node ... instance=ExtResource(...)]` или `instance_placeholder`, кроме inherited root |
| Inherited root | Первый/root `[node]` сцены, сам использующий `instance=ExtResource(...)` как базовую сцену |
| Resolved scene | Существующая и успешно разобранная `.tscn` `format=3` |
| Unresolved instance | Scene instance, effective tree которого нельзя раскрыть в 0.1 |
| Dependency | Разрешённая `.tscn`, достижимая из root scene через instance/inheritance edges |
| Effective metrics | Метрики после рекурсивного раскрытия каждого occurrence scene instance |
| Preset | Встроенный неизменяемый базовый профиль (`mobile`, `steam-deck`, `desktop`) |
| Profile | Пользовательский именованный профиль из `.deadweight.gdt.json` |
| Override | Более приоритетные значения budgets поверх preset/profile |
| Complete | Все данные, необходимые для метрик 0.1, разрешены и поддержаны |
| Partial | Есть unresolved/unsupported элементы, способные повлиять на результат |

## 8. Совместимость и поддерживаемый вход

### 8.1. Поддерживается

- файлы с расширением `.tscn`;
- заголовок `[gd_scene ... format=3 ...]`;
- Godot 4 string resource IDs, например `id="2_abcd"`;
- numeric IDs, если они встречаются в старых/ручных `format=3` файлах;
- `ExtResource("id")` и `ExtResource(id)`;
- `res://path/to/file`;
- resource paths относительно директории текущей `.tscn`;
- отсутствующий или deprecated `load_steps` — значение игнорируется;
- `uid` в заголовках — сохраняется как metadata, но не используется для разрешения пути;
- комментарии `; ...` вне строк;
- неизвестные properties и сложные Variant values — безопасно пропускаются.

### 8.2. Не поддерживается как root input

- Godot 3.x `format=2`;
- binary `.scn`;
- `.escn`;
- `.tres`;
- `.glb`, `.gltf`, `.blend`;
- `uid://...` без обычного filesystem `path`;
- `user://...`;
- PackedScene, хранящийся только как `SubResource(...)`.

Неподдерживаемый root input — fatal error. Неподдерживаемый nested PackedScene — причина `PARTIAL`, а не fatal error.

### 8.3. Версионное обещание

deadweight.gdt 0.1 обещает поддержку **описанного subset `format=3`**, а не каждого настоящего и будущего синтаксического расширения Godot 4.x. Testdata должны включать файлы, сохранённые несколькими версиями Godot 4, но совместимость определяется тестами subset parser, а не наличием Godot в CI.

## 9. Высокоуровневая архитектура

```text
CLI args
  │
  ├─ project root finder + path resolver
  ├─ config loader + preset/profile resolver
  │
  └─ scene analyzer
       ├─ TSCN subset parser
       ├─ dependency graph / DFS / cycle detector
       ├─ parse cache
       ├─ expanded-summary cache
       └─ metrics aggregator
              │
              ├─ inspect report
              └─ budget checker → check report → exit code
```

### 9.1. Слои

1. **Input layer:** CLI, config loading, root discovery.
2. **Syntax layer:** streaming lexer/parser, TSCN AST subset.
3. **Resolution layer:** canonical paths, ExtResource lookup, dependency edges.
4. **Analysis layer:** recursion, cycle detection, cache, aggregation, completeness.
5. **Policy layer:** presets, profiles, overrides, budget evaluation.
6. **Presentation layer:** deterministic console report and process exit code.

### 9.2. Направление зависимостей пакетов

```text
cmd/deadweight.gdt
  → internal/cli
      → internal/app
      → internal/report

internal/app
  → internal/config
  → internal/project
  → internal/scene
  → internal/preset
  → internal/budget

internal/scene
  → internal/tscn
  → internal/project
  → internal/metrics
  → internal/diagnostic

internal/budget
  → internal/metrics
  → internal/preset
```

Циклические package imports запрещены. Пакет `tscn` ничего не знает о budget или console output. Пакет `budget` ничего не знает о TSCN.

## 10. Предлагаемая структура репозитория

```text
deadweight.gdt/
├── cmd/
│   └── deadweight.gdt/
│       └── main.go
├── internal/
│   ├── app/
│   │   ├── app.go
│   │   └── result.go
│   ├── budget/
│   │   ├── checker.go
│   │   ├── checker_test.go
│   │   └── model.go
│   ├── cli/
│   │   ├── check.go
│   │   ├── inspect.go
│   │   ├── presets.go
│   │   └── root.go
│   ├── config/
│   │   ├── load.go
│   │   ├── model.go
│   │   ├── resolve.go
│   │   ├── validate.go
│   │   └── config_test.go
│   ├── diagnostic/
│   │   └── diagnostic.go
│   ├── metrics/
│   │   ├── metrics.go
│   │   └── names.go
│   ├── preset/
│   │   ├── data/
│   │   │   ├── desktop.json
│   │   │   ├── mobile.json
│   │   │   └── steam-deck.json
│   │   ├── builtin.go
│   │   ├── builtin_test.go
│   │   └── model.go
│   ├── project/
│   │   ├── finder.go
│   │   ├── finder_test.go
│   │   ├── paths.go
│   │   └── paths_test.go
│   ├── report/
│   │   ├── check.go
│   │   ├── inspect.go
│   │   ├── presets.go
│   │   └── golden_test.go
│   ├── scene/
│   │   ├── aggregate.go
│   │   ├── analyzer.go
│   │   ├── cache.go
│   │   ├── graph.go
│   │   ├── graph_test.go
│   │   ├── model.go
│   │   ├── resolver.go
│   │   └── analyzer_test.go
│   └── tscn/
│       ├── ast.go
│       ├── lexer.go
│       ├── parser.go
│       ├── parser_test.go
│       ├── position.go
│       └── token.go
├── schema/
│   └── deadweight.gdt.schema.json
├── testdata/
│   ├── projects/
│   │   ├── complete/
│   │   ├── cyclic/
│   │   ├── inherited/
│   │   ├── malformed/
│   │   ├── repeated/
│   │   ├── relative-paths/
│   │   └── unresolved/
│   └── golden/
├── .github/
│   └── workflows/
│       └── ci.yml
├── .gitignore
├── .golangci.yml                  # optional; не блокирует первый milestone
├── CHANGELOG.md
├── LICENSE                        # рекомендуемая лицензия: MIT
├── README.md
├── go.mod
└── go.sum
```

Go module path:

```text
github.com/stfulldev/deadweight.gdt
```

CLI можно реализовать на `github.com/spf13/cobra`. Для JSON-конфига используются `encoding/json` и строгая валидация; Viper не нужен.

## 11. Основные структуры данных

Имена могут слегка измениться при реализации, но семантика должна сохраниться.

```go
type Metrics struct {
    Nodes             int64
    TreeDepth         int64
    SceneInstances    int64
    MeshInstances     int64
    Lights            int64
    ShadowLights      int64
    ExternalResources int64
    SceneDependencies int64
}

type AnalysisStatus string

const (
    AnalysisComplete AnalysisStatus = "complete"
    AnalysisPartial  AnalysisStatus = "partial"
)

type Reliability string

const (
    ReliabilityExact       Reliability = "exact"
    ReliabilityLowerBound  Reliability = "lower_bound"
    ReliabilityApproximate Reliability = "approximate"
)

type Analysis struct {
    Root          string
    ProjectRoot   string
    Status        AnalysisStatus
    Reliability   Reliability
    Metrics       Metrics
    Coverage      Coverage
    Graph         DependencyGraph
    Diagnostics   []diagnostic.Diagnostic
}

type Coverage struct {
    ResolvedSceneInstances   int64
    UnresolvedSceneInstances int64
    ParsedSceneFiles         int64
    InheritedScenes          int64
}
```

Все счётчики — `int64`. Любое переполнение при сложении или умножении является fatal internal analysis error, а не wraparound.

### 11.1. Диагностика

```go
type Diagnostic struct {
    Code       string
    Severity   Severity // warning | error
    Message    string
    File       string
    Line       int
    Column     int
    Resource   string
    Occurrences int64
}
```

Коды стабильны в пределах minor-линии и пригодны для tests. Минимальный набор:

| Code | Значение |
|---|---|
| `SB1001` | unresolved `.tscn` scene instance |
| `SB1002` | imported/binary PackedScene cannot be expanded |
| `SB1003` | inherited scene detected; override merge unsupported |
| `SB1004` | external resource path is missing/unreadable/outside project |
| `SB1005` | `instance_placeholder` cannot be expanded |
| `SB1006` | UID-only or `user://` resource path unsupported |
| `SB1007` | custom node type cannot be classified by base class |
| `SB2001` | invalid or unsupported TSCN root |
| `SB2002` | scene dependency cycle |
| `SB2003` | invalid configuration/profile cycle |
| `SB2004` | arithmetic overflow |

`SB1007` в 0.1 необязательно выводить для каждого custom type: иначе обычные script-driven nodes создадут шум. Он используется только когда parser видит ситуацию, явно влияющую на заявленную классификацию, и по умолчанию не обязан переводить анализ в partial. Ограничение всё равно документируется в README.

## 12. TSCN subset parser

### 12.1. Почему не regexp

TSCN содержит строки с escape-последовательностями, вложенные массивы и словари, вызовы Variant constructors (`Transform3D(...)`, `PackedByteArray(...)`, `ExtResource(...)`) и многострочные values. Regex по строкам быстро начинает ошибочно воспринимать `[` или `;` внутри значения как новый section/comment.

В 0.1 нужен маленький streaming lexer и section parser, а не полный AST всех Variant types.

### 12.2. Что parser обязан распознавать

Sections:

```text
[gd_scene ...]
[ext_resource ...]
[sub_resource ...]
[node ...]
[connection ...]
```

Из `[gd_scene]`:

- `format` — обязательно и равно `3`;
- `uid` — optional metadata;
- `load_steps` — игнорируется.

Из `[ext_resource]`:

- `id`;
- `type`;
- `path`;
- `uid`.

Из `[node]`:

- `name`;
- `type`;
- `parent`;
- `instance`;
- `instance_placeholder`;
- `owner`;
- `index`;
- source position.

Из body properties node:

- `shadow_enabled = true|false`.

Остальные header attributes сохранять необязательно. Остальные body properties должны безопасно пропускаться без materialization больших arrays.

### 12.3. Минимальная AST-модель

```go
type Document struct {
    Header       SceneHeader
    ExtResources map[string]ExtResource
    Nodes        []Node
    Features     Features
}

type SceneHeader struct {
    Format int
    UID    string
}

type ExtResource struct {
    ID       string
    Type     string
    UID      string
    Path     string
    Position Position
}

type ResourceRef struct {
    Kind string // ExtResource | SubResource
    ID   string
}

type Node struct {
    Name                string
    Type                string
    Parent              string
    Owner               string
    Index               *int
    Instance            *ResourceRef
    InstancePlaceholder string
    ShadowEnabled       *bool
    Position            Position
}

type Features struct {
    HasInheritedRoot bool
    HasOverrideNodes bool
    HasEditable      bool
}
```

`ShadowEnabled == nil` означает, что property отсутствует. Для поддерживаемых 3D lights это интерпретируется как Godot default `false`, а не как unknown.

### 12.4. Lexer contract

Lexer работает с `io.Reader` и выдаёт tokens с line/column. Минимальные token classes:

```text
identifier
string
integer
true / false
[
]
(
)
{
}
=
,
newline
EOF
other
```

Требования:

- корректно обрабатывать `\"`, `\\`, `\n`, `\r`, `\t` внутри строк;
- `;` начинает comment только вне строки и до конца физической строки;
- whitespace незначим вне строк;
- unknown bytes/tokens не должны приводить к panic;
- большие unknown values, включая `PackedByteArray`, пропускаются streaming-образом;
- parser должен балансировать `()`, `[]`, `{}` при пропуске value;
- newline завершает property value только при нулевой глубине вложенности и вне строки;
- section header распознаётся только в позиции нового top-level entry, а не внутри property value.

Не следует использовать `bufio.Scanner` с дефолтным token limit для всего файла. Предпочтителен `bufio.Reader` или собственный buffered reader, чтобы большие сериализованные subresources не ломали анализ.

### 12.5. Parser contract

1. Первый meaningful section обязан быть `[gd_scene]`.
2. `format` обязан существовать и быть `3`.
3. Duplicate `ext_resource.id` — syntax/semantic error.
4. Первый `[node]` считается root; у обычной сцены у него нет `parent`.
5. Все последующие nodes должны иметь `parent`, кроме unsupported constructs, которые диагностируются.
6. `instance=ExtResource("id")` должен быть преобразован в `ResourceRef`.
7. `instance=SubResource(...)` не раскрывается и делает соответствующий occurrence partial.
8. Node без `type` и без `instance` считается override stub, не новым узлом; наличие такого node отмечается в `Features`.
9. Неизвестный section безопасно пропускается. Если это `[editable ...]`, устанавливается `HasEditable` и inherited/edited semantics отмечаются как partial.
10. Ошибка синтаксиса root scene fatal. Ошибка синтаксиса разрешённой nested `.tscn` также fatal: файл заявлен как поддерживаемый, поэтому молчаливый partial может скрыть corruption.

### 12.6. Поддерживаемый пример

```text
[gd_scene load_steps=2 format=3 uid="uid://city"]

[ext_resource type="PackedScene" path="res://props/lamp.tscn" id="1_lamp"]

[node name="City" type="Node3D"]

[node name="Sun" type="DirectionalLight3D" parent="."]
shadow_enabled = true

[node name="LampA" parent="." instance=ExtResource("1_lamp")]
[node name="LampB" parent="." instance=ExtResource("1_lamp")]
```

Lamp scene парсится один раз, но её effective summary применяется два раза.

## 13. Поиск project root и разрешение путей

### 13.1. Root discovery

`project.godot` определяет корень Godot-проекта. Алгоритм зависит от входа.

#### Filesystem input

Для `scenes/city.tscn` или абсолютного пути:

1. Преобразовать input в absolute path относительно current working directory.
2. Проверить, что root file существует и имеет `.tscn` extension.
3. Начать с директории root scene.
4. Идти вверх до filesystem root.
5. Выбрать **ближайшую** директорию, содержащую regular file `project.godot`.
6. Если не найдено — fatal error с подсказкой `--project`.

#### `res://` input

Для `res://levels/city.tscn` невозможно сначала найти scene path без project root. Порядок:

1. Если передан `--project`, использовать его.
2. Иначе искать ближайший `project.godot`, начиная с current working directory и поднимаясь вверх.
3. Разрешить `res://...` относительно найденного root.
4. Если project root не найден — fatal error.

#### Explicit project

`--project PATH` принимает директорию проекта или путь к `project.godot`. Директория обязана содержать `project.godot`. Explicit project имеет приоритет над auto-discovery.

### 13.2. PathResolver API

```go
type Resolver struct {
    ProjectRoot string // canonical absolute path
}

func (r Resolver) ResolveSceneInput(input, cwd string) (ResolvedPath, error)
func (r Resolver) ResolveResource(fromScene, raw string) Resolution
func (r Resolver) DisplayPath(abs string) string
```

`ResolvedPath` должен хранить:

- canonical absolute filesystem path для I/O/cache keys;
- normalized `res://...` display path для reports;
- original raw value для diagnostics.

### 13.3. Resource resolution rules

Для `ResolveResource(fromScene, raw)`:

1. `res://foo/bar` → `${projectRoot}/foo/bar`.
2. `../foo/bar` или `foo/bar` → `Join(Dir(fromScene), raw)`.
3. Absolute host path разрешается только если canonical target находится внутри project root; иначе partial warning.
4. `uid://...` без usable filesystem path → unsupported, partial.
5. `user://...` → unsupported, partial.
6. Empty path → unresolved, partial.
7. После clean/canonicalization target обязан оставаться внутри canonical project root.
8. Existing symlinks разрешаются через `EvalSymlinks`; symlink escape за project root запрещён.
9. Никакого case-insensitive fallback: неправильный регистр пути должен обнаруживаться на case-sensitive systems.
10. Display paths всегда используют `/`, даже на Windows.

Проверка «внутри project root» должна быть segment-aware (`filepath.Rel`), а не через небезопасный string prefix.

### 13.4. Config discovery

После нахождения project root:

1. `--config PATH`, если указан;
2. иначе `${projectRoot}/.deadweight.gdt.json`, если существует;
3. иначе конфигурация отсутствует.

Явно указанный, но отсутствующий `--config` — fatal error. Не найденный implicit config — нормальный случай.

## 14. Recursive nested scenes

### 14.1. Классификация instance target

Для каждого local node с `instance=ExtResource(...)`:

| Target | Поведение 0.1 |
|---|---|
| Existing `.tscn`, `format=3` | Рекурсивно разобрать и раскрыть |
| Missing `.tscn` | Count occurrence как unresolved, `PARTIAL` |
| `.glb`, `.gltf`, `.blend`, `.scn` | Imported/binary scene, не раскрывать, `PARTIAL` |
| `SubResource(...)` | Не раскрывать, `PARTIAL` |
| `instance_placeholder` | Не раскрывать, `PARTIAL` |
| Wrong/missing ExtResource ID | Unresolved, `PARTIAL` |
| ExtResource с несовместимым `type` | Diagnostic; если extension `.tscn`, попытаться как scene, иначе partial |

Для unresolved instance deadweight.gdt знает как минимум о корневом occurrence. Поэтому:

- `scene_instances += 1`;
- `nodes += 1` как известная нижняя граница root node инстанса;
- `tree_depth` учитывает mount depth;
- внутренние nodes/render metrics/resources неизвестны;
- `UnresolvedSceneInstances += 1`.

### 14.2. Повторные instances

Если `StreetLamp.tscn` имеет:

```text
nodes:           8
mesh_instances: 1
lights:          1
shadow_lights:   1
```

и встречается 100 раз, вклад в occurrence-based metrics:

```text
nodes:             800
mesh_instances:    100
lights:            100
shadow_lights:     100
scene_instances:   100 + nested instances inside each lamp
```

Но:

```text
scene_dependencies: 1       # StreetLamp.tscn уникальна
external_resources: unique union, а не ×100
```

### 14.3. Почему instance node не считается дважды

Node header с `instance=...` представляет root инстанцированной сцены. При resolved child:

```text
contribution.nodes = childSummary.nodes
```

а не:

```text
1 + childSummary.nodes
```

Иначе root child scene был бы посчитан дважды. Для unresolved child используется `1`, потому что известен только occurrence root.

## 15. Dependency graph

### 15.1. Модель

```go
type EdgeKind string

const (
    EdgeInstance    EdgeKind = "instance"
    EdgeInheritance EdgeKind = "inheritance"
)

type GraphNode struct {
    Path string // canonical absolute cache key
}

type GraphEdge struct {
    From        string
    To          string // empty when unresolved
    RawTarget   string
    ResourceID  string
    Kind        EdgeKind
    Occurrences int64
    Resolved    bool
}
```

Пара одинаковых resolved edges может быть сжата с `Occurrences > 1`. Для агрегации occurrence semantics сохраняется отдельно; graph нужен для объяснения и unique closure.

### 15.2. Cycle detection

Используется DFS с состояниями:

```text
unvisited → visiting → visited
```

При edge в `visiting` строится cycle path из recursion stack:

```text
ERROR SB2002: scene dependency cycle

res://A.tscn
→ res://B.tscn
→ res://C.tscn
→ res://A.tscn
```

Цикл в resolved `.tscn` graph — fatal analysis error, exit code `2`. Анализ не выдаёт budget verdict по усечённому циклу.

Самоссылка `A → A` обрабатывается тем же механизмом.

### 15.3. `scene_dependencies`

Это число **уникальных успешно разрешённых `.tscn` files**, достижимых из root через instance или inheritance edges, без root scene.

- Одна сцена, использованная 100 раз, добавляет `1`.
- Dependency dependency также входит.
- Inherited base scene входит.
- Missing `.tscn` не входит в число resolved dependencies и показывается отдельно как unresolved.
- При fatal cycle metric не публикуется как валидный verdict.

## 16. Caching

В 0.1 cache только in-memory и живёт один CLI invocation.

### 16.1. Parse cache

```go
map[CanonicalPath]*tscn.Document
```

Каждый resolved `.tscn` физически читается и парсится максимум один раз.

### 16.2. Expanded summary cache

```go
map[CanonicalPath]*SceneSummary
```

После успешного DFS summary одной сцены сохраняется и переиспользуется для каждого occurrence. Summary описывает **один instance данной сцены**.

### 16.3. Что входит в summary

```go
type SceneSummary struct {
    MetricsWithoutUniqueCounts Metrics
    Depth                      int64
    ExternalResourceKeys       set[string]
    DependencyPaths            set[string]
    Coverage                   Coverage
    Reliability                Reliability
}
```

Практически лучше не дублировать `Depth` отдельно от `Metrics.TreeDepth`; структура выше объясняет смысл.

При применении summary `N` раз:

- occurrence counters умножаются на `N` с overflow check;
- `tree_depth` не умножается;
- unique sets объединяются;
- diagnostics группируются по resource/path и получают occurrence count;
- parsed file count берётся из глобального cache size, а не умножается.

### 16.4. Cache invalidation

Не нужна: invocation read-only и короткоживущий. Если файл изменился во время анализа, поведение unspecified; watch/incremental mode — roadmap.

## 17. Построение local tree и `tree_depth`

Root depth определяется как `1`.

Для обычного parent path:

```text
root node                      depth 1
parent="."                     depth 2
parent="Arm"                   depth 3
parent="Arm/Hand"              depth 4
```

Mount depth scene instance вычисляется по тем же правилам.

Если resolved child имеет `tree_depth = C`, а его root смонтирован на depth `M`, абсолютная глубина его самого глубокого node:

```text
M + C - 1
```

Итог:

```text
tree_depth = max(
    depth каждого local ordinary node,
    mountDepth + child.tree_depth - 1 для каждого resolved instance,
    mountDepth для каждого unresolved instance
)
```

`parent` с `..`, абсолютным NodePath или синтаксисом, не соответствующим сериализованному scene tree path, вызывает diagnostic и partial, а не silent guess.

## 18. Inherited scenes

### 18.1. Detection

Минимальный сигнал inherited scene — root/первый `[node]` содержит `instance=ExtResource(...)`. Дополнительные сигналы:

- node stubs без `type` и без `instance`;
- `[editable ...]` entries;
- property overrides на таких stubs.

### 18.2. Поведение 0.1

Полный Godot-compatible merge не реализуется. Анализатор:

1. Создаёт inheritance edge к base `.tscn`.
2. Если base — resolved `.tscn`, агрегирует base summary один раз.
3. Не прибавляет отдельный `scene_instances += 1` за inherited root.
4. Не считает inherited root отдельным node: root base уже представляет его.
5. Добавляет local nodes с явным `type` как новые nodes.
6. Node stubs без `type` считает overrides, а не новыми nodes.
7. Не применяет property/type/removal/reorder overrides к base summary.
8. Всегда ставит `Analysis: PARTIAL` и `Reliability: approximate`.
9. Выводит `SB1003` с base path.

Почему `approximate`, а не `lower_bound`: override может как включить, так и выключить `shadow_enabled`, а неподдерживаемая семантика удаления/замены может менять counts в обе стороны.

### 18.3. Что остаётся неизвестным

- property overrides base nodes;
- удалённые/заменённые inherited nodes;
- точный порядок и owner semantics;
- editable children nested instances;
- изменения типа через script/custom class;
- inherited scenes поверх imported/binary base.

Budget check на inherited scene допустим только как приблизительный локальный сигнал. Для CI рекомендуется `fail_on_partial: true`.

## 19. Partial и complete analysis

### 19.1. Complete

`COMPLETE` означает:

- root и все достижимые `.tscn` успешно разобраны;
- каждый scene instance, влияющий на effective tree, раскрыт;
- нет inheritance/placeholder/unsupported scene source;
- все объявленные external resource paths синтаксически разрешены и существуют;
- нет cycle;
- все восемь метрик точны **в рамках статической модели 0.1**.

`COMPLETE` не означает, что учтены runtime nodes, script-created lights, draw calls или фактическая производительность.

### 19.2. Partial lower bound

`PARTIAL + lower_bound` используется, если есть missing/imported/unsupported instances, но нет inheritance approximation. Известные counts включают statically resolved часть и один root node каждого unresolved instance; настоящий результат может быть только больше для occurrence metrics.

### 19.3. Partial approximate

`PARTIAL + approximate` используется при inherited scenes/overrides. Некоторые метрики могут быть выше или ниже показанных.

Если одновременно присутствуют lower-bound и inheritance причины, итоговая reliability — `approximate`.

### 19.4. Что вызывает partial

- missing/unreadable nested `.tscn`;
- imported/binary PackedScene instance;
- `instance_placeholder`;
- unsupported `SubResource` PackedScene;
- UID-only или `user://` target;
- resource path за project root;
- missing declared external resource;
- inherited root/override semantics;
- unsupported parent semantics, влияющая на depth.

Обычная ссылка на существующий `.tres`, texture, material, script или audio не делает анализ partial: в 0.1 от неё требуется только участие в `external_resources`, а не deep parsing.

### 19.5. Fatal vs partial

| Ситуация | Результат |
|---|---|
| Root scene отсутствует/не читается | Fatal, exit 2 |
| Root не `.tscn` или `format != 3` | Fatal, exit 2 |
| Root/nested supported `.tscn` синтаксически повреждён | Fatal, exit 2 |
| Project root не найден | Fatal, exit 2 |
| Config invalid | Fatal, exit 2 |
| Dependency cycle | Fatal, exit 2 |
| Nested `.tscn` отсутствует | Partial |
| Imported/binary nested scene | Partial |
| Missing non-scene external resource | Partial |
| Inheritance | Partial approximate |

## 20. Метрики MVP 0.1

Все metrics — non-negative `int64`. Порядок в output и config фиксирован.

| ID | Console name | Точное определение | Multiplicity |
|---|---|---|---|
| `nodes` | Nodes | Effective runtime-like node count после раскрытия `.tscn`; resolved instance root не считается дважды | Per occurrence |
| `tree_depth` | Tree depth | Максимальная глубина effective tree; root depth = 1 | Maximum |
| `scene_instances` | Scene instances | Число nested scene instance occurrences, включая instances внутри каждой копии nested scene; inherited root исключён | Per occurrence |
| `mesh_instances` | Mesh instances | Nodes с literal `type="MeshInstance3D"` после expansion | Per occurrence |
| `lights` | Lights | Nodes с literal type `DirectionalLight3D`, `OmniLight3D`, `SpotLight3D` | Per occurrence |
| `shadow_lights` | Shadow lights | Подмножество `lights`, где effective parsed property `shadow_enabled = true`; отсутствие = Godot default false | Per occurrence |
| `external_resources` | External resources | Число уникальных external resource targets во всех успешно parsed scenes closure | Unique union |
| `scene_dependencies` | Scene dependencies | Число уникальных resolved dependent `.tscn`, root исключён | Unique union |

### 20.1. `nodes`

Для ordinary local node:

```text
nodes += 1
```

Для resolved instance:

```text
nodes += child.nodes
```

Для unresolved instance:

```text
nodes += 1
```

Для inherited root:

```text
nodes += base.nodes + explicit local additions
```

с `PARTIAL approximate`.

### 20.2. `scene_instances`

Для одного resolved child occurrence:

```text
scene_instances += 1 + child.scene_instances
```

Если child встречается `N` раз:

```text
scene_instances += N * (1 + child.scene_instances)
```

Inherited root не прибавляет `1`, но nested instances base scene входят.

### 20.3. `mesh_instances`

Считается только literal built-in type `MeshInstance3D`. Не считаются автоматически:

- custom GDScript class, наследующий `MeshInstance3D`;
- `MultiMeshInstance3D`;
- `CSG*` nodes;
- meshes внутри нераскрытого `.glb/.blend/.scn`.

Это intentional 0.1 definition, а не попытка измерить draw calls.

### 20.4. `lights`

Считаются только 3D light node types:

```text
DirectionalLight3D
OmniLight3D
SpotLight3D
```

Не считаются `Light3D` как abstract type, `PointLight2D`, `DirectionalLight2D` и custom subclasses.

### 20.5. `shadow_lights`

Для поддерживаемых 3D lights:

```text
shadow_enabled = true   → 1
shadow_enabled = false  → 0
property отсутствует    → 0
```

Godot не сохраняет properties, равные default value; `Light3D.shadow_enabled` имеет default `false`. Inherited overrides не применяются в 0.1, поэтому inherited scene всегда approximate.

### 20.6. `external_resources`

Ключ уникальности:

1. Для resolved target — canonical absolute path.
2. Для unresolved declaration — tuple `(declaring_scene, ext_resource_id, raw_path)`.

Учитываются все `[ext_resource]` во всех успешно parsed scenes closure, включая PackedScene dependencies, scripts, textures, materials, audio и `.tres`. Один canonical target, объявленный в нескольких files, считается один раз.

Resources внутри unresolved imported scene неизвестны. Поэтому partial result является lower bound, если нет inheritance.

### 20.7. Пример агрегации

```text
City.tscn
  local ordinary nodes: 10
  Building.tscn × 2      (20 nodes, 1 nested Lamp instance each)
  Lamp.tscn × 3          (4 nodes)
```

Если `Building` summary уже включает одну `Lamp`:

```text
Building.nodes = 20
Building.scene_instances = 1

City.nodes = 10 + 2×20 + 3×4 = 62
City.scene_instances = 2×(1+1) + 3×(1+0) = 7
```

Если и Building, и Lamp ссылаются на одну texture, `external_resources` объединяет её один раз.

## 21. Budget engine

### 21.1. Модель budget

Каждый limit optional. Отсутствующий limit означает: metric показывается, но не проверяется. Ноль — валидный жёсткий limit.

```go
type Budget struct {
    Nodes             *int64 `json:"nodes,omitempty"`
    TreeDepth         *int64 `json:"tree_depth,omitempty"`
    SceneInstances    *int64 `json:"scene_instances,omitempty"`
    MeshInstances     *int64 `json:"mesh_instances,omitempty"`
    Lights            *int64 `json:"lights,omitempty"`
    ShadowLights      *int64 `json:"shadow_lights,omitempty"`
    ExternalResources *int64 `json:"external_resources,omitempty"`
    SceneDependencies *int64 `json:"scene_dependencies,omitempty"`
}
```

Все limits должны быть целыми `>= 0`. Float, string, null и unknown metric keys — config/CLI error.

### 21.2. Проверка

Для каждого настроенного limit:

```text
PASS: actual <= budget
FAIL: actual > budget
delta = actual - budget
```

Нет warning threshold и score в 0.1. Итоговый status:

- `PASSED` — ни один проверяемый budget не превышен;
- `FAILED` — один или больше budgets превышены;
- `INCOMPLETE` — `fail_on_partial=true` и analysis partial.

При partial output значения называются `Observed`, а не гарантированно точными `Actual`. Известное превышение lower bound всё равно показывается как `FAIL`; inheritance approximation помечается `FAIL*`/примечанием в текстовом report.

### 21.3. Результат проверки

```go
type CheckResult struct {
    Metric string
    Actual int64
    Limit  int64
    Delta  int64
    Passed bool
}
```

Порядок результатов совпадает с порядком metrics из раздела 20, а не зависит от map iteration.

## 22. Built-in heuristic presets

### 22.1. Metadata model

```go
type Preset struct {
    ID          string
    Name        string
    Description string
    Platform    string
    Renderer    string
    TargetFPS   int
    Quality     string
    Status      string // heuristic
    Stability   string // experimental
    Budgets     Budget
}
```

Renderer IDs:

```text
forward_plus
mobile
compatibility
unspecified
```

Quality IDs:

```text
low
balanced
high
custom
```

### 22.2. Frozen preset catalog для 0.1.0

Эти значения — **предлагаемые стартовые guardrails**, а не результат сертификации или benchmark guarantee.

| Metric | `mobile` | `steam-deck` | `desktop` |
|---|---:|---:|---:|
| `nodes` | 1,500 | 3,000 | 6,000 |
| `tree_depth` | 15 | 20 | 30 |
| `scene_instances` | 100 | 250 | 500 |
| `mesh_instances` | 500 | 1,000 | 2,500 |
| `lights` | 16 | 32 | 64 |
| `shadow_lights` | 4 | 8 | 16 |
| `external_resources` | 150 | 300 | 600 |
| `scene_dependencies` | 40 | 80 | 160 |

Metadata:

| ID | Platform | Renderer | Target FPS | Quality | Status | Stability |
|---|---|---|---:|---|---|---|
| `mobile` | Mobile-class 3D hardware | `mobile` | 30 | `low` | `heuristic` | `experimental` |
| `steam-deck` | Steam Deck-class hardware | `forward_plus` | 60 | `balanced` | `heuristic` | `experimental` |
| `desktop` | Mid-range desktop hardware | `forward_plus` | 60 | `high` | `heuristic` | `experimental` |

### 22.3. Правила preset data

- Built-ins хранятся как version-controlled JSON и встраиваются через `go:embed`.
- IDs immutable внутри `0.1.x`.
- Tests snapshot-фиксируют metadata и все limits.
- Patch release не должен молча менять limits. Если исправление необходимо, оно отмечается в CHANGELOG.
- Report всегда показывает tool version, preset ID, `Status: heuristic`, renderer/FPS/quality.
- README прямо говорит, что пользователь должен профилировать игру на target hardware.
- Перед 1.0 значения должны пройти отдельное исследование/benchmarking; это не блокирует 0.1 как experimental feature.

## 23. Overrides и custom profiles

### 23.1. Терминология CLI

- `--preset ID` выбирает только built-in preset.
- `--profile ID` выбирает только custom profile из config.
- Flags взаимно исключающие.
- Custom profile может `extends` built-in preset или другой custom profile.

### 23.2. Merge order

От низшего к высшему приоритету:

```text
1. built-in preset или ancestor profile
2. descendant custom profile budgets/metadata
3. top-level config.budgets
4. repeated CLI --budget metric=value
```

Selector resolution:

```text
CLI --preset/--profile
    overrides
config preset/profile
    otherwise
no base selector
```

Top-level budgets применяются независимо от того, откуда выбран base. Это project-wide final overrides.

### 23.3. Profile inheritance

- `extends` может ссылаться на `mobile`, `steam-deck`, `desktop` или custom profile ID.
- Custom profile ID не может совпадать с built-in ID.
- DFS обнаруживает profile cycles и выводит цепочку.
- Максимальная глубина inheritance chain: 32; превышение — config error.
- Metadata child profile заменяет указанные поля parent; неуказанные наследуются.
- Budgets merge field-by-field.
- Custom profile без `extends` получает metadata defaults: platform `custom`, renderer `unspecified`, target FPS `0`, quality `custom`, status `custom`.
- После merge check обязан иметь хотя бы один budget; иначе fatal usage/config error.

Пример:

```text
builtin:steam-deck
  mesh_instances = 1000
  shadow_lights   = 8

profile:shipping
  mesh_instances = 1600

top-level project overrides
  shadow_lights = 6

CLI
  nodes = 4000

effective
  nodes             = 4000
  mesh_instances    = 1600
  shadow_lights     = 6
  остальные limits = steam-deck
```

## 24. `.deadweight.gdt.json` config schema v1

### 24.1. Полный пример

```json
{
  "version": 1,
  "profile": "shipping",
  "fail_on_partial": true,
  "budgets": {
    "mesh_instances": 1800,
    "shadow_lights": 6
  },
  "profiles": {
    "shipping": {
      "name": "Shipping / Steam Deck",
      "description": "Project-calibrated shipping guardrail",
      "extends": "steam-deck",
      "renderer": "forward_plus",
      "target_fps": 60,
      "quality": "balanced",
      "budgets": {
        "nodes": 5000,
        "mesh_instances": 1600
      }
    },
    "office-laptop": {
      "extends": "desktop",
      "platform": "office_laptop",
      "quality": "low",
      "budgets": {
        "nodes": 2500,
        "lights": 24,
        "shadow_lights": 4
      }
    }
  }
}
```

### 24.2. Top-level fields

| Field | Type | Required | Default | Rules |
|---|---|---:|---|---|
| `version` | integer | yes | — | Только `1` |
| `preset` | string | no | — | Built-in ID; mutually exclusive с `profile` |
| `profile` | string | no | — | Key из `profiles`; mutually exclusive с `preset` |
| `fail_on_partial` | boolean | no | `false` | Влияет на `check`, не делает `inspect` fatal |
| `budgets` | object | no | `{}` | Final project overrides |
| `profiles` | object | no | `{}` | Map custom profile ID → profile |

Unknown top-level fields запрещены. JSON decode выполняется с `DisallowUnknownFields`. Trailing content после единственного JSON document запрещён.

### 24.3. Profile fields

| Field | Type | Required | Правила |
|---|---|---:|---|
| `name` | string | no | Human-readable |
| `description` | string | no | Human-readable |
| `extends` | string | no | Built-in/custom ID |
| `platform` | string | no | Stable metadata ID |
| `renderer` | string | no | Enum renderer IDs |
| `target_fps` | integer | no | `>= 0`; `0` = unspecified |
| `quality` | string | no | Enum quality IDs |
| `budgets` | object | no | Budget fields |

Unknown profile fields запрещены.

### 24.4. IDs

Preset/profile IDs должны соответствовать:

```regex
^[a-z0-9][a-z0-9._-]{0,63}$
```

IDs case-sensitive. Рекомендуются kebab-case names.

### 24.5. Canonical JSON Schema

Репозиторий должен содержать `schema/deadweight.gdt.schema.json` на JSON Schema Draft 2020-12. Схема обязана:

- выставить `additionalProperties: false` на каждом object;
- требовать `version`;
- ограничить `version` через `const: 1`;
- описать восемь budget properties как `integer`, `minimum: 0`;
- запретить одновременные top-level `preset` и `profile`;
- проверять ID pattern и renderer/quality enums;
- разрешать optional profiles map через `patternProperties`;
- не пытаться schema-уровнем проверять существование dynamic `extends`; это делает semantic validator.

Canonical schema, которую можно положить в `schema/deadweight.gdt.schema.json`:

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "additionalProperties": false,
  "required": ["version"],
  "properties": {
    "version": { "const": 1 },
    "preset": { "$ref": "#/$defs/id" },
    "profile": { "$ref": "#/$defs/id" },
    "fail_on_partial": { "type": "boolean", "default": false },
    "budgets": { "$ref": "#/$defs/budgets" },
    "profiles": {
      "type": "object",
      "default": {},
      "patternProperties": {
        "^[a-z0-9][a-z0-9._-]{0,63}$": { "$ref": "#/$defs/profile" }
      },
      "additionalProperties": false
    }
  },
  "not": { "required": ["preset", "profile"] },
  "$defs": {
    "id": {
      "type": "string",
      "pattern": "^[a-z0-9][a-z0-9._-]{0,63}$"
    },
    "budgets": {
      "type": "object",
      "additionalProperties": false,
      "properties": {
        "nodes": { "type": "integer", "minimum": 0 },
        "tree_depth": { "type": "integer", "minimum": 0 },
        "scene_instances": { "type": "integer", "minimum": 0 },
        "mesh_instances": { "type": "integer", "minimum": 0 },
        "lights": { "type": "integer", "minimum": 0 },
        "shadow_lights": { "type": "integer", "minimum": 0 },
        "external_resources": { "type": "integer", "minimum": 0 },
        "scene_dependencies": { "type": "integer", "minimum": 0 }
      }
    },
    "profile": {
      "type": "object",
      "additionalProperties": false,
      "properties": {
        "name": { "type": "string" },
        "description": { "type": "string" },
        "extends": { "$ref": "#/$defs/id" },
        "platform": { "type": "string", "minLength": 1 },
        "renderer": {
          "enum": ["forward_plus", "mobile", "compatibility", "unspecified"]
        },
        "target_fps": { "type": "integer", "minimum": 0 },
        "quality": {
          "enum": ["low", "balanced", "high", "custom"]
        },
        "budgets": { "$ref": "#/$defs/budgets" }
      }
    }
  }
}
```

Dynamic правила — существование `preset`/`profile`/`extends`, отсутствие profile cycles, запрет коллизии custom ID с built-in ID и наличие хотя бы одного effective budget для `check` — проверяет semantic validator после JSON decode.

## 25. CLI contract

### 25.1. Общий синтаксис

```text
deadweight.gdt [global flags] <command> [command flags]
```

Global flags:

```text
--project PATH     explicit Godot project root или project.godot
--config PATH      explicit config file
--no-color         отключить ANSI color
--version          версия binary
-h, --help         help
```

Уважать стандартный `NO_COLOR`: если переменная присутствует, ANSI отключён. При non-TTY output color по умолчанию отключён.

### 25.2. `inspect`

```bash
deadweight.gdt inspect <scene>
```

Назначение: вывести metrics, coverage и diagnostics без budget check.

Rules:

- ровно один positional scene argument;
- config не обязателен;
- partial analysis сам по себе возвращает `0`;
- fatal analysis error возвращает `2`;
- `fail_on_partial` из config не меняет поведение inspect.

### 25.3. `check`

```bash
deadweight.gdt check <scene> [flags]
```

Flags:

```text
--preset ID                 built-in preset
--profile ID                custom config profile
--budget METRIC=LIMIT       repeatable final override
--fail-on-partial           partial → exit 3
--allow-partial             override config true → false
```

Rules:

- `--preset` и `--profile` взаимно исключающие;
- `--fail-on-partial` и `--allow-partial` взаимно исключающие;
- CLI selector заменяет config selector;
- repeated `--budget` последовательно применяются; duplicate metric разрешён, последнее значение побеждает;
- unknown metric/negative/non-integer limit — exit `2`;
- без выбранного base и без хотя бы одного effective budget — exit `2` с понятной подсказкой;
- partial check всё равно печатает observed comparisons;
- exit priority описан в разделе 27.

### 25.4. `presets`

```bash
deadweight.gdt presets
```

Показывает только built-in presets. Команда не требует Godot project или config.

### 25.5. `presets show`

```bash
deadweight.gdt presets show <id>
```

Показывает metadata, status/stability, все budgets и disclaimer. Unknown ID → exit `2` с доступными IDs.

В 0.1 нет `profiles` list/show command; custom profiles видны в config. Это допустимое улучшение 0.2.

## 26. Console output

Output должен быть детерминированным, пригодным для screenshot README и readable без color.

### 26.1. `inspect`, complete

```text
$ deadweight.gdt inspect res://levels/city.tscn

deadweight.gdt 0.1.0

Scene:     res://levels/city.tscn
Project:   /home/me/game
Analysis:  COMPLETE
Accuracy:  exact

Structure
  Nodes                      2,841
  Tree depth                    17
  Scene instances              184
  Scene dependencies            12

Rendering
  Mesh instances             1,024
  Lights                        43
  Shadow lights                   9

Resources
  External resources           218

Coverage
  Parsed scene files             13
  Resolved scene instances      184
  Unresolved scene instances      0
```

Thousands separators должны быть стабильны и не зависеть от OS locale: comma `,`.

### 26.2. `inspect`, partial

```text
$ deadweight.gdt inspect res://levels/city.tscn

deadweight.gdt 0.1.0

Scene:     res://levels/city.tscn
Project:   /home/me/game
Analysis:  PARTIAL
Accuracy:  lower bound

Structure
  Nodes                      2,841+
  Tree depth                    17+
  Scene instances              184
  Scene dependencies            12

Rendering
  Mesh instances             1,024+
  Lights                        43+
  Shadow lights                   9+

Resources
  External resources           218+

Coverage
  Parsed scene files             13
  Resolved scene instances      179
  Unresolved scene instances      5

Unresolved scene instances
  res://models/car.glb            ×3  imported PackedScene
  res://models/tree.glb           ×2  imported PackedScene

WARNING: Expanded metrics are partial. Values marked with + are known
lower bounds because 5 scene instances could not be analyzed statically.
```

При `approximate` знак `~` используется вместо `+`, и пояснение говорит, что inheritance overrides могут менять значения в обе стороны.

`scene_instances` можно показывать без `+`, если все instance occurrences обнаружены в parsed closure; но unresolved scene может содержать дополнительные nested instances, поэтому в строгой реализации также допустимо и предпочтительно пометить `scene_instances` как `+`. Для простоты 0.1 разрешено помечать **все occurrence/closure metrics** одним reliability marker. Marker contract должен быть единообразным и snapshot-tested.

### 26.3. `check`, failed preset

```text
$ deadweight.gdt check res://levels/city.tscn --preset steam-deck

deadweight.gdt 0.1.0

Scene:       res://levels/city.tscn
Analysis:    COMPLETE
Preset:      steam-deck
Status:      heuristic (experimental)
Renderer:    Forward+
Target FPS:  60
Quality:     Balanced

Metric                    Actual     Budget   Result
----------------------------------------------------
Nodes                      2,841      3,000   PASS
Tree depth                    17         20   PASS
Scene instances              184        250   PASS
Mesh instances             1,024      1,000   FAIL  +24
Lights                        43         32   FAIL  +11
Shadow lights                   9          8   FAIL  +1
External resources           218        300   PASS
Scene dependencies            12         80   PASS

FAILED — 3 budgets exceeded

Built-in presets are heuristic guardrails, not performance guarantees.
Profile the game on target hardware.
```

### 26.4. `check`, partial rejected

```text
Analysis:  PARTIAL
Accuracy:  lower bound

... comparisons ...

INCOMPLETE — 5 scene instances could not be analyzed statically
Configured policy fail_on_partial=true rejects partial analysis.
```

### 26.5. `presets`

```text
$ deadweight.gdt presets

Built-in presets (heuristic, experimental)

mobile
  Mobile-class 3D hardware
  Renderer: Mobile · Target: 30 FPS · Quality: Low

steam-deck
  Steam Deck-class hardware
  Renderer: Forward+ · Target: 60 FPS · Quality: Balanced

desktop
  Mid-range desktop hardware
  Renderer: Forward+ · Target: 60 FPS · Quality: High

Use `deadweight.gdt presets show <id>` to see budgets.
```

### 26.6. `presets show steam-deck`

```text
$ deadweight.gdt presets show steam-deck

Preset:      Steam Deck
ID:          steam-deck
Status:      heuristic
Stability:   experimental
Renderer:    Forward+
Target FPS:  60
Quality:     Balanced

Budgets
  Nodes                      3,000
  Tree depth                    20
  Scene instances              250
  Mesh instances             1,000
  Lights                        32
  Shadow lights                   8
  External resources           300
  Scene dependencies            80

This preset is a starting guardrail, not a performance guarantee.
```

### 26.7. Error style

Ошибки идут в stderr:

```text
ERROR SB2002: scene dependency cycle

  res://scenes/A.tscn
  → res://scenes/B.tscn
  → res://scenes/C.tscn
  → res://scenes/A.tscn
```

Никаких Go stack traces в обычном CLI. Panic считается bug.

## 27. Exit codes

| Code | Значение |
|---:|---|
| `0` | Command succeeded; `inspect` may be partial; `check` passed or partial allowed |
| `1` | Один или больше budgets превышены |
| `2` | CLI usage, config, project/root, parse, cycle или internal fatal error |
| `3` | Analysis partial и effective `fail_on_partial=true` |

Priority, если одновременно применимы несколько состояний:

```text
fatal error (2)
  > partial rejected (3)
  > budget exceeded (1)
  > success (0)
```

При exit `3` таблица всё равно показывает известные budget exceedances, но итоговая причина — неполный анализ. Это позволяет CI отличать «реально превышен budget» от «анализатор не смог доказать полноту».

Cobra по умолчанию может возвращать usage errors как `2`; приложение должно централизованно преобразовывать domain results в эти codes, а `main.go` должен быть тонким.

## 28. Детерминированность и UX

- Metrics всегда в фиксированном порядке.
- Presets сортируются в продуктовом порядке: `mobile`, `steam-deck`, `desktop`, а не map order.
- Diagnostics сортируются по severity, code, display path, line, resource.
- Grouped unresolved entries сортируются по display path/reason.
- Paths в пользовательском output по возможности `res://`; absolute project path показывается только в header/error, где это полезно.
- Color никогда не является единственным носителем смысла: присутствуют `PASS`, `FAIL`, `WARNING`.
- Output при redirect не содержит ANSI.
- Ошибки содержат действие: например «run from inside a Godot project or pass `--project`».
- Built-in preset disclaimer показывается в `check` и `presets show`.

## 29. Tests и testdata

### 29.1. Unit tests: lexer/parser

Обязательные cases:

- minimal `format=3` scene;
- header attributes в разном порядке;
- string и numeric resource IDs;
- escaped quote и semicolon внутри string;
- comment после header/property;
- multi-line arrays/dictionaries/constructors;
- огромный `PackedByteArray` безопасно пропускается;
- `ExtResource("id")` parse;
- `shadow_enabled` true/false/absent;
- unknown section/property;
- duplicate ExtResource ID;
- missing `gd_scene`;
- `format=2` rejection;
- malformed/unclosed string/bracket;
- root, ordinary node, instance node, inherited root, override stub;
- no panic на arbitrary bytes.

Добавить fuzz target:

```go
func FuzzParse(f *testing.F)
```

Инвариант: parser либо возвращает typed result/error, либо завершает работу; panic, hang и unbounded allocation недопустимы.

### 29.2. Unit tests: project paths

- nearest `project.godot` wins;
- explicit `--project` directory/file;
- `res://` from cwd inside project;
- relative scene input;
- relative ext resource from declaring scene;
- `../` внутри project;
- попытка escape за root;
- symlink escape;
- `uid://`, `user://`, absolute inside/outside root;
- normalization to forward-slash `res://` display;
- Windows path behavior tests через platform-neutral helpers, где возможно.

### 29.3. Unit tests: graph/cache/aggregation

- simple chain `A → B → C`;
- diamond `A → B`, `A → C`, `B → D`, `C → D`;
- repeated instance ×100;
- self-cycle;
- multi-node cycle с точной цепочкой;
- parse cache открывает D один раз в diamond/repeated cases;
- summary cache корректно умножает occurrence metrics;
- unique dependencies/resources не умножаются;
- instance root не double-counted;
- depth mount formula;
- unresolved child даёт +1 known node и partial lower bound;
- arithmetic overflow возвращает `SB2004`.

### 29.4. Unit tests: budget/config/presets

- absent budget metric не проверяется;
- zero budget валиден;
- boundary `actual == limit` pass;
- `actual == limit+1` fail;
- merge order всех четырёх layers;
- profile extends built-in;
- profile extends profile;
- profile cycle;
- missing parent;
- ID collision с built-in;
- unknown fields и metrics rejected;
- negative/float/string/null limits rejected;
- selector mutual exclusion;
- fail/allow partial CLI precedence;
- snapshots exact built-in preset values и metadata.

### 29.5. Integration testdata projects

```text
testdata/projects/complete/
  project.godot
  simple.tscn
  nested.tscn
  deps/child.tscn

testdata/projects/repeated/
  project.godot
  city.tscn
  lamp.tscn

testdata/projects/relative-paths/
  project.godot
  levels/city.tscn
  props/lamp.tscn

testdata/projects/unresolved/
  project.godot
  missing-tscn.tscn
  imported-glb.tscn
  placeholder.tscn

testdata/projects/inherited/
  project.godot
  enemy.tscn
  zombie.tscn

testdata/projects/cyclic/
  project.godot
  A.tscn
  B.tscn
  C.tscn

testdata/projects/malformed/
  project.godot
  format2.tscn
  unclosed-string.tscn
  bad-ext-id.tscn
```

Каждый fixture должен быть минимальным и содержать комментарий в соседнем README или имя, объясняющее expected behavior.

### 29.6. Golden CLI tests

Snapshot-test:

- complete inspect;
- partial lower-bound inspect;
- inherited approximate inspect;
- passing check;
- failing check;
- partial rejected check;
- presets list/show;
- cycle error;
- missing project/root config errors;
- `NO_COLOR` output.

Golden output запускается с фиксированными root placeholders, например `<PROJECT>`, чтобы snapshots не зависели от temp directory.

### 29.7. CI quality gate

Минимум:

```bash
go test ./...
go test -race ./...
go vet ./...
```

CI matrix: Linux, macOS, Windows на поддерживаемой stable Go release. Godot не устанавливается.

## 30. Acceptance criteria MVP 0.1

MVP считается готовым, когда выполнены все пункты:

1. `deadweight.gdt inspect` корректно анализирует simple fixture.
2. Root scene можно передать absolute, relative и `res://` path.
3. Project root auto-discovery выбирает ближайший `project.godot`.
4. Relative `ExtResource.path` разрешается относительно declaring `.tscn`.
5. Repeated scene парсится один раз и агрегируется с правильной multiplicity.
6. Resolved instance root не double-counted в `nodes`.
7. `tree_depth` следует root-depth-1 и mount formula из спецификации.
8. Все восемь metrics имеют unit/integration tests с точными expected values.
9. `external_resources` и `scene_dependencies` используют unique semantics.
10. Cycle `A → B → C → A` возвращает code `2` и печатает полную цепочку.
11. Imported/missing nested scene не выдаёт ложный `COMPLETE`.
12. Inherited scene обнаруживается и даёт `PARTIAL approximate`.
13. Parser пропускает complex unknown values без regex-based corruption и без сохранения больших blobs.
14. Встроенные presets доступны через `presets` и `presets show`.
15. Значения `mobile`, `steam-deck`, `desktop` совпадают с таблицей 22.2.
16. `check --preset steam-deck` применяет все восемь limits.
17. Project overrides, custom profiles и CLI overrides соблюдают merge order.
18. Profile/config cycles и unknown keys дают actionable exit `2`.
19. Budget exceed даёт exit `1`.
20. Partial + `fail_on_partial=true` даёт exit `3` независимо от budget verdict.
21. `inspect` partial без fatal error даёт exit `0` и заметный warning.
22. Console output детерминирован и проходит golden tests без ANSI.
23. `go test ./...`, `go test -race ./...` и `go vet ./...` проходят.
24. CLI не ищет, не запускает и не требует Godot executable.
25. README содержит positioning и disclaimers из раздела 33.

## 31. Milestones разработки

### Milestone 0 — repository skeleton

Результат:

- Go module;
- `cmd/deadweight.gdt`;
- Cobra root/help/version;
- core models `metrics`, `diagnostic`;
- CI с `go test`/`go vet`;
- пустые команды с compile-time wiring.

Gate: проект собирается на Linux/macOS/Windows, tests green.

### Milestone 1 — TSCN subset parser

Результат:

- lexer с positions/comments/strings/nesting;
- parser sections/header attrs;
- AST subset;
- streaming skip unknown values;
- parser unit/fuzz tests.

Gate: parser извлекает ext resources, nodes, instances и shadow property из testdata; malformed input даёт typed errors.

### Milestone 2 — project/path resolution

Результат:

- project root finder;
- filesystem/`res://` input resolution;
- relative ExtResource resolution;
- canonical/display paths;
- root-boundary/symlink checks.

Gate: полный path test matrix green.

### Milestone 3 — graph, recursion и cache

Результат:

- ExtResource → scene edge resolution;
- DFS graph;
- cycle detection с chain;
- parse и summary caches;
- repeated occurrence aggregation.

Gate: chain/diamond/repeated/cycle tests green; instrumentation доказывает one parse per file.

### Milestone 4 — metrics и completeness

Результат:

- все восемь metrics;
- mount depth;
- unique resource/dependency sets;
- unresolved grouping/coverage;
- complete/partial/reliability;
- inherited-scene detection и approximate behavior.

Gate: integration fixture expected metrics green, нет silent partial.

### Milestone 5 — presets, config и budget engine

Результат:

- embedded built-in presets;
- JSON config v1 + JSON Schema;
- profile inheritance/cycle detection;
- merge order;
- checker results;
- fail-on-partial policy.

Gate: config/budget/preset test matrix green.

### Milestone 6 — CLI reports и exit codes

Результат:

- `inspect`, `check`, `presets`, `presets show`;
- deterministic text output;
- ANSI/NO_COLOR rules;
- exact exit mapping;
- golden tests.

Gate: acceptance CLI examples воспроизводятся.

### Milestone 7 — release hardening

Результат:

- README, LICENSE, CHANGELOG;
- install instructions (`go install`, release binaries later);
- cross-platform CI;
- `-race`/vet;
- minimal benchmarks parser/repeated graph;
- release checklist/tag `v0.1.0`.

Gate: все acceptance criteria выполнены, roadmap features не просочились в scope.

## 32. Implementation order для Codex

Codex должен реализовывать вертикальными, проверяемыми slices и не начинать с красивого report до корректного ядра.

### Шаг 1. Зафиксировать domain types и invariants

Создать `metrics`, `diagnostic`, TSCN AST contracts и typed errors. Добавить table-driven tests для metric names/order. Не добавлять parser logic в CLI.

### Шаг 2. Реализовать lexer/parser отдельно

Работать только с in-memory strings/fixtures. Сначала headers/strings/comments, затем balanced skip values, затем recognized property. Добавить fuzz seed corpus до scene resolver.

### Шаг 3. Реализовать project/path layer

Не смешивать filesystem semantics с parser. Все `res://`/relative/canonical/display transformations проходят через один Resolver.

### Шаг 4. Реализовать single-scene local summary

Без recursion посчитать ordinary nodes/types/local depth/ext resource keys и выделить instance mounts. Тестировать root counting отдельно.

### Шаг 5. Добавить DFS, dependency graph и cache

Сначала resolved `.tscn` chain, затем repeated instances и unique sets, затем cycles. Использовать injectable file opener/parse function для проверки cache behavior.

### Шаг 6. Добавить unresolved/inheritance semantics

Каждая unsupported branch обязана создавать diagnostic, coverage change и правильный reliability. Никаких `continue` без статуса partial.

### Шаг 7. Добавить budget/preset/config

Budget engine принимает готовые `Metrics` и не читает сцены. Сначала checker, затем built-ins, затем profile resolution/merge.

### Шаг 8. Подключить CLI и reports

CLI только orchestrates application service и отображает domain result. Централизовать exit codes. Golden tests запускать без color.

### Шаг 9. Закрыть acceptance matrix

Пройти каждый пункт раздела 30, добавить отсутствующий test, обновить README и только потом считать 0.1 готовой.

### Правила работы Codex

- Перед изменениями прочитать текущие tests и эту спецификацию.
- Не заменять state machine набором regex.
- Не использовать map iteration для user-visible output.
- Не добавлять `.tres`, JSON output, GitHub Action или Godot bridge в 0.1.
- После каждого milestone запускать targeted tests, затем `go test ./...`.
- Любое расхождение с frozen semantics вынести отдельным decision note, а не менять молча.
- Сохранять публичные domain interfaces маленькими; почти всё может оставаться в `internal/`.
- Не оптимизировать concurrency до появления benchmark: cache важнее parallel parsing в 0.1.

## 33. README positioning и обязательные disclaimers

### 33.1. Рекомендуемый opening

```markdown
# deadweight.gdt

deadweight.gdt is a standalone Go CLI that statically analyzes Godot 4 `.tscn`
scenes, expands nested text scenes, and checks scene-complexity metrics against
project budgets.

It does not require Godot to be installed.
```

### 33.2. Короткое value proposition

```markdown
$ deadweight.gdt check res://levels/city.tscn --preset steam-deck

Mesh instances   1,024 / 1,000   FAIL
Lights              43 /    32   FAIL
Shadow lights         9 /     8   FAIL
```

### 33.3. Обязательный preset disclaimer

README должен содержать дословно или эквивалентно:

> **Built-in presets are heuristic guardrails, not performance guarantees. Always profile your game on target hardware.**

И дополнительно объяснить:

- `steam-deck` не является официальным Valve certification profile;
- одинаковое число nodes может иметь совершенно разную стоимость;
- renderer, scripts, physics, shaders, materials, view/culling и runtime behavior не моделируются;
- presets предназначены как starting point и должны калиброваться под проект.

### 33.4. Обязательный static-analysis disclaimer

```markdown
deadweight.gdt reads serialized text scenes. It cannot see nodes created by scripts
at runtime and cannot expand imported or binary scenes without Godot's import
pipeline. When this affects the result, the report is marked PARTIAL.
```

### 33.5. README sections для 0.1

1. What/why.
2. Screenshot/terminal example.
3. Install.
4. Quick start.
5. Metrics definitions.
6. Presets and custom profiles.
7. Config example.
8. Complete vs partial.
9. Supported/unsupported Godot inputs.
10. Exit codes/CI example.
11. Roadmap.
12. Contributing/license.

Не писать «guarantees 60 FPS», «Steam Deck compatible» или «performance profiler».

## 34. Roadmap 0.2+

### 0.2 — deeper static analysis и automation-friendly output

Кандидаты:

- deep parsing `.tres` dependency closure;
- machine-readable `--format json` со schema version;
- dependency tree/graph command;
- per-scene contribution breakdown (`top contributors`);
- custom profile list/show;
- baseline/diff между двумя reports;
- confidence metadata per metric;
- дополнительные metrics только после точного определения семантики.

### 0.3 — hard Godot semantics

Кандидаты:

- inherited scene merge;
- editable children/overrides;
- UID resolution;
- optional Godot bridge/headless mode;
- imported `.glb/.gltf/.blend/.scn` expansion через Godot import pipeline;
- optional exact custom class hierarchy resolution.

Standalone static mode должен остаться первым классом продукта, даже если появляется optional bridge.

### 0.4 — CI и ecosystem

Кандидаты:

- official GitHub Action;
- SARIF/JUnit/GitHub annotations;
- PR comment с delta;
- release binaries, Homebrew/Scoop packages;
- project-wide multi-scene checks;
- ignore/include patterns;
- watch/incremental on-disk cache.

### Позже

- Godot editor plugin как thin frontend;
- HTML reports;
- texture/VRAM/material/triangle metrics;
- renderer-aware rules;
- calibrated community presets;
- plugin API для custom metrics/rules.

## 35. Риски и способы их удержать

| Риск | Как удерживается в 0.1 |
|---|---|
| Попытка написать полный TSCN parser | Чёткий subset и streaming skip unknown values |
| Ложная точность imported scenes | `PARTIAL`, coverage, fail-on-partial |
| Double count instance root | Зафиксированная aggregation formula |
| Repeated scenes делают анализ медленным | Parse + expanded-summary cache |
| Cycle вызывает recursion overflow | DFS colors + explicit cycle error |
| Presets воспринимаются как гарантия FPS | `heuristic`/`experimental` в data, output и README |
| Config inheritance становится неуправляемым | Single inheritance, depth 32, cycle detection, strict schema |
| Cross-platform path bugs | Централизованный Resolver и OS CI matrix |
| Большие embedded arrays съедают память | Streaming lexer, unknown values не materialize |
| Scope расползается | Frozen non-goals и milestone gates |

## 36. Frozen decisions для 0.1

Следующие решения считаются закрытыми до явного изменения спецификации:

- бинарник называется `deadweight.gdt`;
- конфиг называется `.deadweight.gdt.json`;
- root input — ровно одна `.tscn`;
- Godot не требуется;
- target — `format=3`;
- recursion по nested `.tscn` обязательна;
- cycle fatal;
- inherited scenes partial approximate;
- imported/binary instances partial;
- root depth = 1;
- resolved instance root не double-counted;
- occurrence metrics умножаются, unique metrics объединяются;
- восемь metric IDs фиксированы;
- limits — inclusive upper bounds;
- built-in preset IDs и значения фиксированы для `0.1.0`;
- preset и profile — разные термины/CLI flags;
- config strict JSON v1;
- default `fail_on_partial=false`;
- exit codes `0/1/2/3` определены разделом 27;
- console text — единственный report format 0.1.

## 37. Технические источники

Спецификация опирается на документированную семантику Godot:

- [Godot 4.6: TSCN file format](https://docs.godotengine.org/en/4.6/engine_details/file_formats/tscn.html) — `format=3`, sections, nodes, `ExtResource`, relative paths и отсутствие default-valued properties в сериализации.
- [Godot: File system](https://docs.godotengine.org/en/stable/tutorials/scripting/filesystem.html) — `project.godot` определяет project root, `res://` указывает на него.
- [Godot 4.4: Light3D](https://docs.godotengine.org/en/4.4/classes/class_light3d.html) — `shadow_enabled` default `false` и performance cost shadows.
- [Godot: Optimizing 3D performance](https://docs.godotengine.org/en/stable/tutorials/performance/optimizing_3d_performance.html) — realtime lights и shadows зависят от сцены и могут быть дорогими, особенно на mobile-class hardware.

Документация Godot указывает, что `format=3` использует string-based resource IDs, но parser 0.1 намеренно принимает и numeric ID representation как дешёвую совместимость с ручными/переходными fixtures.

---

## Итоговая граница MVP

deadweight.gdt 0.1 завершён, когда он надёжно выполняет одну узкую задачу:

> Берёт одну Godot 4 `.tscn`, честно раскрывает все доступные nested `.tscn`, явно сообщает о недоступной части, считает восемь строго определённых metrics и проверяет их по воспроизводимому heuristic/custom budget.

Всё, что требует runtime, Godot import pipeline или более глубокого resource understanding, остаётся видимым ограничением и будущим roadmap, а не скрытой неточностью 0.1.
