# 🧱 WAL Design (Write-Ahead Logging)

## 🤔 Why do I even need WAL?

When I started building my own database, everything looked simple:

* insert data
* update data
* delete data

But then a real problem appeared:

> 💥 What happens if my program crashes in the middle of a write?

Example:

* User inserts data
* Database writes half of it
* Suddenly: power off / crash

Now the database is:

* ❌ corrupted
* ❌ inconsistent
* ❌ partially written

This is where WAL becomes necessary.

---

## 💡 What problem WAL solves

WAL solves one critical problem:

> "How do I make sure my data is not lost or corrupted after a crash?"

Instead of writing directly to the database, WAL changes the flow:

1. First → write operation to WAL (safe log)
2. Then → apply change to database

So even if crash happens:

👉 I still have the operation saved in WAL
👉 I can replay it later

---

## 🧠 Why I am using WAL in my database

Because I want my database to behave like real systems (PostgreSQL, MySQL):

* survive crashes
* recover state
* guarantee durability

Without WAL, my database is just a "toy".

With WAL → it becomes a **real system**.

---

## 🔥 Real-life analogy

Think about banking system:

* You transfer money
* Bank writes transaction into a log first
* Then updates your balance

If system crashes:

👉 They replay the log and restore your balance correctly

WAL = that transaction history

---

## 🧩 How it works in my database

Instead of doing this:

❌ WRONG:

* write data directly to table

I do this:

✅ CORRECT:

* create WAL record
* save it to file
* then apply change

---

## ⚠️ Example problem WAL fixes

### Case: INSERT without WAL

* insert user "John"
* crash happens during write

Result:

* half-written row
* database corrupted

---

### Case: INSERT with WAL

* write "insert John" to WAL
* crash happens

After restart:

* read WAL
* re-apply insert

Result:

✅ data is restored correctly

---

## 🧠 Important insight I learned

At first I thought:

> "I can just store values directly"

But that’s wrong.

The correct way is:

> "Store operations, not just data"

Because operations tell the database **how to rebuild state**.

---

## 🚨 Mistake I faced

While building this, I made a serious mistake:

* WAL was calling catalog
* catalog was calling WAL

👉 infinite recursion → stack overflow

This taught me:

> WAL must ONLY log — never execute logic

---

## ⚡ Why WAL is powerful

Because it allows:

* crash recovery
* debugging history of operations
* future support for transactions

---

## 💬 Final thoughts

WAL is not just a feature.

It is the reason databases are reliable.

Without it:

* crashes = data loss

With it:

* crashes = recoverable events

---

> Building WAL made me understand how real databases guarantee safety.
