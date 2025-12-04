$ErrorActionPreference = "Stop"

param(
  [string]$ServiceName = "dis-agent",
  [string]$DisplayName = "DIS Agent",
  [string]$BinaryPath = "C:\Program Files\DIS Agent\dis-agent.exe",
  [string]$ConfigPath = "C:\ProgramData\DIS Agent\agent.yaml"
)

if (!(Test-Path $BinaryPath)) { Write-Error "Binary not found at $BinaryPath"; exit 1 }

sc.exe create $ServiceName binPath= "\"$BinaryPath\"" start= auto DisplayName= "$DisplayName"
sc.exe description $ServiceName "DIS Agent"
sc.exe failure $ServiceName reset= 30 actions= restart/5000
sc.exe start $ServiceName
