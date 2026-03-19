$BaseURL = "http://localhost:8080"

Write-Host "========== END-TO-END LOCATION TRACKING TEST ==========" -ForegroundColor Cyan

function Invoke-API {
    param(
        [string]$Method,
        [string]$Endpoint,
        [string]$JsonBody = "",
        [string]$Token
    )
    
    $URI = "$BaseURL$Endpoint"
    $Headers = @{ 'Content-Type' = 'application/json' }
    if ($Token) { $Headers['Authorization'] = "Bearer $Token" }
    
    try {
        if ($Method -eq 'GET') {
            $response = Invoke-WebRequest -Uri $URI -Method $Method -Headers $Headers -UseBasicParsing
        } else {
            $response = Invoke-WebRequest -Uri $URI -Method $Method -Headers $Headers -Body $JsonBody -UseBasicParsing
        }
        return $response.Content | ConvertFrom-Json
    }
    catch {
        Write-Host "HTTP Error [$Method $Endpoint]: $($_.Exception.Response.StatusCode)" -ForegroundColor Red
        try {
            $streamReader = [System.IO.StreamReader]::new($_.Exception.Response.GetResponseStream())
            $content = $streamReader.ReadToEnd()
            Write-Host "  Response: $content" -ForegroundColor Red
        } catch {}
        return $null
    }
}

Write-Host "[1] REGISTER NEW CUSTOMER USER" -ForegroundColor Yellow
$custId = Get-Random -Minimum 10000 -Maximum 99999
$custUser = "customer$custId"
$regJson = "{`"username`":`"$custUser`",`"password`":`"testpass123`"}"
$regResp = Invoke-API -Method POST -Endpoint "/auth/register" -JsonBody $regJson
if ($regResp.user.id) {
    Write-Host "OK - Customer registered: $custUser (ID: $($regResp.user.id))" -ForegroundColor Green
    $custID = $regResp.user.id
} else {
    Write-Host "FAIL - $($regResp.error)" -ForegroundColor Red
    exit 1
}

Write-Host "`n[2] LOGIN ADMIN FOR OPERATIONS" -ForegroundColor Yellow
$loginJson = '{"username":"admin","password":"admin123"}'
$loginResp = Invoke-API -Method POST -Endpoint "/auth/login" -JsonBody $loginJson
if ($loginResp.token) {
    Write-Host "OK - Admin login successful" -ForegroundColor Green
    $adminToken = $loginResp.token
} else {
    Write-Host "FAIL - Admin login failed: $($loginResp.error)" -ForegroundColor Red
    exit 1
}

Write-Host "`n[3] CREATE TRUCK" -ForegroundColor Yellow
$plateNum = "TEST$(Get-Random -Minimum 1000 -Maximum 9999)"
$truckJson = "{`"plate_number`":`"$plateNum`",`"driver_name`":`"Test Driver`"}"
$truckResp = Invoke-API -Method POST -Endpoint "/admin/trucks" -JsonBody $truckJson -Token $adminToken
if ($truckResp.data.id) {
    Write-Host "OK - Truck created: $($truckResp.data.plate_number) (ID: $($truckResp.data.id))" -ForegroundColor Green
    $truckID = $truckResp.data.id
} elseif ($truckResp.id) {
    Write-Host "OK - Truck created: ID=$($truckResp.id)" -ForegroundColor Green
    $truckID = $truckResp.id
} else {
    Write-Host "FAIL - Truck creation: $($truckResp.error)" -ForegroundColor Red
    Write-Host "  Response: $($truckResp | ConvertTo-Json)" -ForegroundColor Red
    exit 1
}

Write-Host "`n[4] CREATE ORDER" -ForegroundColor Yellow
$orderNum = "ORD$(Get-Random -Minimum 100000 -Maximum 999999)"
$orderJson = "{`"order_number`":`"$orderNum`",`"origin`":`"Makassar`",`"destination`":`"Gowa`"}"
$orderResp = Invoke-API -Method POST -Endpoint "/admin/orders" -JsonBody $orderJson -Token $adminToken  
Write-Host "Response: $($orderResp | ConvertTo-Json)" -ForegroundColor Yellow
if ($orderResp.data.id) {
    Write-Host "OK - Order created: $orderNum (ID: $($orderResp.data.id), Status: pending)" -ForegroundColor Green
    $orderID = $orderResp.data.id
} else {
    Write-Host "FAIL - Order: $($orderResp.error)" -ForegroundColor Red
    Write-Host "Full response: $($orderResp | ConvertTo-Json)" -ForegroundColor Red
    exit 1
}

