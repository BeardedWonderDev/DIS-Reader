$ErrorActionPreference = "Stop"

param(
  [string]$ServiceName   = "dis-agent",
  [string]$DisplayName   = "DIS Agent",
  [string]$InstallDir    = "C:\Program Files\DIS Agent",
  [string]$BinaryName    = "dis-agent.exe",
  [string]$ConfigPath    = "C:\ProgramData\DIS Agent\agent.yaml",
  [string]$CredsPath     = "C:\ProgramData\DIS Agent\bridge_agents.yaml",
  [string]$ServiceUser   = "NT AUTHORITY\\LocalService",
  [switch]$NoStart
)

$BinaryPath = Join-Path $InstallDir $BinaryName

# Ensure directories exist
$null = New-Item -ItemType Directory -Force -Path $InstallDir
$null = New-Item -ItemType Directory -Force -Path (Split-Path $ConfigPath)

# Copy binary and sample configs if present
if (Test-Path ".\dis-agent.exe") { Copy-Item ".\dis-agent.exe" $BinaryPath -Force }
if (Test-Path ".\agent.yaml")    { Copy-Item ".\agent.yaml" $ConfigPath -Force -ErrorAction SilentlyContinue }
if (Test-Path ".\bridge_agents.yaml") { Copy-Item ".\bridge_agents.yaml" $CredsPath -Force -ErrorAction SilentlyContinue }

if (!(Test-Path $BinaryPath)) { Write-Error "Binary not found at $BinaryPath"; exit 1 }

# Restrict ACL on config directory to service account and Administrators
try {
  $configDir = Split-Path $ConfigPath
  $acl = Get-Acl $configDir
  $acl.SetAccessRuleProtection($true, $false)
  $ruleService = New-Object System.Security.AccessControl.FileSystemAccessRule($ServiceUser,"Modify","ContainerInherit, ObjectInherit","None","Allow")
  $ruleAdmins  = New-Object System.Security.AccessControl.FileSystemAccessRule("BUILTIN\\Administrators","FullControl","ContainerInherit, ObjectInherit","None","Allow")
  $acl.SetAccessRule($ruleService)
  $acl.SetAccessRule($ruleAdmins)
  Set-Acl -Path $configDir -AclObject $acl
} catch {
  Write-Warning "ACL hardening failed: $_"
}

# Create service account? (uses LocalService by default)
$binQuoted = '"' + $BinaryPath + '"'
sc.exe create $ServiceName binPath= $binQuoted start= auto DisplayName= "$DisplayName" obj= "$ServiceUser"
sc.exe description $ServiceName "DIS Agent"
sc.exe failure $ServiceName reset= 30 actions= restart/5000

if (-not $NoStart) {
  sc.exe start $ServiceName
  Write-Output "Service $ServiceName started."
} else {
  Write-Output "Service $ServiceName installed (not started)."
}
