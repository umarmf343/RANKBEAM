; RankBeam Installer
; Basic Windows setup package using Inno Setup

[Setup]
AppName=RankBeam
AppVersion=1.0.0
AppPublisher=Amazon Intelligence Labs
AppId={{A5E0D1E7-8F2E-4A83-8369-726F94F97884}}
DefaultDirName={commonpf64}\RankBeam
DefaultGroupName=RankBeam
DisableProgramGroupPage=yes
UninstallDisplayIcon={app}\rankbeam.exe
OutputDir=.
OutputBaseFilename=rankbeam-setup
Compression=lzma
SolidCompression=yes
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible

[Files]
Source: "..\bin\rankbeam.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\README.md"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\RankBeam"; Filename: "{app}\rankbeam.exe"
Name: "{commondesktop}\RankBeam"; Filename: "{app}\rankbeam.exe"; Tasks: desktopicon

[Tasks]
Name: "desktopicon"; Description: "Create a &desktop shortcut"; GroupDescription: "Additional icons:"; Flags: unchecked

[Run]
Filename: "{app}\rankbeam.exe"; Description: "Launch RankBeam"; Flags: nowait postinstall skipifsilent
