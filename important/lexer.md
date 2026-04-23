# 🧠 Understanding Lexers: From Theory to Practice

## 📌 What is a Lexer?

A **lexer (lexical analyzer)** is the first step in processing text for programming languages, query engines, or interpreters.

It takes raw input (a string) and converts it into **tokens** — small meaningful pieces.

---

## 🔁 Full Pipeline

```text
Input → Lexer → Tokens → Parser → AST → Execution
```

---

## ✨ Simple Example

### Input

```sql
SELECT name FROM users WHERE age > 18;
```

### Output Tokens

```text
[SELECT]
[IDENT(name)]
[FROM]
[IDENT(users)]
[WHERE]
[IDENT(age)]
[GT]
[NUMBER(18)]
[SEMICOLON]
```

---

## 🧩 What is a Token?

```go
type Token struct {
    Type  TokenType
    Value string
}
```

### Example

```go
Token{Type: IDENT, Value: "users"}
Token{Type: NUMBER, Value: "18"}
```

---

## 🏗️ Why Do We Need a Lexer?

Raw text is hard to process:

```text
SELECT name FROM users
```

Lexer transforms it into structured data:

```text
[SELECT][IDENT(name)][FROM][IDENT(users)]
```

👉 Now parser can understand structure.

---

## 📜 History (Why Lexers Exist)

Lexers appeared in early compilers (1960s–1970s) when programming languages like C and Pascal were being developed.

Key idea:

* Separate **text processing** from **syntax understanding**

Tools like `lex` (Unix) automated lexer generation.

---

## ⚙️ How a Lexer Works

A lexer reads input **character by character**:

```text
S → SE → SEL → SELE → SELECT
```

Then decides:

```text
"SELECT" → keyword token
```

---

## 🧠 Core Responsibilities

A lexer must:

1. Skip whitespace
2. Recognize symbols (`,`, `;`, `>`)
3. Read identifiers (`users`, `name`)
4. Read numbers (`123`)
5. Read strings (`'John'`)
6. Detect errors

---

## 🧪 Real Go Example

```go
func (l *Lexer) NextToken() Token {
    l.skipWhitespace()

    if l.pos >= len(l.input) {
        return Token{Type: EOF}
    }

    ch := l.input[l.pos]

    switch ch {
    case ',':
        l.pos++
        return Token{Type: COMMA, Value: ","}

    case '>':
        l.pos++
        return Token{Type: GT, Value: ">"}

    case '\'':
        return l.readString()
    }

    if isLetter(ch) {
        return l.readIdentifier()
    }

    if isDigit(ch) {
        return l.readNumber()
    }

    l.pos++
    return Token{Type: ILLEGAL, Value: string(ch)}
}
```

---

## 🔍 Real Problem Solving

### Problem 1: Parsing User Query

User input:

```text
age > 18
```

Lexer output:

```text
[IDENT(age)][GT][NUMBER(18)]
```

👉 Now parser can evaluate condition.

---

### Problem 2: Extracting Strings

Input:

```text
name = 'John'
```

Lexer must return:

```text
[IDENT(name)][EQ][STRING(John)]
```

---

### Problem 3: Ignoring Spaces

Input:

```text
SELECT   name
```

Lexer must treat it same as:

```text
SELECT name
```

---

## ⚠️ Common Mistakes

### ❌ 1. No bounds checking

```go
ch := input[pos] // can panic
```

### ❌ 2. Mixing token type with value

```go
GT = ">" // bad design
```

### ❌ 3. Not skipping whitespace

### ❌ 4. Infinite loop in string parsing

---

## ✅ Best Practices

✔ Keep token types semantic (`GT`, not `">"`)
✔ Store real value separately
✔ Use helper functions (`readIdentifier`, `readNumber`)
✔ Always check bounds
✔ Add `EOF` token

---

## 🧠 Mental Model

Think of lexer as:

```text
Scanner → Classifier
```

It scans characters and classifies them.

---

## 🚀 Where Lexers Are Used

* Compilers (Go, C, Rust)
* Databases (SQL parsing)
* Interpreters
* Config parsers (JSON, YAML)
* Search engines

---

## 🔥 What Comes Next?

After lexer:

👉 Build a **Parser**
👉 Parser creates **AST**

---

## 🧩 Summary

* Lexer = converts text → tokens
* Tokens = structured pieces
* Parser = understands structure
* AST = final representation

---

## 💡 Final Thought

A good lexer is:

* Simple
* Predictable
* Fast
* Easy to extend

---

If you understand this, you’ve already entered **compiler engineering** 🚀
