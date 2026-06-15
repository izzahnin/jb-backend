function Invoke-CurlJson {
    param($Method, $Url, $Body, $Token)
    $headers = @("-H", "Content-Type: application/json")
    if ($Token) { $headers += "-H", "Authorization: Bearer $Token" }
    
    # Use -d as a single string to avoid quoting issues
    $cmdParams = @("-sS", "-i", "-X", $Method, $Url) + $headers
    if ($Body) {
        $jsonStr = $Body | ConvertTo-Json -Compress
        $cmdWithBody = $cmdParams + @("--data-binary", "@-")
        $raw = $jsonStr | & curl.exe @cmdWithBody 2>&1
    } else {
        $raw = & curl.exe @cmdParams 2>&1
    }
    $exitCode = $LASTEXITCODE
    $statusLine = $raw | Select-Object -First 1
    $status = if ($statusLine -match "HTTP/\d\.\d\s+(\d+)") { $matches[1] } else { "000" }
    
    $rawStr = $raw -join "`n"
    $parts = $rawStr -split "\r?\n\r?\n", 2
    $json = if ($parts.Count -gt 1) { $parts[1] | ConvertFrom-Json -ErrorAction SilentlyContinue } else { $null }
    
    return [PSCustomObject]@{ Status = [int]$status; Data = $json; Raw = $rawStr; ExitCode = $exitCode }
}

$baseUrl = "http://localhost:8080"
$ts = Get-Date -Format "HHmmss"

Write-Host "--- 1. Setup / Login ---"
$res = Invoke-CurlJson -Method POST -Url "$baseUrl/admin/setup" -Body @{ username="superadmin"; password="password123"; full_name="Super Admin" }
if ($res.Status -eq 409 -or $res.Data.error -match "already exists") {
    Write-Host "Admin already setup or exists. Logging in..."
    $res = Invoke-CurlJson -Method POST -Url "$baseUrl/auth/login" -Body @{ username="superadmin"; password="password123" }
}
$superToken = $res.Data.token
if (-not $superToken) { 
    Write-Host "Failed to get superadmin token. Status: $($res.Status)"
    if ($res.Data) { $res.Data | ConvertTo-Json }
    if ($res.Raw) { Write-Host $res.Raw }
    exit 
}
Write-Host "SuperAdmin Status: $($res.Status), Token: $($superToken.Substring(0, [Math]::Min(10, $superToken.Length)))..."

Write-Host "`n--- 2. Manage Users ---"
$users = Invoke-CurlJson -Method GET -Url "$baseUrl/admin/users" -Token $superToken
Write-Host "Get Users Status: $($users.Status)"

$salesUser = "sales_$ts"
$opsUser = "ops_$ts"
$resSales = Invoke-CurlJson -Method POST -Url "$baseUrl/admin/users" -Body @{ username=$salesUser; password="password123"; role="admin_sales"; full_name="Sales Admin" } -Token $superToken
$resOps = Invoke-CurlJson -Method POST -Url "$baseUrl/admin/users" -Body @{ username=$opsUser; password="password123"; role="admin_ops"; full_name="Ops Admin" } -Token $superToken
Write-Host "Create Sales ($salesUser): $($resSales.Status), Create Ops ($opsUser): $($resOps.Status)"

Write-Host "`n--- 3. Login as Sales/Ops ---"
$salesRes = Invoke-CurlJson -Method POST -Url "$baseUrl/auth/login" -Body @{ username=$salesUser; password="password123" }
$salesToken = $salesRes.Data.token
$opsRes = Invoke-CurlJson -Method POST -Url "$baseUrl/auth/login" -Body @{ username=$opsUser; password="password123" }
$opsToken = $opsRes.Data.token
Write-Host "Sales Token Status: $($salesRes.Status), Ops Token Status: $($opsRes.Status)"

