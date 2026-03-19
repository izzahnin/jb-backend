#!/usr/bin/env pwsh

# End-to-End Test Script for Jalur Berlian Backend
# Tests the complete flow: Register -> Login -> Truck -> Order -> Assign -> Pickup -> Locations -> Delivery -> Tracking

$BaseURL = "http://localhost:8080"
$timestamp = Get-Date -Format "yyyyMMddHHmmss"
$username = "testadmin$timestamp"

Write-Host "`n========== END-TO-END LOCATION TRACKING TEST ==========" -ForegroundColor Cyan

# Utility function to make API calls
function Invoke-API {
    param(
        [string]$Method,
        [string]$Endpoint,
        [hashtable]$Body,
        [string]$Token
    )
    
    $URI = "$BaseURL$Endpoint"
    $Headers = @{
        'Content-Type' = 'application/json'
    }
    if ($Token) {
        $Headers['Authorization'] = "Bearer $Token"
    }
    
    try {
        $response = Invoke-WebRequest -Uri $URI -Method $Method -Headers $Headers -Body ($Body | ConvertTo-Json) -UseBasicParsing
        return $response.Content | ConvertFrom-Json
    }
    catch {
        $response = $_.Exception.Response
        if ($response) {
            Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
            return $null
        }
    }
}

# Test 1: Register
Write-Host "`n[1] REGISTER NEW USER" -ForegroundColor Yellow
$registerBody = @{
    username = $username
    password = "testpass123"
    role = "admin"
}
$regResp = Invoke-API -Method POST -Endpoint "/auth/register" -Body $registerBody
if ($regResp.user.id) {
    Write-Host "[OK] Registered: $($regResp.user.username) (ID: $($regResp.user.id))" -ForegroundColor Green
    $userID = $regResp.user.id
    $token = $regResp.token
} else {
    Write-Host "[FAIL] Registration failed: $($regResp.error)" -ForegroundColor Red
    exit 1
}

# Test 2: Login (verify login works)
Write-Host "`n[2] LOGIN USER" -ForegroundColor Yellow
$loginBody = @{
    username = $username
    password = "testpass123"
}
$loginResp = Invoke-API -Method POST -Endpoint "/auth/login" -Body $loginBody
if ($loginResp.token) {
    Write-Host "✓ Login successful, token obtained" -ForegroundColor Green
} else {
    Write-Host "✗ Login failed" -ForegroundColor Red
    exit 1
}

# Test 3: Create Truck
Write-Host "`n[3] CREATE TRUCK" -ForegroundColor Yellow
$plateNum = "TEST$(Get-Random -Minimum 1000 -Maximum 9999)"
$truckBody = @{
    plate_number = $plateNum
    driver_name = "Budiman Test Driver"
}
$truckResp = Invoke-API -Method POST -Endpoint "/admin/trucks" -Body $truckBody -Token $token
if ($truckResp.id) {
    Write-Host "✓ Truck created: $($truckResp.plate_number) (ID: $($truckResp.id), Driver: $($truckResp.driver_name))" -ForegroundColor Green
    $truckID = $truckResp.id
} else {
    Write-Host "✗ Truck creation failed: $($truckResp.error)" -ForegroundColor Red
    exit 1
}

# Test 4: Create Order
Write-Host "`n[4] CREATE ORDER" -ForegroundColor Yellow
$orderNum = "ORD$(Get-Random -Minimum 100000 -Maximum 999999)"
$orderBody = @{
    order_number = $orderNum
    origin = "Makassar City Center"
    destination = "Gowa Harbor"
}
$orderResp = Invoke-API -Method POST -Endpoint "/admin/orders" -Body $orderBody -Token $token
if ($orderResp.id) {
    Write-Host "✓ Order created: $($orderResp.order_number) (ID: $($orderResp.id), Status: $($orderResp.status))" -ForegroundColor Green
    $orderID = $orderResp.id
} else {
    Write-Host "✗ Order creation failed: $($orderResp.error)" -ForegroundColor Red
    exit 1
}

# Test 5: Assign Truck to Order
Write-Host "`n[5] ASSIGN TRUCK TO ORDER" -ForegroundColor Yellow
$assignBody = @{
    order_id = $orderID
    truck_id = $truckID
}
$assignResp = Invoke-API -Method PATCH -Endpoint "/admin/orders/assign" -Body $assignBody -Token $token
if ($assignResp.message -or $assignResp.id) {
    Write-Host "✓ Truck assigned to order successfully" -ForegroundColor Green
} else {
    Write-Host "✗ Assignment failed" -ForegroundColor Red
    exit 1
}

