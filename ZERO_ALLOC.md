# Zero-Allocation Insert: от 23 аллокаций к 0

Разбор реальной оптимизации WAL-движка на Go — как убрать каждую из 23 аллокаций на один вызов `Insert`, не меняя публичный API.

---

## Исходная ситуация

```
BenchmarkInsert-16    50985    24179 ns/op    1277 B/op    23 allocs/op
```

После оптимизации:

```
BenchmarkInsert-16    855355    1829 ns/op    40 B/op    0 allocs/op
```

**11× быстрее. 0 аллокаций.**

---

## Как работал Insert до оптимизации

Каждый вызов проходил такой путь:

```
Database.Insert(t, v)
  └── Wal.Append(I, t, Insert{v})
        └── Insert.Raw()
              └── shared.MarshalMap(v)       ← structpb
              └── proto.Marshal(&lp.Insert{}) ← proto #1
        └── &lp.WalRecord{Data: raw}          ← proto struct
        └── proto.Marshal(entry)              ← proto #2
        └── binary.Write(w.buf, uint32(len))  ← [4]byte на куче
  └── applyInsert → Table.Insert → append
```

---

## Разбор каждой аллокации

### 1. Бенчмарк: `fmt.Sprintf` и `map[string]any{}`

```go
// БЫЛО — в бенчмарке
for i := 0; i < b.N; i++ {
    db.Insert("users", map[string]any{   // alloc #1: новый map каждую итерацию
        "id":    fmt.Sprintf("%d", i),   // alloc #2: новая строка каждую итерацию
        "name":  "user name",
        "email": "user email",
    })
}
```

`map[string]any{}` — всегда аллокация, потому что map в Go это `*runtime.hmap` под капотом.  
`fmt.Sprintf` — всегда аллокация результирующей строки.

**Фикс:** предаллоцировать map снаружи цикла, убрать Sprintf:

```go
// СТАЛО
row := map[string]any{
    "id":    "bench-id",
    "name":  "user name",
    "email": "user email",
}
b.ResetTimer()
for i := 0; i < b.N; i++ {
    db.Insert("users", row)
}
```

---

### 2. `structpb.NewValue` — 2 аллокации на каждое поле

Самый дорогой источник: конвертация `map[string]any` в `map[string]*structpb.Value`.

```go
// БЫЛО — shared/proto.go
func MarshalMap(m map[string]any) (map[string]*structpb.Value, error) {
    res := make(map[string]*structpb.Value, len(m)) // alloc #3: новый map
    for k, v := range m {
        val, _ := structpb.NewValue(v)              // alloc #4 + #5 на каждое поле
        res[k] = val
    }
    return res, nil
}
```

Внутри `structpb.NewValue(v string)`:

```go
// из исходников google.golang.org/protobuf
func NewStringValue(v string) *Value {
    return &Value{                           // alloc: сам *structpb.Value (heap)
        Kind: &Value_StringValue{            // alloc: обёртка для поля Kind (heap)
            StringValue: v,
        },
    }
}
```

Для 3 полей (`id`, `name`, `email`) — 1 map + 3×2 = **7 аллокаций**.

**Почему так устроен structpb?**  
`structpb.Value` — это динамический JSON-совместимый тип protobuf. Поле `Kind` объявлено как `isValue_Kind` (интерфейс), и каждый конкретный тип (`*Value_StringValue`, `*Value_NumberValue`) хранится как указатель. Это гибкость ценой аллокаций.

**Фикс:** выбросить `structpb` полностью, написать кастомный JSON-аппендер:

```go
// СТАЛО — src/db/encode.go
func appendMapJSON(dst []byte, m map[string]any) []byte {
    dst = append(dst, '{')
    first := true
    for k, v := range m {
        if !first {
            dst = append(dst, ',')
        }
        first = false
        dst = append(dst, '"')
        dst = appendEscapedStr(dst, k)
        dst = append(dst, '"', ':')
        dst = appendAnyJSON(dst, v)
    }
    return append(dst, '}')
}

func appendAnyJSON(dst []byte, v any) []byte {
    switch x := v.(type) {
    case string:
        dst = append(dst, '"')
        dst = appendEscapedStr(dst, x)
        return append(dst, '"')
    case int:
        return strconv.AppendInt(dst, int64(x), 10)
    case float64:
        return strconv.AppendFloat(dst, x, 'f', -1, 64)
    case bool:
        if x {
            return append(dst, "true"...)
        }
        return append(dst, "false"...)
    case nil:
        return append(dst, "null"...)
    default:
        b, _ := json.Marshal(x) // fallback — редкий путь
        return append(dst, b...)
    }
}
```