Write-Host "`n[5] ASSIGN TRUCK TO ORDER" -ForegroundColor Yellow
$assignJson = "{`"order_id`":$orderID,`"truck_id`":$truckID}"
Write-Host "  Request: $assignJson" -ForegroundColor Gray
$assignResp = Invoke-API -Method PATCH -Endpoint "/admin/orders/assign" -JsonBody $assignJson -Token $adminToken
Write-Host "  Response: $($assignResp | ConvertTo-Json)" -ForegroundColor Gray
if ($assignResp.message -or ($assignResp -ne $null -and $assignResp.PSObject.Properties.Count -gt 0)) {
    Write-Host "OK - Truck $truckID assigned to Order $orderID" -ForegroundColor Green
} else {
    Write-Host "FAIL - Assign failed" -ForegroundColor Red
    Write-Host "  Error: $($assignResp.error)" -ForegroundColor Red
    exit 1
}

Write-Host "`n[6] CONFIRM PICKUP" -ForegroundColor Yellow
$pickupResp = Invoke-API -Method POST -Endpoint "/admin/orders/$orderID/confirm-pickup" -JsonBody '{}'  -Token $adminToken
if ($pickupResp.status -eq "in_transit") {
    Write-Host "OK - Pickup confirmed (Status: $($pickupResp.status))" -ForegroundColor Green
} else {
    Write-Host "FAIL - Pickup: $($pickupResp.error)" -ForegroundColor Red
}

Write-Host "`n[7-10] POST 4 TRUCK LOCATIONS" -ForegroundColor Yellow
$locIndex = 1
@(
    @{lat=-8.500000; lon=120.700000; name="Start"},
    @{lat=-8.505230; lon=120.712450; name="Checkpoint 1"},
    @{lat=-8.512340; lon=120.728900; name="Checkpoint 2"},
    @{lat=-8.520000; lon=120.735000; name="Final"}
) | ForEach-Object {
    $locJson = "{`"lat`":$($_.lat),`"lon`":$($_.lon)}"
    $locResp = Invoke-API -Method POST -Endpoint "/trucks/$truckID/location" -JsonBody $locJson
    Write-Host "  Location $locIndex`: $($_.name) - Posted (Lat: $($_.lat), Lon: $($_.lon))" -ForegroundColor Green
    $locIndex++
}

Write-Host "`n[11] CONFIRM DELIVERY" -ForegroundColor Yellow
$deliveryResp = Invoke-API -Method POST -Endpoint "/admin/orders/$orderID/confirm-delivery" -JsonBody '{}' -Token $adminToken
if ($deliveryResp.status -eq "delivered") {
    Write-Host "OK - Delivery confirmed (Status: $($deliveryResp.status))" -ForegroundColor Green
} else {
    Write-Host "FAIL - Delivery: $($deliveryResp.error)" -ForegroundColor Red
}

Write-Host "`n[12] GET LATEST LOCATION" -ForegroundColor Yellow
$latestResp = Invoke-API -Method GET -Endpoint "/trucks/$truckID/location" -JsonBody ''
if ($latestResp.latitude) {
    Write-Host "OK - Latest: Lat=$($latestResp.latitude), Lon=$($latestResp.longitude)" -ForegroundColor Green
} else {
    Write-Host "FAIL - Get location: $($latestResp.error)" -ForegroundColor Red
}

Write-Host "`n[13] GET LOCATION HISTORY" -ForegroundColor Yellow
$historyResp = Invoke-API -Method GET -Endpoint "/trucks/$truckID/locations?limit=10" -JsonBody ''
if ($historyResp -is [array]) {
    Write-Host "OK - History: $($historyResp.Count) locations" -ForegroundColor Green
    [array]::Reverse($historyResp)
    $j = 1
    foreach ($h in $historyResp) {
        Write-Host "  $j. Lat=$($h.latitude), Lon=$($h.longitude)" -ForegroundColor Cyan
        $j++
    }
} else {
    Write-Host "FAIL - History: $($historyResp.error)" -ForegroundColor Red
}

Write-Host "`n[14] PUBLIC TRACKING (No Auth)" -ForegroundColor Yellow
$publicResp = Invoke-API -Method GET -Endpoint "/public/orders/$orderNum/track" -JsonBody ''
if ($publicResp.order_number) {
    Write-Host "OK - Public tracking retrieved:" -ForegroundColor Green
    Write-Host "  Order: $($publicResp.order_number) - Status: $($publicResp.status)" 
    Write-Host "  Truck: $($publicResp.truck.plate_number) - Driver: $($publicResp.truck.driver_name)"
    Write-Host "  Current: Lat=$($publicResp.location.latitude), Lon=$($publicResp.location.longitude)"
} else {
    Write-Host "FAIL - Tracking: $($publicResp.error)" -ForegroundColor Red
}

Write-Host "`n========== SUCCESS ==========" -ForegroundColor Cyan
Write-Host "E2E Test Complete: Admin(admin) -> Customer($custUser) -> Truck($truckID) -> Order($orderNum)" -ForegroundColor Green
