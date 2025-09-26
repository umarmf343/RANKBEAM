#define LicenseApiBaseUrl "https://rankbeam.hannyshive.com.ng"
#define LicenseApiToken "F6BFD62E2CD91CED258005CBDE1FED2423DBD8775F2430A75F882CDF3ADC6750"
#define LicenseStorageSubDir "RankBeam"
#define LicenseFileName "license.key"

; RankBeam Installer with License Activation
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
Source: "..\bin\fingerprint-helper.exe"; DestDir: "{tmp}"; Flags: ignoreversion deleteafterinstall
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

function ExtractJsonValue(const Json, Key: string): string;
var
  Pattern, Tail: string;
  ColonPos, QuotePos, EndPos: Integer;
begin
  Result := '';
  Pattern := '"' + Key + '"';
  QuotePos := Pos(Pattern, Json);
  if QuotePos = 0 then
    Exit;
  Tail := Copy(Json, QuotePos + Length(Pattern), MaxInt);
  ColonPos := Pos(':', Tail);
  if ColonPos = 0 then
    Exit;
  Tail := Trim(Copy(Tail, ColonPos + 1, MaxInt));
  if (Tail = '') or (Tail[1] <> '"') then
    Exit;
  Tail := Copy(Tail, 2, MaxInt);
  EndPos := Pos('"', Tail);
  if EndPos = 0 then
    Exit;
  Result := Copy(Tail, 1, EndPos - 1);
end;

function GetCustomerEmail(): string;
begin
  Result := Trim(CustomerInfoPage.Values[0]);
end;

function GetMachineFingerprint(): string;
var
  ResultCode: Integer;
  OutputPath: string;
  Output: AnsiString;
begin
  OutputPath := ExpandConstant('{tmp}') + '\\fingerprint.out';
  if FileExists(OutputPath) then
    DeleteFile(OutputPath);

  if not Exec(ExpandConstant('{tmp}') + '\\fingerprint-helper.exe', '--output ' + AddQuotes(OutputPath), '', SW_HIDE,
    ewWaitUntilTerminated, ResultCode) then
  begin
    RaiseException('Unable to start fingerprint helper.');
  end;
  if ResultCode <> 0 then
    RaiseException('Fingerprint helper exited with code ' + IntToStr(ResultCode) + '.');

  if not LoadStringFromFile(OutputPath, Output) then
    RaiseException('Failed to read fingerprint output.');

  Result := Trim(Output);
  if Result = '' then
    RaiseException('Fingerprint helper returned an empty fingerprint.');

  Log(Format('Derived machine fingerprint %s', [Result]));
end;

function RequestLicenseFromServer(const Fingerprint: string): string;
var
  WinHttpReq: Variant;
  Url, Payload, Email: string;
  Status: Integer;
begin
  Email := GetCustomerEmail();
  if Email = '' then
    RaiseException('Email address is required for license activation.');

  Url := '{#LicenseApiBaseUrl}/api/v1/licenses';
  Payload := '{"customerId":"' + EscapeJson(Email) + '","fingerprint":"' + EscapeJson(Fingerprint) + '"}';

  WinHttpReq := CreateOleObject('WinHttp.WinHttpRequest.5.1');
  WinHttpReq.Open('POST', Url, False);
  WinHttpReq.Option[9] := 2048; // WINHTTP_FLAG_SECURE_PROTOCOL_TLS1_2
  WinHttpReq.SetRequestHeader('Content-Type', 'application/json');
  if '{#LicenseApiToken}' <> '' then
    WinHttpReq.SetRequestHeader('X-Installer-Token', '{#LicenseApiToken}');
  WinHttpReq.Send(Payload);

  Status := WinHttpReq.Status;
  Log(Format('License server responded with %d', [Status]));
  if (Status <> 200) and (Status <> 201) then
    RaiseException(Format('License request failed (%d): %s', [Status, WinHttpReq.ResponseText]));

  Result := WinHttpReq.ResponseText;
end;

procedure PersistLicenseKey(const Key: string);
var
  StoragePath, StorageDir: string;
begin
  StoragePath := LicenseStoragePath();
  StorageDir := ExtractFilePath(StoragePath);
  if not DirExists(StorageDir) then
  begin
    if not ForceDirectories(StorageDir) then
      RaiseException('Unable to create directory for license key at ' + StorageDir);
  end;
  if not SaveStringToFile(StoragePath, Key, False) then
    RaiseException('Unable to write license key to ' + StoragePath);
  Log('Stored license key at ' + StoragePath);
end;

procedure ShowLicenseKey(const Key: string);
begin
  MsgBox('Installation complete! Your license key is:\n\n' + Key + '\n\nIt has been saved automatically to ' + LicenseStoragePath() +
    '. Keep a copy for your records.', mbInformation, MB_OK);
end;

function ActivateLicense(): string;
var
  Fingerprint, Response, Key: string;
begin
  Fingerprint := GetMachineFingerprint();
  Response := RequestLicenseFromServer(Fingerprint);
  Key := ExtractJsonValue(Response, 'licenseKey');
  if Key = '' then
    RaiseException('License server response did not include a license key.');
  PersistLicenseKey(Key);
  ShowLicenseKey(Key);
  Result := Key;
end;

procedure InitializeWizard;
begin
  CustomerInfoPage := CreateInputQueryPage(wpUserInfo, 'License Activation', 'Enter your customer details',
    'Provide the email address or order identifier used at purchase. It will be used to issue your license key.');
  CustomerInfoPage.Add('Email address:', False);
end;

procedure CurStepChanged(CurStep: TSetupStep);
begin
  if CurStep = ssPostInstall then
  begin
    try
      GeneratedLicenseKey := ActivateLicense();
      ActivationFailed := False;
    except
      ActivationFailed := True;
      GeneratedLicenseKey := '';
      SuppressibleMsgBox('License activation failed:\n\n' + GetExceptionMessage + '\n\nYou can rerun the installer or contact support to complete activation.',
        mbError, MB_OK, IDOK);
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
