# 🐳 MiniContainer — Go দিয়ে তৈরি Minimal Container Runtime

> Docker-এর মতো কাজ করে এমন একটি ছোট্ট Container Runtime, যা Go ও Linux Kernel Namespaces ব্যবহার করে তৈরি।

---

## 📌 প্রজেক্ট পরিচিতি

এই প্রজেক্টটি একটি **শিক্ষামূলক Minimal Container Runtime**, যেখানে দেখানো হয়েছে কীভাবে Docker আসলে ভেতরে ভেতরে কাজ করে। কোনো third-party লাইব্রেরি ছাড়া, শুধু Go এবং Linux Kernel-এর built-in features ব্যবহার করে একটি isolated container environment তৈরি করা হয়েছে।

**মূল কনসেপ্টগুলো:**
- Linux Namespaces (UTS, PID, Mount, Network)
- OverlayFS (Copy-on-Write filesystem)
- Chroot (filesystem isolation)
- Process Isolation

---

## 🏗️ আর্কিটেকচার — কীভাবে কাজ করে?

```
User Command
    │
    ▼
┌─────────────────────────────────────┐
│          Parent Process             │
│  (run() function)                   │
│                                     │
│  /proc/self/exe কে "child" arg      │
│  দিয়ে নতুন Namespaces-এ spawn করে  │
└──────────────┬──────────────────────┘
               │  CLONE_NEWUTS
               │  CLONE_NEWPID
               │  CLONE_NEWNS
               │  CLONE_NEWNET
               ▼
┌─────────────────────────────────────┐
│          Child Process              │
│  (child() function)                 │
│                                     │
│  1. OverlayFS Mount                 │
│  2. Chroot → mergedDir              │
│  3. /proc Mount                     │
│  4. Hostname Set                    │
│  5. Command Execute                 │
└─────────────────────────────────────┘
```

---

## 📂 ফাইল স্ট্রাকচার

```
mini-container/
├── main.go                      # মূল কোড
├── alpine-rootfs/               # Container এর Base Image (Alpine Linux rootfs)
│   ├── bin/
│   ├── etc/
│   ├── lib/
│   └── ...
└── containers/
    └── container-<timestamp>/   # প্রতিটি Container এর আলাদা ফোল্ডার
        ├── upper/               # Container এর নিজের write layer (নতুন ফাইল এখানে যায়)
        ├── work/                # OverlayFS এর internal কাজের জন্য
        └── merged/              # upper + lowerdir মিলিয়ে যা দেখা যায় (Container দেখে এটাই)
```

---

## 🔍 ধাপে ধাপে কোড ব্যাখ্যা

### ধাপ ১ — Parent Process: `run()`

```
User: sudo go run main.go run /bin/sh
         │
         └─→ run() চালু হয়
```

Parent process নিজেকেই আবার execute করে, কিন্তু এবার `child` argument দিয়ে এবং **নতুন Linux Namespaces** তৈরি করে:

| Namespace Flag | কাজ |
|---|---|
| `CLONE_NEWUTS` | Container এর নিজস্ব Hostname থাকবে |
| `CLONE_NEWPID` | Container এর ভেতরে PID 1 থেকে শুরু হবে |
| `CLONE_NEWNS` | Host এর filesystem আলাদা থাকবে |
| `CLONE_NEWNET` | Container এর নিজস্ব Network Stack থাকবে |

এই `Cloneflags` গুলোই হলো Docker Isolation এর মূল রহস্য।

---

### ধাপ ২ — Child Process: Unique Container ID তৈরি

```go
containerID := fmt.Sprintf("container-%d", time.Now().UnixNano())
```

প্রতিটি container-এর জন্য nanosecond timestamp ব্যবহার করে একটি unique ID তৈরি হয়। এতে একই সাথে একাধিক terminal থেকে container চালালেও কোনো conflict হয় না।

---

### ধাপ ৩ — OverlayFS: স্তরে স্তরে Filesystem

OverlayFS হলো Docker Image Layer-এর মূল প্রযুক্তি।

```
alpine-rootfs/  ←── lowerDir (Read-Only, Base Image)
      +
containers/<id>/upper/  ←── upperDir (Read-Write, Container এর নিজস্ব)
      +
containers/<id>/work/   ←── workDir (OverlayFS এর internal use)
      ║
      ▼
containers/<id>/merged/ ←── mergedDir (Container এটাই দেখে — সব মিলিয়ে)
```

**কীভাবে কাজ করে:**
- Container কোনো ফাইল **read** করলে → `lowerDir` (alpine-rootfs) থেকে আসে
- Container কোনো ফাইল **write/create** করলে → `upperDir`-এ যায়, `lowerDir` অক্ষত থাকে
- Container মুছে দিলে → শুধু `upperDir` ডিলিট হয়, base image নিরাপদ

এই প্রযুক্তিকে **Copy-on-Write (CoW)** বলে।

