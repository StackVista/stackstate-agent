---
theme: gaia
_class: lead
paginate: true
backgroundColor: #fff
backgroundImage: url('https://marp.app/assets/hero-background.svg')
---

# **Black-box testing framework**

---

# Black-box testing

Examines functionalities without poking into the product.

Testers know what the software is supposed to do but is not aware of how it does it.

---

# Goal: end to end verification

What is needed to verify that an integration works

- check
- stackpack

---

# Goal: end to end verification (2)

What is needed to verify that an integration works

- check
- stackpack

## tests characteristics

- independent
- isolated
- reproducible

---

# Requirements

- **simplicity**
  - same experience like you would run unit tests
  - bootstrap integration use cases without prior knowledge
- **reusability**
  - different variations in SUT
  - explore how integration works in certain condition not automated yet

---

# Anatomy of a test scenario

**SUT**:
```
  Infrastructure
        ↑
   Application
```

---

# Anatomy of a test scenario (2)

**SUT**:
```
  Infrastructure      Infrastructure   ...
        ↑                   ↑
   Application         Application     ...
```

Expectation → SUT

---

<!-- _class: lead -->

# **Beest** (bee testing)

---

# **Beest** (bee testing)

(SUT) Yard:
```
  (Infrastructure) Hive   Hive
                    ↑      ↑
   (Application)   Bee    Bee
```

Expectation → Yard

---

# Beest Test cycle

```
   Hive
    ↑  
   Bee 
    ↑
Expectation
```

---

# Beest Test cycle (2)

```
   Hive                (provisioning) Create    Destroy
    ↑                                   ↓          ↑
   Bee                      (deploy) Prepare    Cleanup
    ↑                                   ↓          ↑
Expectation                               Verify
```

---

# How Beest does it

OS tools:
- Hive - Terraform module
- Bee - Ansible roles
- Expectation - Pytest
- CLI - Cobra
- some Go glue code

---

# Kubernetes check tests

Yard:
- Hive: eks, vpc, security groups
- Hive: vm, vpc, security groups
- Bee: receiver
- Bee: node/cluster agent, kubemetrics, sample workload

existing Pytest

---

<!-- _class: lead -->

# Demo

---

<!-- _class: lead -->

# Next:
# Kubernetes Stackpack tests
