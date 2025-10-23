Build and use the Inno Setup installer for Mocp

Prerequisites
- Inno Setup installed (https://jrsoftware.org/isinfo.php)
- The ISCC compiler in your PATH (Inno Setup's command-line compiler)

Files
- `mocp_installer.iss` - Inno Setup script
- `TERMS.txt` - Terms shown during installation
- `mocp.exe` - The application executable to be bundled
- `send_receive.ico` - Icon file used for shortcuts

Build
From the repository root, run:

ISCC mocp_installer.iss

This produces `mocp_installer.exe` in the output folder configured by Inno Setup (usually the same directory).

Install
Run the produced `mocp_installer.exe` and follow the installer wizard. Options include:
- Install for all users (requires admin) or only for current user
- Add Mocp to PATH (checkbox)
- Create desktop shortcut

Notes
- Adding to PATH modifies the system/user registry and broadcasts the environment change.
- If you want silent installs, adjust `[Run]` and flags in the script accordingly.
