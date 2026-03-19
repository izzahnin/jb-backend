$BaseURL = "http://localhost:8080"

Write-Host "========== END-TO-END LOCATION TRACKING TEST ==========" -ForegroundColor Cyan

function Invoke-API {
    param(
        [string]$Method,
        [string]$Endpoint,
        [hashtable]$Body,
        [string]$Token
    )
    
    $URI = "$BaseURL$Endpoint"
    $Headers = @{ 'Content-Type' = 'application/json' }
    if ($Token) { $Headers['Authorization'] = "Bearer $Token" }
    
    try {
        $response = Invoke-WebRequest -Uri $URI -Method $Method -Headers $Headers -Body ($Body | ConvertTo-Json) -UseBasicParsing
        return $response.Content | ConvertFrom-Json
    }
    catch {
        Write-Host "HTTP Error for $Method $Endpoint : $($_.Exception.Message)" -ForegroundColor Red
        if ($_.Exception.Response) {
            try {
                $streamReader = [System.IO.StreamReader]::new($_.Exception.Response.GetResponseStream())
                $respContent = $streamReader.ReadToEnd()
                Write-Host "Response Body: $respContent" -ForegroundColor Red
            } catch {}
        }
        return $null
    }
}

Write-Host "[1] REGISTER NEW CUSTOMER USER" -ForegroundColor Yellow
$timestamp = Get-Date -Format "yyyyMMddHHmmss"
$customerId = Get-Random -Minimum 10000 -Maximum 99999
$customerUsername = "customer$customerId"
$regResp = Invoke-API -Method POST -Endpoint "/auth/register" -Body @{ username = $customerUsername; password = "testpass123" }
if ($regResp.user.id) {
    Write-Host "OK - Customer registered: $customerUsername (ID: $($regResp.user.id))" -ForegroundColor Green
    $customerUserID = $regResp.user.id
    $customerToken = $regResp.token
} else {
    Write-Host "FAIL - $($regResp.error)" -ForegroundColor Red
    exit 1
}

Write-Host "`n[2] LOGIN ADMIN FOR OPERATIONS" -ForegroundColor Yellow
$loginResp = Invoke-API -Method POST -Endpoint "/auth/login" -Body @{ username = "admin"; password = "admin123" }
if ($loginResp.token) {
    Write-Host "OK - Admin login successful" -ForegroundColor Green
    $adminToken = $loginResp.token
    $adminUserID = $loginResp.user.id
} else {
    Write-Host "FAIL - Admin login failed" -ForegroundColor Red
    exit 1
}

Write-Host "`n[3] CREATE TRUCK" -ForegroundColor Yellow
$plateNum = "TEST$(Get-Random -Minimum 1000 -Maximum 9999)"
$truckResp = Invoke-API -Method POST -Endpoint "/admin/trucks" -Body @{ plate_number = $plateNum; driver_name = "Budiman Test Driver" } -Token $adminToken
if ($truckResp.id) {
    Write-Host "OK - Truck created: $($truckResp.plate_number) (ID: $($truckResp.id))" -ForegroundColor Green
    $truckID = $truckResp.id
} else {
    Write-Host "FAIL - $($truckResp.error)" -ForegroundColor Red
    exit 1
}

Write-Host "`n[4] CREATE ORDER (for customer)" -ForegroundColor Yellow
$orderNum = "ORD$(Get-Random -Minimum 100000 -Maximum 999999)"
$orderResp = Invoke-API -Method POST -Endpoint "/admin/orders" -Body @{ order_number = $orderNum; origin = "Makassar City"; destination = "Gowa Harbor"; customer_id = $customerUserID } -Token $adminToken
if ($orderResp.id) {
    Write-Host "OK - Order created: $orderNum (ID: $($orderResp.id), Status: $($orderResp.status))" -ForegroundColor Green
    $orderID = $orderResp.id
} else {
    Write-Host "FAIL - $($orderResp.error)" -ForegroundColor Red
    exit 1
}

Write-Host "`n[5] ASSIGN TRUCK TO ORDER" -ForegroundColor Yellow
$assignResp = Invoke-API -Method PATCH -Endpoint "/admin/orders/assign" -Body @{ order_id = $orderID; truck_id = $truckID } -Token $adminToken
if ($assignResp.id) {
    Write-Host "OK - Truck assigned successfully" -ForegroundColor Green
} else {
    Write-Host "FAIL - Assignment error" -ForegroundColor Red
    exit 1
}

