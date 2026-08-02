;; =============================================================
;; BLX — Instalador (Inno Setup 6 ou 7)
;; Como compilar:
;;   1. Instale o Inno Setup 6 ou 7 (https://jrsoftware.org/isdl.php)
;;   2. Abra este arquivo e pressione F9 (ou: Compile > Compile)
;;   3. O instalador é gerado em output\BLX-Setup.exe
;; =============================================================

#define MyAppName "BLX"
#define MyAppVersion "1.0.0"
#define MyAppPublisher "BLX"
#define MyAppExeName "BLX.exe"

[Setup]
AppId={{B7C1E0F2-9D4A-4E6B-8F3C-5A1D2E3F4A5B}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppVerName={#MyAppName} {#MyAppVersion}
AppPublisher={#MyAppPublisher}
AppComments=Gerador de comandos de pentest para profissionais autorizados
DefaultDirName={autopf}\{#MyAppName}
DefaultGroupName={#MyAppName}
OutputDir=output
OutputBaseFilename=BLX-Setup
SetupIconFile=assets\icon.ico
WizardImageFile=assets\wizard-large.png
WizardSmallImageFile=assets\wizard-small.png
WizardStyle=modern
LicenseFile=LICENSE.txt
Compression=lzma2
SolidCompression=yes
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
PrivilegesRequired=admin
DisableProgramGroupPage=yes
ShowLanguageDialog=no
SetupLogging=yes
UninstallDisplayIcon={app}\{#MyAppExeName}
VersionInfoVersion={#MyAppVersion}
VersionInfoDescription=Gerador de comandos de pentest

[Languages]
Name: "portuguesebrazilian"; MessagesFile: "compiler:Languages\BrazilianPortuguese.isl"

[Tasks]
Name: "desktopicon"; Description: "Criar atalho na Área de Trabalho"; GroupDescription: "Ícones:"; Flags: unchecked

[Files]
Source: "bin\blx.exe"; DestDir: "{app}"; DestName: "{#MyAppExeName}"; Flags: ignoreversion

[Icons]
Name: "{group}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"
Name: "{group}\{#MyAppName} (uninstall)"; Filename: "{uninstallexe}"
Name: "{autodesktop}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; Tasks: desktopicon

[Run]
Filename: "{app}\{#MyAppExeName}"; Description: "Iniciar {#MyAppName} agora"; Flags: nowait postinstall skipifsilent

;; Opcional: liberar a porta 8080 no Firewall do Windows para acesso
;; a partir de OUTRAS máquinas da rede (descomente se necessário).
;; Atenção: só faça isso se a equipe acessar o BLX remotamente.
;[Run]
;Filename: "netsh.exe"; Parameters: "advfirewall firewall add rule name=""BLX 8080"" dir=in action=allow protocol=TCP localport=8080"; Flags: runhidden

[Code]
procedure InitializeWizard();
begin
  WizardForm.PageNameLabel.Caption := 'BLX';
end;
