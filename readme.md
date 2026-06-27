# Mini Container Runtime (Go)

Go ভাষা ব্যবহার করে তৈরি একটি ছোট (Mini) Container Runtime। এই প্রজেক্টের মূল উদ্দেশ্য হলো Docker বা runc-এর মতো Container Runtime কীভাবে Linux Kernel-এর বিভিন্ন Feature ব্যবহার করে কাজ করে, তা হাতে-কলমে শেখা।

এই প্রজেক্টে Linux Namespace, OverlayFS, `chroot()`, Process Isolation এবং Alpine Linux RootFS ব্যবহার করে একটি সাধারণ Container Environment তৈরি করা হয়েছে।

> **শুধুমাত্র শেখার উদ্দেশ্যে তৈরি।**
> এটি Production Environment-এ ব্যবহারের জন্য নয়।

---

# ✨ কী কী Feature আছে?

* ✅ PID Namespace (আলাদা Process Tree)
* ✅ UTS Namespace (আলাদা Hostname)
* ✅ Mount Namespace
* ✅ Network Namespace
* ✅ OverlayFS ভিত্তিক Writable Filesystem
* ✅ Alpine Linux RootFS স্বয়ংক্রিয় Download
* ✅ প্রতিটি Container-এর জন্য আলাদা Writable Layer
* ✅ Container বন্ধ হলে Temporary File Cleanup
* ✅ `/proc` Mount
* ✅ Container-এর ভিতরে আলাদা Hostname

---

# 📂 Project Structure

```text
.
├── main.go
├── utils/
│   └── helpers.go
├── alpine-rootfs/
├── containers/
│   └── container-xxxxxxxx/
│       ├── upper/
│       ├── work/
│       └── merged/
```

---

# ⚙️ কীভাবে কাজ করে?

পুরো Process-এর Flow নিচে দেখানো হলো—

```text
Host Machine
      │
      │
      ├── Alpine RootFS Download (প্রথমবার)
      │
      ├── OverlayFS তৈরি
      │
      ├── নতুন Linux Namespace তৈরি
      │      ├── PID
      │      ├── UTS
      │      ├── Mount
      │      └── Network
      │
      ├── chroot()
      │
      ├── /proc Mount
      │
      └── Container-এর ভিতরে Command Execute
```

---

# 🛠️ প্রয়োজনীয় সফটওয়্যার

এই Project চালানোর জন্য নিচের জিনিসগুলো লাগবে—

* Linux
* Go 1.22 বা তার উপরের Version
* Root Privilege (`sudo`)
* OverlayFS Support

---

# 🔨 Build

```bash
go build -o mycontainer
```

---

# 🚀 ব্যবহার করার নিয়ম

Container-এর ভিতরে Shell চালাতে—

```bash
sudo ./mycontainer run /bin/sh
```

অথবা যেকোনো Command চালাতে—

```bash
sudo ./mycontainer run ls
```

আরও উদাহরণ—

```bash
sudo ./mycontainer run uname -a
```

---

# 📥 প্রথমবার চালানোর সময় কী হবে?

যদি `alpine-rootfs/` Folder না থাকে, তাহলে Runtime নিজেই—

* Alpine Linux Mini RootFS Download করবে
* Extract করবে
* পরবর্তী Container-গুলোর জন্য সেটি ব্যবহার করবে

অর্থাৎ RootFS একবারই Download হবে।

---

# 🗂️ Container Filesystem

প্রতিটি Container-এর জন্য আলাদা Writable Layer তৈরি হয়।

```text
containers/
└── container-xxxxxxxx/
    ├── upper/
    ├── work/
    └── merged/
```

Filesystem-এর ধারণা—

```text
Read Only Alpine RootFS
          +
Writable Layer
          =
Merged Filesystem
```

ফলে মূল RootFS কখনো পরিবর্তন হয় না।

---

# 🐧 Linux Kernel-এর যেসব Feature ব্যবহার করা হয়েছে

* Linux Namespace
* PID Namespace
* UTS Namespace
* Mount Namespace
* Network Namespace
* OverlayFS
* chroot()
* proc Filesystem
* Process Isolation
* Hostname Isolation

---

# ⚠️ বর্তমান সীমাবদ্ধতা

এই Project এখনো একটি Basic Implementation।

এখনো নিচের Feature-গুলো যোগ করা হয়নি—

* ❌ cgroups
* ❌ User Namespace
* ❌ pivot_root()
* ❌ Linux Capability Management
* ❌ Seccomp
* ❌ veth Pair
* ❌ Linux Bridge
* ❌ Internet Access
* ❌ OCI Image Support
* ❌ Container Image Management

---

# 🚧 ভবিষ্যৎ পরিকল্পনা

আগামীতে নিচের Feature-গুলো যোগ করার পরিকল্পনা আছে—

* [ ] `pivot_root()` ব্যবহার করা
* [ ] User Namespace যোগ করা
* [ ] cgroups v2 যুক্ত করা
* [ ] CPU Limit
* [ ] Memory Limit
* [ ] Loopback Interface স্বয়ংক্রিয়ভাবে চালু করা
* [ ] veth Pair তৈরি করা
* [ ] Linux Bridge Networking
* [ ] NAT Configuration
* [ ] Internet Access
* [ ] Linux Capability Drop
* [ ] Seccomp Filter
* [ ] OCI Image Support
* [ ] উন্নত CLI
* [ ] Logging System

---

# 🎯 এই Project থেকে কী শেখা যাবে?

এই Project-এর মাধ্যমে নিচের বিষয়গুলো বাস্তবে শেখা যাবে—

* Container কীভাবে কাজ করে
* Docker-এর ভিতরের মূল ধারণা
* Linux Namespace
* Process Isolation
* Filesystem Isolation
* OverlayFS
* Root Filesystem
* Container Lifecycle
* Linux Kernel Feature

---

# ⚠️ সতর্কতা

এই Project শুধুমাত্র শেখার জন্য তৈরি করা হয়েছে।

এটি Production Environment-এ ব্যবহার করা নিরাপদ নয়।

---

# 📜 License

MIT License
