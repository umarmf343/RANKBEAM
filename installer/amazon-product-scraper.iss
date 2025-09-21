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
Source: "..\bin\amazon-product-scraper.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\bin\license-seeder.exe"; DestDir: "{tmp}"; Flags: deleteafterinstall
Source: "..\README.md"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\Amazon Product Intelligence"; Filename: "{app}\amazon-product-scraper.exe"
Name: "{commondesktop}\Amazon Product Intelligence"; Filename: "{app}\amazon-product-scraper.exe"; Tasks: desktopicon

[Tasks]
Name: "desktopicon"; Description: "Create a &desktop shortcut"; GroupDescription: "Additional icons:"; Flags: unchecked

[Run]
Filename: "{tmp}\license-seeder.exe"; Parameters: "--api-base=\"{code:GetLicenseServerUrl}\" --customer=\"{code:GetCustomerId}\" --output \"{app}\license-key.txt\""; StatusMsg: "Issuing per-machine license"; Flags: runhidden waituntilterminated
Filename: "{app}\amazon-product-scraper.exe"; Description: "Launch Amazon Product Intelligence"; Flags: nowait postinstall skipifsilent

[Code]
var
  LicensePage: TInputQueryWizardPage;

function GetCustomerId(Param: string): string;
begin
  Result := Trim(LicensePage.Values[0]);
end;

function GetLicenseServerUrl(Param: string): string;
begin
  // Replace with your production license server endpoint.
  Result := 'http://localhost:8080';
end;

procedure InitializeWizard();
begin
  LicensePage := CreateInputQueryPage(wpSelectTasks,
    'License activation',
    'Issue a machine-bound license',
    'Provide the email address or order number used during purchase. The installer will contact the licensing server and bind the license to this computer.');
  LicensePage.Add('Customer email / order #:', False);
end;

function NextButtonClick(CurPageID: Integer): Boolean;
begin
  Result := True;
  if (CurPageID = LicensePage.ID) and (Trim(LicensePage.Values[0]) = '') then
  begin
    MsgBox('Please provide the customer email or order number to issue a license.', mbError, MB_OK);
    Result := False;
  end;
end;

procedure CurStepChanged(CurStep: TSetupStep);
var
  LicensePath, LicenseKey: string;
begin
  if CurStep = ssPostInstall then
  begin
    LicensePath := ExpandConstant('{app}\license-key.txt');
    if LoadStringFromFile(LicensePath, LicenseKey) then
    begin
      MsgBox('Installer activation complete. Your license key is:'#13#10#13#10 + Trim(LicenseKey) +
        #13#10#13#10'TIP: The desktop app stores the key under your profile''s configuration directory and validates it on each launch.',
        mbInformation, MB_OK);
    end;
  end;
end;