Write-Host "`n[6] CONFIRM PICKUP" -ForegroundColor Yellow
$pickupResp = Invoke-API -Method POST -Endpoint "/admin/orders/$orderID/confirm-pickup" -Body @{} -Token $adminToken
if ($pickupResp.status -eq "pickup") {
    Write-Host "OK - Pickup confirmed (Status: $($pickupResp.status))" -ForegroundColor Green
} else {
    Write-Host "FAIL - Pickup error" -ForegroundColor Red
}

Write-Host "`n[7-10] POST 4 TRUCK LOCATIONS" -ForegroundColor Yellow
$locations = @(
    @{ lat = -8.500000; lon = 120.700000; name = "Start Point" },
    @{ lat = -8.505230; lon = 120.712450; name = "Checkpoint 1" },
    @{ lat = -8.512340; lon = 120.728900; name = "Checkpoint 2" },
    @{ lat = -8.520000; lon = 120.735000; name = "Final Point" }
)

$locCount = 0
foreach ($loc in $locations) {
    $locCount++
    $locResp = Invoke-API -Method POST -Endpoint "/trucks/$truckID/location" -Body @{ latitude = $loc.lat; longitude = $loc.lon }
    Write-Host "  Location $locCount posted: $($loc.name) (Lat: $($loc.lat), Lon: $($loc.lon))" -ForegroundColor Green
}

Write-Host "`n[11] CONFIRM DELIVERY" -ForegroundColor Yellow
$deliveryResp = Invoke-API -Method POST -Endpoint "/admin/orders/$orderID/confirm-delivery" -Body @{} -Token $adminToken
if ($deliveryResp.status -eq "delivered") {
    Write-Host "OK - Delivery confirmed (Status: $($deliveryResp.status))" -ForegroundColor Green
} else {
    Write-Host "FAIL - Delivery error" -ForegroundColor Red
}

Write-Host "`n[12] GET LATEST LOCATION" -ForegroundColor Yellow
$latestResp = Invoke-API -Method GET -Endpoint "/trucks/$truckID/location" -Body @{}
if ($latestResp.latitude) {
    Write-Host "OK - Latest location: Lat=$($latestResp.latitude), Lon=$($latestResp.longitude)" -ForegroundColor Green
} else {
    Write-Host "FAIL - Could not get location" -ForegroundColor Red
}

Write-Host "`n[13] GET LOCATION HISTORY" -ForegroundColor Yellow
$historyResp = Invoke-API -Method GET -Endpoint "/trucks/$truckID/locations?limit=10" -Body @{}
if ($historyResp -is [array]) {
    Write-Host "OK - Location history: $($historyResp.Count) locations stored" -ForegroundColor Green
    $i = 1
    foreach ($h in $historyResp) {
        Write-Host "  $i. Lat=$($h.latitude), Lon=$($h.longitude)" -ForegroundColor Cyan
        $i++
    }
} else {
    Write-Host "FAIL - Could not get history" -ForegroundColor Red
}

Write-Host "`n[14] CUSTOMER PUBLIC TRACKING (No Auth Required)" -ForegroundColor Yellow
$trackingResp = Invoke-API -Method GET -Endpoint "/public/orders/$orderNum/track" -Body @{}
if ($trackingResp.order_number) {
    Write-Host "OK - Public tracking retrieved:" -ForegroundColor Green
    Write-Host "  Order: $($trackingResp.order_number)"
    Write-Host "  Status: $($trackingResp.status)"
    Write-Host "  Truck: $($trackingResp.truck.plate_number) (Driver: $($trackingResp.truck.driver_name))"
    Write-Host "  Location: Lat=$($trackingResp.location.latitude), Lon=$($trackingResp.location.longitude)"
} else {
    Write-Host "FAIL - Could not get tracking data" -ForegroundColor Red
}

Write-Host "`n========== TEST COMPLETE ==========" -ForegroundColor Cyan
Write-Host "Summary: Admin User=admin, Customer=$customerUsername, Truck=$truckID, Order=$orderNum" -ForegroundColor Green