---

### ধাপ ৪ — Chroot: Container এর দুনিয়া আলাদা করা

```go
syscall.Chroot(mergedDir)
os.Chdir("/")
```

`Chroot` call করলে container process এর কাছে `mergedDir` হয়ে যায় `/` (root)। সে host এর বাকি filesystem দেখতেই পায় না — যেন সে সম্পূর্ণ আলাদা একটা লিনাক্স সিস্টেমে আছে।

---

### ধাপ ৫ — /proc Mount ও Hostname

```go
syscall.Mount("proc", "proc", "proc", 0, "")
syscall.Sethostname([]byte("my-isolated-container"))
```

- Container এর নিজস্ব `/proc` mount করা হয় — যাতে `ps aux` চালালে host এর process না দেখে
- Hostname আলাদা সেট করা হয় — container এর ভেতরে `hostname` চালালে `my-isolated-container` দেখাবে

---

### ধাপ ৬ — Command Execute ও Cleanup

```go
cmd := exec.Command(command, args[1:]...)
```

User এর দেওয়া command (যেমন `/bin/sh`) container environment এর ভেতরে চালানো হয়।

কাজ শেষে `defer` দিয়ে `/proc` unmount করা হয়।

---

## 🚀 কীভাবে চালাবে?

### Prerequisites

```bash
# Alpine Linux rootfs ডাউনলোড করতে হবে
mkdir alpine-rootfs
cd alpine-rootfs
curl -LO https://dl-cdn.alpinelinux.org/alpine/v3.19/releases/x86_64/alpine-minirootfs-3.19.0-x86_64.tar.gz
tar -xzf alpine-minirootfs-3.19.0-x86_64.tar.gz
rm alpine-minirootfs-3.19.0-x86_64.tar.gz
cd ..
```

### Container চালানো

```bash
# root permission দরকার (Namespace ও Mount এর জন্য)
sudo go run main.go run /bin/sh

# অথবা bash চালাতে
sudo go run main.go run /bin/bash
```

### Container এর ভেতরে

```sh
# এটা আলাদা hostname দেখাবে
hostname
# Output: my-isolated-container

# শুধু container এর process দেখাবে, host এর না
ps aux

# নতুন ফাইল তৈরি করলে সেটা upper/ তে যাবে
echo "hello from container" > /test.txt
```

---

## 🧪 OverlayFS যাচাই করা

Container চলার সময় আরেকটি terminal থেকে:

```bash
# Container এর upper layer দেখো — নতুন ফাইলগুলো এখানে আসবে
ls containers/

# নির্দিষ্ট container এর upper layer
ls containers/container-<timestamp>/upper/

# Base image অপরিবর্তিত আছে কিনা দেখো
ls alpine-rootfs/
```

---

## 🔒 নিরাপত্তা বিষয়ক নোট

এই প্রজেক্টটি **শুধুমাত্র শিক্ষামূলক উদ্দেশ্যে**। Production-grade container runtime এ আরও অনেক কিছু থাকে:

| Feature | এই প্রজেক্ট | Docker/runc |
|---|---|---|
| Namespace Isolation | ✅ আছে | ✅ আছে |
| OverlayFS | ✅ আছে | ✅ আছে |
| cgroups (Resource Limit) | ❌ নেই | ✅ আছে |
| Seccomp (Syscall Filter) | ❌ নেই | ✅ আছে |
| Capabilities Drop | ❌ নেই | ✅ আছে |
| User Namespace | ❌ নেই | ✅ আছে |
| Network Setup (veth) | ❌ নেই | ✅ আছে |

---

## 📚 এই প্রজেক্ট থেকে যা শেখা যায়

- Docker আসলে কীভাবে কাজ করে তার ভেতরের রহস্য
- Linux Namespace কী এবং কেন isolation দরকার
- OverlayFS / Copy-on-Write filesystem কীভাবে কাজ করে
- `runc` এবং OCI Runtime Spec এর ভিত্তি কী
- Go দিয়ে low-level Linux syscall কীভাবে করা যায়

---

## 🛠️ Tech Stack

- **Language:** Go (Golang)
- **OS:** Linux (Arch, Ubuntu — যেকোনো)
- **Kernel Features:** Namespaces, OverlayFS, Chroot, Mount
- **Packages:** `os`, `os/exec`, `syscall`, `path/filepath` (সব standard library)

---

## 📖 আরও জানতে

- [Linux Namespaces — man7.org](https://man7.org/linux/man-pages/man7/namespaces.7.html)
- [OverlayFS Documentation](https://www.kernel.org/doc/html/latest/filesystems/overlayfs.html)
- [OCI Runtime Specification](https://github.com/opencontainers/runtime-spec)
- [runc — Reference Container Runtime](https://github.com/opencontainers/runc)

---

*"Containers are not magic — they are just Linux features put together." 🐧*