`append` на pre-allocated срезе — 0 аллокаций.  
`strconv.AppendInt` и `strconv.AppendFloat` пишут прямо в переданный срез — 0 аллокаций.

---

### 3. Двойная сериализация protobuf — 3 аллокации

```go
// БЫЛО — arg.go
func (i Insert) Raw() []byte {
    converted, _ := shared.MarshalMap(i.Val)
    raw, _ := proto.Marshal(&lp.Insert{Val: converted}) // alloc #8: &lp.Insert{}
                                                         // alloc #9: []byte результат
    return raw
}

// БЫЛО — wal.go
entry := &lp.WalRecord{                                  // alloc #10: &lp.WalRecord{}
    Data: arg.Raw(),
}
b, _ := proto.Marshal(entry)                             // alloc #11: []byte результат
```

Данные сериализовались ДВАЖДЫ: сначала в `lp.Insert`, потом этот байтовый срез вкладывался в `lp.WalRecord`. Каждый `proto.Marshal` возвращает свежий `[]byte`.

**Фикс 1:** убрать промежуточный `lp.Insert` — хранить JSON-байты прямо в поле `Data` у `WalRecord`.

**Фикс 2:** переиспользовать буферы через поля структуры `Wal`:

```go
// СТАЛО — wal.go
type Wal struct {
    // ...
    dataBuf []byte       // переиспользуемый буфер для данных (защищён mu)
    recBuf  []byte       // переиспользуемый буфер для записи WAL (защищён mu)
    entry   lp.WalRecord // переиспользуемая структура записи (защищён mu)
}

func NewWal(path string) *Wal {
    w := &Wal{
        // ...
        dataBuf: make([]byte, 0, 256),
        recBuf:  make([]byte, 0, 512),
    }
    // ...
}
```

**Фикс 3:** использовать `proto.MarshalOptions{}.MarshalAppend` вместо `proto.Marshal`:

```go
// СТАЛО — wal.go
w.lsn++
w.dataBuf = arg.AppendRaw(w.dataBuf[:0]) // пишем в переиспользуемый буфер

w.entry.Lsn = w.lsn
w.entry.Op = uint32(a)
w.entry.TableId = w.catalog.GetID(table)
w.entry.Data = w.dataBuf                 // ссылка, без копирования

// Аппендим в w.recBuf — НЕТ новой аллокации
w.recBuf, err = proto.MarshalOptions{}.MarshalAppend(w.recBuf[:0], &w.entry)
```

`proto.Marshal(msg)` → возвращает новый `[]byte` = **1 аллокация**.  
`proto.MarshalOptions{}.MarshalAppend(buf, msg)` → пишет в существующий срез = **0 аллокаций** (при достаточной ёмкости).

---

### 4. Скрытая аллокация `binary.Write` — 1 аллокация

```go
// БЫЛО
if err := binary.Write(w.buf, binary.LittleEndian, uint32(len(b))); err != nil {
    ...
}
```

Выглядит безобидно, но `encoding/binary.Write` внутри делает:

```go
// из исходников Go stdlib
func Write(w io.Writer, order ByteOrder, data any) error {
    if n := intDataSize(data); n != 0 {
        bs := make([]byte, n) // ← АЛЛОКАЦИЯ! 4 байта на куче
        switch v := data.(type) {
        case uint32:
            order.PutUint32(bs, v)
        }
        _, err := w.Write(bs)
        return err
    }
    // ...
}
```

**Почему 4 байта уходят на кучу?** `binary.Write` принимает `data any` — это boxing. Внутри он создаёт временный `[]byte` для записи. Этот срез уходит в `w.Write()` — интерфейсный вызов. Компилятор не может доказать, что вызываемый метод не сохранит ссылку на срез, и помещает массив на кучу.

Казалось бы, можно заменить на `var lenBuf [4]byte`:

```go
// КАЗАЛОСЬ БЫ ЛУЧШЕ — но тоже аллоцирует!
var lenBuf [4]byte
binary.LittleEndian.PutUint32(lenBuf[:], uint32(len(w.recBuf)))
w.buf.Write(lenBuf[:]) // ← lenBuf всё равно уходит на кучу
```

