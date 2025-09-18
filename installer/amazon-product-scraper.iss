; Amazon Product Scraper Installer
; Generated for building a Windows setup package using Inno Setup

[Setup]
AppName=Amazon Product Intelligence Suite
AppVersion=1.0.0
AppPublisher=Amazon Intelligence Labs
DefaultDirName={pf64}\AmazonProductIntelligence
DefaultGroupName=Amazon Product Intelligence
DisableProgramGroupPage=yes
UninstallDisplayIcon={app}\amazon-product-scraper.exe
OutputDir=.
OutputBaseFilename=amazon-product-intelligence-setup
; Provide an .ico file path if you want a custom installer icon.
; SetupIconFile=..\assets\amazon-product-intelligence.ico
Compression=lzma
SolidCompression=yes

[Files]
Source: "..\amazon-product-scraper.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\README.md"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\Amazon Product Intelligence"; Filename: "{app}\amazon-product-scraper.exe"
Name: "{commondesktop}\Amazon Product Intelligence"; Filename: "{app}\amazon-product-scraper.exe"; Tasks: desktopicon

[Tasks]
Name: "desktopicon"; Description: "Create a &desktop shortcut"; GroupDescription: "Additional icons:"; Flags: unchecked

[Run]
Filename: "{app}\amazon-product-scraper.exe"; Description: "Launch Amazon Product Intelligence"; Flags: nowait postinstall skipifsilent
