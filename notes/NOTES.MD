## 1. The Transaction Wrapper (`execTx`)

The `execTx` method is a **Higher-Order Function**. It abstracts the complexity of managing a database connection lifecycle.

* **Responsibility:** It handles the "Begin," "Commit," and "Rollback" logic.
* **Safety:** If your business logic inside the function returns an `error`, `execTx` catches it and triggers a `Rollback`. This ensures that if the money leaves Account A but fails to arrive at Account B, the whole operation is undone.
* **Isolation:** By creating a new `Queries` object with `New(tx)`, you ensure all subsequent SQL commands are executed within the **same database pipe**.

---

## 2. The Transfer Logic (`TransferTx`)

A transfer isn't just one update; it’s a five-step process that must be **Atomic** (all or nothing).

1. **Create Transfer Record:** This is the "Receipt." It records the intent (Who sent how much to whom).
2. **Create From-Entry:** A ledger entry for the sender (negative amount).
3. **Create To-Entry:** A ledger entry for the receiver (positive amount).
4. **Update Sender Balance:** Subtract the amount from the sender’s account.
5. **Update Receiver Balance:** Add the amount to the receiver’s account.

> **Audit Trail Note:** We record `Entries` separately from the `Accounts` balance. This is a best practice in fintech. If the `balance` is ever in question, you can sum up all `Entries` for that account to verify the "source of truth."

---

## 3. The Deadlock Prevention Strategy

The most sophisticated part of your code is the comparison: `if arg.FromAccountID < arg.ToAccountID`.

### The Problem: Circular Wait

Without this check, if User 1 sends money to User 2, and User 2 sends money to User 1 at the same time:

* **Tx A** locks User 1.
* **Tx B** locks User 2.
* **Tx A** tries to lock User 2 (Blocked by Tx B).
* **Tx B** tries to lock User 1 (Blocked by Tx A).
* **Result:** The database engine kills one transaction to break the loop.

### The Solution: Resource Hierarchy

By comparing IDs, you force every transaction to **acquire locks in the same order**.

* Both Tx A and Tx B will try to lock the **smaller ID first**.
* If User 1 is the smaller ID, Tx B will wait for Tx A to finish with User 1 *before* it even tries to touch User 2.
* This turns a "Circular Wait" into a "Queue."

---

## 4. Helper Function: `addMoney`

This function is a clean way to handle the balance updates.

* **Optimization:** Instead of a `SELECT` followed by an `UPDATE`, it uses `AddAccountBalance` (which likely translates to `SET balance = balance + $1`).
* **Atomic Increment:** Doing the math inside the SQL statement (`SET balance = balance + amount`) is safer than doing it in Go because it avoids "lost updates" where two threads overwrite each other's math.

---

## 5. Summary of Best Practices Used

| Best Practice | Implementation in your code |
| --- | --- |
| **Atomicity** | Handled by `execTx` and SQL transactions. |
| **Auditability** | Creation of `Entry` records for every movement. |
| **Concurrency Control** | ID comparison to prevent deadlocks. |
| **Idempotency** | Using `TransferID` and `EntryID` to track unique events. |
| **Clean Code** | Separation of concerns between `Store`, `Queries`, and Transaction logic. |

---

### Potential Next Step