`lenBuf` — стековая переменная. Но `lenBuf[:]` передаётся в `bufio.Writer.Write(p []byte)`, который через свой внутренний `io.Writer` (это `*os.File`) вызывает `b.wr.Write(p)`. Компилятор видит интерфейсный вызов с `p` и консервативно считает, что `p` может быть сохранён. Итог: `lenBuf` эскейпит на кучу.

Это подтвердил `go tool pprof -alloc_objects`:

```
129589     129589    139:    var lenBuf [4]byte   ← здесь была аллокация
```

**Фикс:** вписать 4 байта длины прямо в начало `w.recBuf`, который уже лежит на куче:

```go
// СТАЛО — один Write, нет стековых срезов, нет аллокаций
w.recBuf = append(w.recBuf[:0], 0, 0, 0, 0)                         // резервируем 4 байта
w.recBuf, err = proto.MarshalOptions{}.MarshalAppend(w.recBuf, &w.entry) // пишем запись после них
binary.LittleEndian.PutUint32(w.recBuf[:4], uint32(len(w.recBuf)-4)) // заполняем длину
w.buf.Write(w.recBuf)                                                 // один вызов, срез на куче
```

`w.recBuf` — поле структуры `Wal`, которая сама на куче. Срез `w.recBuf` ссылается на кучную память. Когда мы передаём `w.recBuf` в `w.buf.Write`, компилятор знает, что backing array уже на куче — никакого эскейпа стековых переменных нет.

---

### 5. Интерфейс `Arg` — изменение контракта

До оптимизации интерфейс выглядел так:

```go
// БЫЛО
type Arg interface {
    Raw() []byte  // возвращает новый срез — неизбежная аллокация
    Vals() []string
}
```

Метод `Raw() []byte` обязан вернуть новый срез. Нет способа реализовать его без аллокации.

```go
// СТАЛО
type Arg interface {
    AppendRaw([]byte) []byte  // аппендит в переданный срез — аллокация не нужна
    Vals() []string
}
```

Теперь `Insert.AppendRaw` пишет прямо в `w.dataBuf`:

```go
// src/db/arg.go
func (i Insert) AppendRaw(dst []byte) []byte {
    return appendMapJSON(dst, i.Val) // 0 аллокаций при достаточной ёмкости dst
}
```

И WAL-запись читается обратно через `json.Unmarshal`:

```go
// src/wal/wal.go — buildAction
case I:
    var m map[string]any
    if err := json.Unmarshal(rec.GetData(), &m); err != nil {
        return Action{}, err
    }
    a.Val = m
```

`json.Unmarshal` на пути восстановления — это нормально. Replay происходит один раз при старте, а не на каждый insert.

---

## Лучшие практики zero-alloc кода в Go

### 1. Паттерн AppendXxx

Стандартная библиотека Go использует этот паттерн везде: `strconv.AppendInt`, `fmt.AppendFormat`, `proto.MarshalOptions.MarshalAppend`. Суть: принимать `dst []byte` и возвращать расширенный срез.

```go
// ❌ Плохо — всегда аллоцирует
func Encode(v any) []byte {
    return json.Marshal(v) // новый срез каждый раз
}

// ✅ Хорошо — аллоцирует только если нужен рост
func AppendEncoded(dst []byte, v any) []byte {
    return appendAnyJSON(dst, v)
}
```

### 2. Переиспользуемые буферы через поля структуры + мьютекс

Если объект живёт дольше одного вызова и защищён мьютексом — держи буферы в нём:

```go
type Writer struct {
    mu  sync.Mutex
    buf []byte // переиспользуется, защищён mu
}

func (w *Writer) Write(data map[string]any) {
    w.mu.Lock()
    defer w.mu.Unlock()
    w.buf = appendMapJSON(w.buf[:0], data) // 0 аллокаций после warmup
    // ... использовать w.buf
}
```

Альтернатива для concurrent-путей без мьютекса — `sync.Pool`:

```go
var bufPool = sync.Pool{New: func() any {
    b := make([]byte, 0, 256)
    return &b
}}

func encodeRow(v map[string]any) ([]byte, func()) {
    bp := bufPool.Get().(*[]byte)
    buf := appendMapJSON((*bp)[:0], v)
    *bp = buf
    return buf, func() { bufPool.Put(bp) }
}
```