# Test 6: Confirm Pickup
Write-Host "`n[6] CONFIRM PICKUP" -ForegroundColor Yellow
$pickupBody = @{}
$pickupResp = Invoke-API -Method POST -Endpoint "/admin/orders/$orderID/confirm-pickup" -Body $pickupBody -Token $token
if ($pickupResp.status -eq "pickup") {
    Write-Host "✓ Pickup confirmed (Status: $($pickupResp.status))" -ForegroundColor Green
} else {
    Write-Host "✗ Pickup confirmation failed" -ForegroundColor Red
    exit 1
}

# Test 7-10: Post Locations
Write-Host "`n[7-10] POST 4 TRUCK LOCATIONS" -ForegroundColor Yellow
$locations = @(
    @{lat = -8.5; lon = 120.7; name = "Start Point - City Center"},
    @{lat = -8.50523; lon = 120.71245; name = "Checkpoint 1 - Halfway"},
    @{lat = -8.51234; lon = 120.72890; name = "Checkpoint 2 - Almost There"},
    @{lat = -8.52000; lon = 120.73500; name = "Final Point - Harbor"}
)

$locationCount = 0
foreach ($loc in $locations) {
    $locationCount++
    $locBody = @{
        latitude = $loc.lat
        longitude = $loc.lon
    }
    $locResp = Invoke-API -Method POST -Endpoint "/trucks/$truckID/location" -Body $locBody
    if ($locResp.message -or $locResp.latitude) {
        Write-Host "✓ Location $locationCount posted: $($loc.name) (Lat: $($loc.lat), Lon: $($loc.lon))" -ForegroundColor Green
    } else {
        Write-Host "✗ Location $locationCount posting failed" -ForegroundColor Red
    }
}

# Test 11: Confirm Delivery
Write-Host "`n[11] CONFIRM DELIVERY" -ForegroundColor Yellow
$deliveryBody = @{}
$deliveryResp = Invoke-API -Method POST -Endpoint "/admin/orders/$orderID/confirm-delivery" -Body $deliveryBody -Token $token
if ($deliveryResp.status -eq "delivered") {
    Write-Host "✓ Delivery confirmed (Status: $($deliveryResp.status))" -ForegroundColor Green
} else {
    Write-Host "✗ Delivery confirmation failed" -ForegroundColor Red
    exit 1
}

# Test 12: Get Latest Location
Write-Host "`n[12] GET LATEST LOCATION" -ForegroundColor Yellow
$latestResp = Invoke-API -Method GET -Endpoint "/trucks/$truckID/location" -Body @{}
if ($latestResp.latitude) {
    Write-Host "✓ Latest location retrieved: Lat=$($latestResp.latitude), Lon=$($latestResp.longitude)" -ForegroundColor Green
} else {
    Write-Host "✗ Failed to retrieve latest location" -ForegroundColor Red
}

# Test 13: Get Location History
Write-Host "`n[13] GET LOCATION HISTORY" -ForegroundColor Yellow
$historyResp = Invoke-API -Method GET -Endpoint "/trucks/$truckID/locations?limit=10" -Body @{}
if ($historyResp -is [array]) {
    Write-Host "✓ Location history retrieved: $($historyResp.Count) locations" -ForegroundColor Green
    $i = 1
    foreach ($h in $historyResp) {
        Write-Host "  $i. Lat=$($h.latitude), Lon=$($h.longitude)" -ForegroundColor Cyan
        $i++
    }
} else {
    Write-Host "✗ Failed to retrieve location history" -ForegroundColor Red
}

# Test 14: Customer Public Tracking (No Auth Required)
Write-Host "`n[14] CUSTOMER PUBLIC TRACKING (No Auth)" -ForegroundColor Yellow
$trackingResp = Invoke-API -Method GET -Endpoint "/public/orders/$orderNum/track" -Body @{}
if ($trackingResp.order_number) {
    Write-Host "✓ Public tracking data retrieved:" -ForegroundColor Green
    Write-Host "  Order Number: $($trackingResp.order_number)"
    Write-Host "  Status: $($trackingResp.status)"
    Write-Host "  Truck: $($trackingResp.truck.plate_number) (Driver: $($trackingResp.truck.driver_name))"
    Write-Host "  Current Location: Lat=$($trackingResp.location.latitude), Lon=$($trackingResp.location.longitude)"
} else {
    Write-Host "✗ Failed to retrieve tracking data" -ForegroundColor Red
}

Write-Host "`n========== ALL TESTS COMPLETED ==========" -ForegroundColor Cyan
Write-Host "Test Summary:" -ForegroundColor Green
Write-Host "  User: $username"
Write-Host "  Truck ID: $truckID"
Write-Host "  Order Number: $orderNum"
Write-Host "  Final Status: Delivered"
Write-Host "  Location Points: 4"
