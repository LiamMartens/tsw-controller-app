Write-Host "Building version: $env:VERSION"
$env:CGO_CFLAGS="-I$PWD/vcpkg_installed/x64-windows/include"
$env:CGO_LDFLAGS="-L$PWD/vcpkg_installed/x64-windows/lib"
wails build -ldflags "-X 'main.VERSION=$env:VERSION' -X 'main.AXIOM_TOKEN=$env:AXIOM_TOKEN'"