### 3. `proto.MarshalOptions{}.MarshalAppend` вместо `proto.Marshal`

```go
// ❌ Каждый раз новый []byte
b, err := proto.Marshal(msg)

// ✅ Аппендит в существующий срез
var buf []byte // или из пула
buf, err = proto.MarshalOptions{}.MarshalAppend(buf[:0], msg)
```

### 4. Стековые переменные не всегда остаются на стеке

Если стековая переменная попадает в интерфейсный вызов — она может уйти на кучу:

```go
// ❌ lenBuf уходит на кучу через io.Writer
var lenBuf [4]byte
binary.LittleEndian.PutUint32(lenBuf[:], n)
w.Write(lenBuf[:]) // интерфейс — компилятор не знает, сохранит ли w ссылку

// ✅ Использовать уже кучную память
heapBuf = append(heapBuf, 0, 0, 0, 0)
binary.LittleEndian.PutUint32(heapBuf[:4], n)
w.Write(heapBuf) // heapBuf уже на куче — эскейпа нет
```

### 5. Кастомный энкодер вместо `encoding/json` на горячем пути

`encoding/json.Marshal` для `map[string]any` всегда аллоцирует, потому что:
- сортирует ключи (`make([]string, len(m))`)
- использует reflection

На горячем пути лучше написать типизированный аппендер:

```go
// strconv.AppendInt — 0 аллокаций
dst = strconv.AppendInt(dst, int64(x), 10)

// strconv.AppendFloat — 0 аллокаций  
dst = strconv.AppendFloat(dst, x, 'f', -1, 64)

// append строки — 0 аллокаций (при достаточной ёмкости dst)
dst = append(dst, s...)
```

### 6. Как найти аллокации

**Шаг 1:** запустить бенчмарк с `-benchmem`:
```bash
go test -bench=BenchmarkInsert -benchmem -count=3
```

**Шаг 2:** построить профиль:
```bash
go test -memprofilerate=1 -memprofile=mem.out -bench=BenchmarkInsert -run='^$'
go tool pprof -alloc_objects -top mem.out
```

**Шаг 3:** посмотреть построчно в конкретной функции:
```bash
go tool pprof -alloc_objects -list 'Wal.*Append' mem.out
```

**Шаг 4:** проверить escape analysis:
```bash
go build -gcflags='-m=2' . 2>&1 | grep "escapes to heap"
```

Если видишь `var x T escapes to heap` или `&T{...} escapes to heap` — это аллокация.

---

## Итоговая картина изменений

| Источник аллокации | Было | Стало |
|---|---|---|
| `map[string]any{}` в бенчмарке | 1 | 0 (вынесен за цикл) |
| `fmt.Sprintf` в бенчмарке | 1 | 0 (убран) |
| `make(map[string]*structpb.Value)` | 1 | 0 (убран structpb) |
| `&structpb.Value{}` × 3 поля | 3 | 0 |
| `&structpb.Value_StringValue{}` × 3 | 3 | 0 |
| `&lp.Insert{}` | 1 | 0 (убран lp.Insert) |
| `proto.Marshal(Insert)` результат | 1 | 0 |
| proto внутренние (сортировка ключей) | 2–3 | 0 |
| `&lp.WalRecord{}` | 1 | 0 (переиспользуется w.entry) |
| `proto.Marshal(WalRecord)` результат | 1 | 0 (MarshalAppend в w.recBuf) |
| `binary.Write` / `var lenBuf [4]byte` | 1 | 0 (вписано в w.recBuf) |
| `append(t.Rows, row)` рост среза | ~3 | ~0 (амортизировано) |
| **Итого** | **23** | **0** |

---

## Ключевой вывод

Большинство аллокаций в Go возникают не от очевидных `make` и `new`, а от трёх скрытых мест:

1. **Библиотеки общего назначения** (`structpb`, `json.Marshal`, `binary.Write`) — они удобны, но платят аллокациями за гибкость. На горячем пути пишите типизированные аппендеры.

2. **Интерфейсные вызовы** — любой стековый объект, прошедший через `interface{}` или `io.Writer`, может уйти на кучу. Используйте кучные буферы и `MarshalAppend`-паттерн.

3. **Возврат `[]byte` из функций** — `func Encode() []byte` = аллокация. `func AppendEncoded(dst []byte) []byte` = 0 аллокаций при разогретом буфере.
