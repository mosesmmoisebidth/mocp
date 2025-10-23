; mocp_installer.iss
; Inno Setup 6 script for Mocp (mocp.exe)
; Place mocp.exe, send_receive.ico and TERMS.txt in the same folder as this .iss

[Setup]
AppName=Mocp
AppVersion=1.0
AppPublisher=Moses Mucyo
AppPublisherURL=https://moses.it.com
DefaultDirName={pf}\Mocp
DefaultGroupName=Mocp
OutputBaseFilename=mocp_installer
Compression=lzma2/normal
SolidCompression=yes
ArchitecturesInstallIn64BitMode=x64
WizardStyle=modern
; Default to per-user install; user can select All Users on the scope page.
PrivilegesRequired=lowest
LicenseFile=TERMS.txt
ChangesEnvironment=yes

[Files]
Source: "mocp.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "send_receive.ico"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\Mocp"; Filename: "{app}\mocp.exe"; IconFilename: "{app}\send_receive.ico"
Name: "{commondesktop}\Mocp"; Filename: "{app}\mocp.exe"; IconFilename: "{app}\send_receive.ico"; Tasks: desktopicon

[Tasks]
Name: "addtopath"; Description: "Add Mocp to PATH (allows running from command line)"; GroupDescription: "Additional tasks:"; Flags: unchecked
Name: "desktopicon"; Description: "Create a Desktop icon"; Flags: unchecked

[Run]
Filename: "{app}\mocp.exe"; Description: "Launch Mocp"; Flags: nowait postinstall skipifsilent

[Code]
// Use String for lParam (Inno Pascal Script doesn't support PChar)
function SendMessageTimeout(hWnd: LongWord; Msg: LongWord; wParam: LongWord;
  lParam: String; fuFlags: LongWord; uTimeout: LongWord; var lpdwResult: LongWord): LongWord;
  external 'SendMessageTimeoutW@user32.dll stdcall';

// Numeric literals used instead of re-declaring constants that may already exist
// HWND_BROADCAST = $FFFF
// WM_SETTINGCHANGE = $001A
// SMTO_ABORTIFHUNG = $0002

var
  ScopePage: TInputOptionWizardPage;

// Create a page allowing user to pick Install for all users vs only this user
procedure InitializeWizard();
begin
  ScopePage := CreateInputOptionPage(
    wpWelcome,
    'Installation scope',
    'Choose installation scope',
    'Choose whether to install Mocp for all users (system-wide) or only for the current user.',
    True, False);
  ScopePage.Add('Install for all users (requires administrator privileges)');
  ScopePage.Add('Install only for me (no administrator privileges required)');
  // Default to per-user install
  ScopePage.Values[1] := True;
end;

// When leaving the Scope page update the Destination directory shown to the user
function NextButtonClick(CurPageID: Integer): Boolean;
begin
  Result := True;
  if CurPageID = ScopePage.ID then
  begin
    if ScopePage.Values[0] then
      WizardForm.DirEdit.Text := ExpandConstant('{pf}\Mocp')
    else
      WizardForm.DirEdit.Text := ExpandConstant('{userappdata}\Programs\Mocp');
  end;
end;

// Adds {app} to PATH. If AllUsers requested and running elevated, write to HKLM; otherwise HKCU.
procedure AddToPath(AllUsers: Boolean);
var
  RootKey: Integer;
  KeyName: String;
  Existing: String;
  NewValue: String;
  Res: LongWord;
begin
  if AllUsers and IsAdminLoggedOn then
  begin
    RootKey := HKEY_LOCAL_MACHINE;
    KeyName := 'SYSTEM\CurrentControlSet\Control\Session Manager\Environment';
  end
  else
  begin
    RootKey := HKEY_CURRENT_USER;
    KeyName := 'Environment';
  end;

  if not RegQueryStringValue(RootKey, KeyName, 'Path', Existing) then
    Existing := '';

  NewValue := ExpandConstant('{app}');
  if (Existing = '') or (Pos(UpperCase(NewValue), UpperCase(Existing)) = 0) then
  begin
    if Existing <> '' then
      Existing := Existing + ';' + NewValue
    else
      Existing := NewValue;

    if RegWriteStringValue(RootKey, KeyName, 'Path', Existing) then
    begin
      // Notify other processes to reload environment variables
      // Use numeric literal for HWND_BROADCAST and flags
      SendMessageTimeout($FFFF, $001A, 0, 'Environment', $0002, 5000, Res);
    end
    else
    begin
      // Inform the user; HKLM write likely failed due to lack of elevation
      MsgBox('Failed to update PATH. To add Mocp system-wide to PATH, run the installer as Administrator or add it manually.', mbInformation, MB_OK);
    end;
  end;
end;

// After installation - apply PATH change if user selected that task
procedure CurStepChanged(CurStep: TSetupStep);
begin
  if CurStep = ssPostInstall then
  begin
    if IsTaskSelected('addtopath') then
      AddToPath(ScopePage.Values[0]);
  end;
end;

// Slightly customize "Ready to Install" memo text
function UpdateReadyMemo(Space, NewLine, MemoUserInfoInfo, MemoDirInfo, MemoTypeInfo,
  MemoComponentsInfo, MemoGroupInfo, MemoTasksInfo: String): String;
begin
  Result := 'Ready to install Mocp with these settings:' + NewLine +
            NewLine +
            'Install directory:' + NewLine + Space + WizardForm.DirEdit.Text + NewLine + NewLine +
            'Selected tasks:' + NewLine;
  if IsTaskSelected('desktopicon') then
    Result := Result + Space + '- Create Desktop icon' + NewLine;
  if IsTaskSelected('addtopath') then
    Result := Result + Space + '- Add Mocp to PATH' + NewLine;
end;
