## 🎯 What This Does

<!-- One sentence: user-facing value or technical improvement -->

---

## 📊 Visual Overview

```mermaid
graph LR
    A[Before] -->|Change| B[After]
    B --> C[Impact]

    style A fill:#30363d,stroke:#1f6feb,color:#c9d1d9
    style B fill:#1f6feb,stroke:#58a6ff,color:#ffffff
    style C fill:#238636,stroke:#2ea043,color:#ffffff
```

<!-- Replace with actual diagram showing:
     - Architecture changes: use architecture-beta
     - Data flow: use flowchart LR
     - State changes: use stateDiagram-v2
     - API interactions: use sequenceDiagram

     Dark theme colors (GitHub defaults):
     - Neutral: fill:#30363d,stroke:#1f6feb,color:#c9d1d9
     - Primary: fill:#1f6feb,stroke:#58a6ff,color:#ffffff
     - Success: fill:#238636,stroke:#2ea043,color:#ffffff
     - Warning: fill:#9e6a03,stroke:#d29922,color:#ffffff
     - Error: fill:#da3633,stroke:#f85149,color:#ffffff
-->

---

## 🔍 Details

### Changed Files
<!-- Keep to 3-5 most important files -->
- `path/to/file.go` - Brief description
- `path/to/test.go` - Added tests

### Technical Notes
<!-- Only if needed - link to issue/doc for deep dive -->

---

## ✅ Verification

```mermaid
flowchart TD
    Build[Build] --> Tests[Tests]
    Tests --> Manual[Manual Check]
    Manual --> Done[✓ Ready]

    style Build fill:#30363d,stroke:#1f6feb,color:#c9d1d9
    style Tests fill:#30363d,stroke:#1f6feb,color:#c9d1d9
    style Manual fill:#30363d,stroke:#1f6feb,color:#c9d1d9
    style Done fill:#238636,stroke:#2ea043,color:#ffffff
```

**Quick Test:**
```bash
# What to run
make test

# Expected output
✓ All tests pass
```

---

## 📋 Checklist

- [ ] Tests pass
- [ ] Linter clean
- [ ] No security issues

---

**Related:** Refs #issue
