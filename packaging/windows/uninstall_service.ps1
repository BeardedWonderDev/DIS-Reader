$ErrorActionPreference = "Stop"
param([string]$ServiceName = "dis-agent")
sc.exe stop $ServiceName 2>$null
sc.exe delete $ServiceName 2>$null
Write-Output "Service $ServiceName removed (configs left intact)."
