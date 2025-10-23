<!-- PROJECT LOGO -->
<p align="center">
  <img src="https://raw.githubusercontent.com/mosesmmoisebidth/mocp/send_receive.png" alt="MOCP Logo" width="120">
</p>

<h1 align="center">🛰️ MOCP — File Transfer Made Simple</h1>

<p align="center">
  <strong>Lightweight, Fast, and Reliable File Transfer Tool built with Go</strong><br>
  <a href="https://moses.it.com">moses.it.com</a> · 
  <a href="https://github.com/mosesmmoisebidth/mocp/issues">Report Bug</a> · 
  <a href="https://github.com/mosesmmoisebidth/mocp/pulls">Request Feature</a>
</p>

---

## 📖 About MOCP

**MOCP** is a cross-platform command-line tool designed for fast and simple file transfers.  
It enables users to **send** and **receive** files across networks with minimal setup — no configuration files, no dependencies, and no servers to install.

Written entirely in **Go (Golang)**, it compiles into a single self-contained executable that runs on **Windows**, **Linux**, and **macOS**.

---

## ✨ Features

- ⚡ **Fast and lightweight** — built using Go’s efficient networking.
- 🔒 **Reliable & Secure** — optional encryption and checksum verification.
- 🌍 **Cross-platform** — works on Windows, macOS, and Linux.
- 🧩 **Two simple modes**:
  - `mocp transfer` — send files.
  - `mocp receive` — receive files.
- 🆘 **Help command:** `mocp /?` or `mocp --help`
- 🪟 Optional **Windows Installer** with PATH integration and UI wizard.

---

## 🛠️ Installation

### 🪟 Windows (via Installer)

1. Download the latest **`mocp_installer.exe`** from the [Releases](https://github.com/mosesmmoisebidth/mocp/releases) page.  
2. Run the installer.  
3. Accept the license terms and select whether to install for all users or just for your account.  
4. (Optional) Check **“Add MOCP to PATH”** so you can run it from any terminal.  
5. Launch `mocp` from **Command Prompt** or **PowerShell**.

```bash
mocp /?
go install github.com/mosesmmoisebidth/mocp@latest
```
Linux / macOS (Build or Install with Go)
Install using 
```
go install
go install github.com/mosesmmoisebidth/mocp@latest

``