Write-Host "`n--- 4. Master Data (SuperAdmin) ---"
$cust = Invoke-CurlJson -Method POST -Url "$baseUrl/admin/customers" -Body @{ company_name="PT Test $ts"; pic_name="PIC $ts"; phone="+62812$ts"; email="pic$ts@test.co.id"; address="Jl Test $ts"; npwp="01.234.567.8-901.000" } -Token $superToken
$custId = $cust.Data.data.id
$driver = Invoke-CurlJson -Method POST -Url "$baseUrl/admin/drivers" -Body @{ name="Driver $ts"; license_number="SIM-$ts"; phone="+62811$ts"; status="available"; is_active=$true } -Token $superToken
$driverId = $driver.Data.data.id
$truck = Invoke-CurlJson -Method POST -Url "$baseUrl/admin/trucks" -Body @{ plate_number="B-$ts-XY"; truck_type="Fuso Box"; status="available"; is_active=$true } -Token $superToken
$truckId = $truck.Data.data.id
Write-Host "Created Cust: $custId ($($cust.Status)), Driver: $driverId ($($driver.Status)), Truck: $truckId ($($truck.Status))"

Write-Host "`n--- 5. Orders (Sales Admin) ---"
$orderBody = @{
    customer_id = [int]$custId
    origin = "Origin $ts"
    destination = "Dest $ts"
    total_containers = 1
}
$order = Invoke-CurlJson -Method POST -Url "$baseUrl/admin/orders" -Body $orderBody -Token $salesToken
$orderId = $order.Data.data.id
Write-Host "Order Created: $orderId, Status: $($order.Status)"

$orderGet = Invoke-CurlJson -Method GET -Url "$baseUrl/admin/orders/$orderId" -Token $salesToken
$orderUpdate = Invoke-CurlJson -Method PATCH -Url "$baseUrl/admin/orders/$orderId" -Body @{ status="partial" } -Token $salesToken
Write-Host "Get Order: $($orderGet.Status), Update Status: $($orderUpdate.Status)"

Write-Host "`n--- 6. Trips (Ops Admin) ---"
$tripBody = @{
    order_id = [int]$orderId
    driver_id = [int]$driverId
    truck_id = [int]$truckId
}
$trip = Invoke-CurlJson -Method POST -Url "$baseUrl/admin/trips" -Body $tripBody -Token $opsToken
$tripId = $trip.Data.data.id
Write-Host "Trip Created: $tripId, Status: $($trip.Status)"

$tripsByOrder = Invoke-CurlJson -Method GET -Url "$baseUrl/admin/orders/$orderId/trips" -Token $opsToken
Write-Host "Trips by Order Status: $($tripsByOrder.Status)"

$startTrip = Invoke-CurlJson -Method PATCH -Url "$baseUrl/admin/trips/$tripId/start" -Body @{ container_number="CONT-$ts"; seal_number="SEAL-$ts" } -Token $opsToken
$loc = Invoke-CurlJson -Method POST -Url "$baseUrl/trips/$tripId/location" -Body @{ lat=-6.2; lon=106.816666 } -Token $opsToken
$latestLoc = Invoke-CurlJson -Method GET -Url "$baseUrl/trips/$tripId/location" -Token $opsToken
$historyLoc = Invoke-CurlJson -Method GET -Url "$baseUrl/trips/$tripId/locations?limit=5" -Token $opsToken
$completeTrip = Invoke-CurlJson -Method PATCH -Url "$baseUrl/admin/trips/$tripId/deliver" -Token $opsToken
Write-Host "Start: $($startTrip.Status), Loc: $($loc.Status), Complete: $($completeTrip.Status)"

Write-Host "`n--- 7. Tracking & Cleanup ---"
$track = Invoke-CurlJson -Method GET -Url "$baseUrl/public/orders/$($order.Data.data.order_number)/track"
Write-Host "Public Track Status: $($track.Status)"

$deactivate = Invoke-CurlJson -Method DELETE -Url "$baseUrl/admin/trucks/$truckId" -Token $superToken
Write-Host "Deactivate Truck Status: $($deactivate.Status)"

$logout = Invoke-CurlJson -Method POST -Url "$baseUrl/auth/logout" -Token $superToken
Write-Host "Logout Status: $($logout.Status)"
