#define LicenseStorageSubDir "RankBeam"
#define LicenseFileName "license.json"

; RankBeam Installer with Paystack Subscription Activation
; Generated for building a Windows setup package using Inno Setup

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
Filename: "{app}\rankbeam.exe"; Description: "Launch RankBeam"; Parameters: "/licensekey={code:GetGeneratedLicenseKey}"; Flags: nowait postinstall skipifsilent; Check: ActivationCompletedSuccessfully

[Code]
var
  CustomerInfoPage: TInputQueryWizardPage;
  GeneratedLicenseKey: string;
  ActivationFailed: Boolean;

function LicenseStoragePath(): string;
begin
  Result := ExpandConstant('{localappdata}') + '\\' + '{#LicenseStorageSubDir}' + '\\' + '{#LicenseFileName}';
end;

function EscapeJson(const Value: string): string;
var
  I: Integer;
  Ch: Char;
begin
  Result := '';
  for I := 1 to Length(Value) do
  begin
    Ch := Value[I];
    case Ch of
      '"': Result := Result + '\\"';
      '\\': Result := Result + '\\\\';
      Chr(8): Result := Result + '\\b';
      Chr(9): Result := Result + '\\t';
      Chr(10): Result := Result + '\\n';
      Chr(13): Result := Result + '\\r';
    else
      Result := Result + Ch;
    end;
  end;
end;

function NormaliseEmail(const Value: string): string;
begin
  Result := LowerCase(Trim(Value));
end;

function GetCustomerEmail(): string;
begin
  Result := NormaliseEmail(CustomerInfoPage.Values[0]);
end;

function GetLicenseKey(): string;
begin
  Result := Trim(CustomerInfoPage.Values[1]);
end;

procedure PersistLicenseDetails(const Email, Key: string);
var
  StoragePath, StorageDir, Payload: string;
begin
  if Trim(Key) = '' then
    RaiseException('License key is required.');
  if Trim(Email) = '' then
    RaiseException('Subscription email is required.');

  StoragePath := LicenseStoragePath();
  StorageDir := ExtractFilePath(StoragePath);
  if not DirExists(StorageDir) then
    if not ForceDirectories(StorageDir) then
      RaiseException('Unable to create directory for license data at ' + StorageDir);

  Payload := '{' + #13#10 +
    '  "licenseKey": "' + EscapeJson(Trim(Key)) + '",' + #13#10 +
    '  "email": "' + EscapeJson(NormaliseEmail(Email)) + '"' + #13#10 +
    '}' + #13#10;

  if not SaveStringToFile(StoragePath, Payload, False) then
    RaiseException('Unable to write license details to ' + StoragePath);

  Log('Stored license details at ' + StoragePath);
end;

procedure ShowLicenseSummary(const Email, Key: string);
begin
  MsgBox('Installation complete! Your Paystack subscription email is recorded as ' + NormaliseEmail(Email) + '.' + #13#10#13#10 +
    'Your license key:' + #13#10#13#10 + Key + #13#10#13#10 +
    'These details have been saved to ' + LicenseStoragePath() + '.', mbInformation, MB_OK);
end;

procedure InitializeWizard;
begin
  CustomerInfoPage := CreateInputQueryPage(wpUserInfo, 'License Activation', 'Enter your subscription details',
    'Provide the email address tied to your Paystack subscription and the active license key that was emailed to you.');
  CustomerInfoPage.Add('Subscription email:', False);
  CustomerInfoPage.Add('License key:', True);
end;

function NextButtonClick(CurPageID: Integer): Boolean;
begin
  Result := True;
  if CurPageID = CustomerInfoPage.ID then
  begin
    if GetCustomerEmail() = '' then
    begin
      MsgBox('Enter the email address linked to your Paystack subscription before continuing.', mbError, MB_OK);
      Result := False;
      Exit;
    end;
    if GetLicenseKey() = '' then
    begin
      MsgBox('Enter the license key from your activation email before continuing.', mbError, MB_OK);
      Result := False;
      Exit;
    end;
  end;
end;

procedure CurStepChanged(CurStep: TSetupStep);
var
  Email, Key: string;
begin
  if CurStep = ssPostInstall then
  begin
    Email := GetCustomerEmail();
    Key := GetLicenseKey();
    try
      PersistLicenseDetails(Email, Key);
      ShowLicenseSummary(Email, Key);
      GeneratedLicenseKey := Key;
      ActivationFailed := False;
    except
      ActivationFailed := True;
      GeneratedLicenseKey := '';
      SuppressibleMsgBox('Saving your license failed:' + #13#10#13#10 + GetExceptionMessage + #13#10#13#10 +
        'You can rerun the installer after resolving the issue.', mbError, MB_OK, IDOK);
    end;
  end;
end;

function ActivationCompletedSuccessfully(): Boolean;
begin
  Result := (not ActivationFailed) and (GeneratedLicenseKey <> '');
end;

function GetGeneratedLicenseKey(Param: string): string;
begin
  Result := GeneratedLicenseKey;
end;

function NeedRestart(): Boolean;
begin
  Result := False;
end;
