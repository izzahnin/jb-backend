#!/usr/bin/env pwsh
# Comprehensive Endpoint Testing Script for PT. Jalur Berlian Backend
# Author: AI Assistant
# Date: March 10, 2026
# Usage: .\test-endpoints.ps1

Write-Host "🚀 PT. Jalur Berlian - Comprehensive Endpoint Testing" -ForegroundColor Cyan
Write-Host "======================================================" -ForegroundColor Cyan
Write-Host ""

$BaseURL = "http://localhost:8080"
$TestResults = @()
$TokenExpires = $null
$TruckID = $null
$OrderID = $null

function Test-Endpoint {
    param(
        [string]$Name,
        [string]$Method = "GET",
        [string]$Endpoint = "",
        [string]$Headers = "",
        [object]$Body = $null,
        [int]$ExpectedStatus = 200
    )
    
    $url = "$BaseURL$Endpoint"
    
    try {
        $params = @{
            Uri     = $url
            Method  = $Method
            Headers = @{
                "Content-Type" = "application/json"
            }
        }
        
        # Add custom headers if provided
        if ($Headers) {
            foreach ($header in $Headers -split ';') {
                $key, $value = $header -split '='
                $params.Headers[$key.Trim()] = $value.Trim()
            }
        }
        
        # Add body if provided
        if ($Body) {
            $params.Body = $Body | ConvertTo-Json
        }
        
        $response = Invoke-RestMethod @params -ErrorAction Stop
        $status = 200
        $success = $true
    }
    catch {
        $status = [int]$_.Exception.Response.StatusCode
        $response = $_.Exception.Response
        $success = $status -eq $ExpectedStatus
    }
    
    $statusEmoji = if ($success) { "✅" } else { "❌" }
    Write-Host "$statusEmoji [$Method] $Endpoint - Status: $status" -ForegroundColor (if ($success) { "Green" } else { "Red" })
    
    $TestResults += @{
        Name           = $Name
        Endpoint       = $Endpoint
        Method         = $Method
        Status         = $status
        ExpectedStatus = $ExpectedStatus
        Success        = $success
        Response       = $response
    }
    
    return @{
        Response = $response
        Status   = $status
        Success  = $success
    }
}

# ==========================================
# 1. AUTHENTICATION TEST
# ==========================================
Write-Host ""
Write-Host "📝 STEP 1: Authentication" -ForegroundColor Yellow
Write-Host "=========================" -ForegroundColor Yellow

$loginBody = @{
    username = "admin"
    password = "admin123"
}

$loginResult = Test-Endpoint -Name "Login" -Method "POST" -Endpoint "/auth/login" -Body $loginBody -ExpectedStatus 200

if ($loginResult.Success) {
    $token = $loginResult.Response.token
    $TokenExpires = $loginResult.Response.expires_at
    Write-Host "🎫 JWT Token acquired (expires: $(Get-Date -UnixTimeSeconds $TokenExpires -Format 'yyyy-MM-dd HH:mm:ss'))" -ForegroundColor Green
}
else {
    Write-Host "❌ FATAL: Could not obtain JWT token. Aborting tests." -ForegroundColor Red
    exit 1
}

# ==========================================
# 2. TRUCK MANAGEMENT TESTS (5 endpoints)
# ==========================================
Write-Host ""
Write-Host "🚛 STEP 2: Truck Management (5 endpoints)" -ForegroundColor Yellow
Write-Host "==========================================" -ForegroundColor Yellow

# 2.1 CREATE TRUCK
Write-Host ""
Write-Host "Creating trucks..." -ForegroundColor Cyan

$truck1Body = @{
    plate_number = "B 1001 AA"
    driver_name  = "Budi Santoso"
    is_active    = $true
}

$createTruckResult = Test-Endpoint -Name "Create Truck 1" -Method "POST" -Endpoint "/admin/trucks" `
    -Headers "Authorization=Bearer $token" -Body $truck1Body -ExpectedStatus 201

if ($createTruckResult.Success) {
    $TruckID = $createTruckResult.Response.data.id
    Write-Host "Truck created with ID: $TruckID" -ForegroundColor Green
}

$truck2Body = @{
    plate_number = "B 1002 BB"
    driver_name  = "Andi Wijaya"
    is_active    = $true
}
Test-Endpoint -Name "Create Truck 2" -Method "POST" -Endpoint "/admin/trucks" `
    -Headers "Authorization=Bearer $token" -Body $truck2Body -ExpectedStatus 201

# 2.2 LIST TRUCKS WITH PAGINATION
Write-Host ""
Write-Host "Testing truck pagination..." -ForegroundColor Cyan
Test-Endpoint -Name "List Trucks (page=1)" -Method "GET" -Endpoint "/admin/trucks" `
    -Headers "Authorization=Bearer $token" -ExpectedStatus 200

Test-Endpoint -Name "List Trucks (page=1, limit=5)" -Method "GET" -Endpoint "/admin/trucks?page=1&limit=5" `
    -Headers "Authorization=Bearer $token" -ExpectedStatus 200

# 2.3 GET SINGLE TRUCK
Write-Host ""
Write-Host "Testing truck detail..." -ForegroundColor Cyan
Test-Endpoint -Name "Get Truck Detail" -Method "GET" -Endpoint "/admin/trucks/$TruckID" `
    -Headers "Authorization=Bearer $token" -ExpectedStatus 200

# 2.4 UPDATE TRUCK
Write-Host ""
Write-Host "Testing truck update..." -ForegroundColor Cyan
$updateTruckBody = @{
    driver_name = "Budi Santoso Updated"
    is_active   = $true
}
Test-Endpoint -Name "Update Truck" -Method "PUT" -Endpoint "/admin/trucks/$TruckID" `
    -Headers "Authorization=Bearer $token" -Body $updateTruckBody -ExpectedStatus 200

# 2.5 DEACTIVATE TRUCK (soft delete)
# We'll deactivate truck 2 to test soft delete
Write-Host ""
Write-Host "Testing truck deactivation (soft delete)..." -ForegroundColor Cyan
Test-Endpoint -Name "Deactivate Truck 2" -Method "DELETE" -Endpoint "/admin/trucks/2" `
    -Headers "Authorization=Bearer $token" -ExpectedStatus 200

# ==========================================
# 3. ORDER MANAGEMENT TESTS (6 endpoints)
# ==========================================
Write-Host ""
Write-Host "📦 STEP 3: Order Management (6 endpoints)" -ForegroundColor Yellow
Write-Host "=========================================" -ForegroundColor Yellow

# 3.1 CREATE ORDER
Write-Host ""
Write-Host "Creating orders..." -ForegroundColor Cyan

$order1Body = @{
    order_number = "ORD-20260310-0001"
    origin       = "Jakarta"
    destination  = "Bandung"
    truck_id     = $null
    status       = "pending"
}

$createOrderResult = Test-Endpoint -Name "Create Order 1" -Method "POST" -Endpoint "/admin/orders" `
    -Headers "Authorization=Bearer $token" -Body $order1Body -ExpectedStatus 201

if ($createOrderResult.Success) {
    $OrderID = $createOrderResult.Response.data.id
    Write-Host "Order created with ID: $OrderID" -ForegroundColor Green
}

$order2Body = @{
    order_number = "ORD-20260310-0002"
    origin       = "Surabaya"
    destination  = "Malang"
    truck_id     = $null
    status       = "pending"
}
Test-Endpoint -Name "Create Order 2" -Method "POST" -Endpoint "/admin/orders" `
    -Headers "Authorization=Bearer $token" -Body $order2Body -ExpectedStatus 201

# 3.2 LIST ORDERS
Write-Host ""
Write-Host "Testing order pagination..." -ForegroundColor Cyan
Test-Endpoint -Name "List Orders (page=1)" -Method "GET" -Endpoint "/admin/orders" `
    -Headers "Authorization=Bearer $token" -ExpectedStatus 200

Test-Endpoint -Name "List Orders (limit=5)" -Method "GET" -Endpoint "/admin/orders?limit=5" `
    -Headers "Authorization=Bearer $token" -ExpectedStatus 200

# 3.3 GET SINGLE ORDER
Write-Host ""
Write-Host "Testing order detail..." -ForegroundColor Cyan
Test-Endpoint -Name "Get Order Detail" -Method "GET" -Endpoint "/admin/orders/$OrderID" `
    -Headers "Authorization=Bearer $token" -ExpectedStatus 200

# 3.4 ASSIGN TRUCK TO ORDER
Write-Host ""
Write-Host "Testing truck assignment..." -ForegroundColor Cyan
$assignBody = @{
    order_id = $OrderID
    truck_id = $TruckID
}
Test-Endpoint -Name "Assign Truck to Order" -Method "PATCH" -Endpoint "/admin/orders/assign" `
    -Headers "Authorization=Bearer $token" -Body $assignBody -ExpectedStatus 200

# 3.5 UPDATE ORDER STATUS
Write-Host ""
Write-Host "Testing order status updates..." -ForegroundColor Cyan

# pending → pickup
$statusBody = @{
    status = "pickup"
}
Test-Endpoint -Name "Update Order Status (pending→pickup)" -Method "PATCH" -Endpoint "/admin/orders/$OrderID" `
    -Headers "Authorization=Bearer $token" -Body $statusBody -ExpectedStatus 200

# pickup → in_transit
$statusBody = @{
    status = "in_transit"
}
Test-Endpoint -Name "Update Order Status (pickup→in_transit)" -Method "PATCH" -Endpoint "/admin/orders/$OrderID" `
    -Headers "Authorization=Bearer $token" -Body $statusBody -ExpectedStatus 200

# in_transit → delivered
$statusBody = @{
    status = "delivered"
}
Test-Endpoint -Name "Update Order Status (in_transit→delivered)" -Method "PATCH" -Endpoint "/admin/orders/$OrderID" `
    -Headers "Authorization=Bearer $token" -Body $statusBody -ExpectedStatus 200

# 3.6 CANCEL ORDER (test on second order)
Write-Host ""
Write-Host "Testing order cancellation..." -ForegroundColor Cyan
Test-Endpoint -Name "Cancel Order 2" -Method "DELETE" -Endpoint "/admin/orders/2" `
    -Headers "Authorization=Bearer $token" -ExpectedStatus 200

# ==========================================
# 4. LOCATION TRACKING TESTS (3 endpoints) - PUBLIC
# ==========================================
Write-Host ""
Write-Host "📍 STEP 4: Location Tracking (3 endpoints, PUBLIC)" -ForegroundColor Yellow
Write-Host "=================================================" -ForegroundColor Yellow

# 4.1 POST LOCATION UPDATE
Write-Host ""
Write-Host "Sending location updates..." -ForegroundColor Cyan

$location1Body = @{
    lat = -6.200000
    lon = 106.816666
    ts  = "2026-03-10T10:15:30Z"
}
Test-Endpoint -Name "Send Location 1" -Method "POST" -Endpoint "/trucks/$TruckID/location" `
    -Body $location1Body -ExpectedStatus 200

$location2Body = @{
    lat = -6.195000
    lon = 106.820000
    ts  = "2026-03-10T10:16:30Z"
}
Test-Endpoint -Name "Send Location 2" -Method "POST" -Endpoint "/trucks/$TruckID/location" `
    -Body $location2Body -ExpectedStatus 200

# 4.2 GET LOCATION HISTORY
Write-Host ""
Write-Host "Testing location history..." -ForegroundColor Cyan
Test-Endpoint -Name "Get Location History (default limit)" -Method "GET" -Endpoint "/trucks/$TruckID/locations" `
    -ExpectedStatus 200

Test-Endpoint -Name "Get Location History (limit=10)" -Method "GET" -Endpoint "/trucks/$TruckID/locations?limit=10" `
    -ExpectedStatus 200

# 4.3 GET CURRENT LOCATION
Write-Host ""
Write-Host "Testing current location..." -ForegroundColor Cyan
Test-Endpoint -Name "Get Current Location" -Method "GET" -Endpoint "/trucks/$TruckID/location" `
    -ExpectedStatus 200

# ==========================================
# 5. PUBLIC TRACKING TESTS (1 endpoint)
# ==========================================
Write-Host ""
Write-Host "🌐 STEP 5: Public Tracking (1 endpoint)" -ForegroundColor Yellow
Write-Host "=====================================" -ForegroundColor Yellow

Write-Host ""
Write-Host "Testing customer order tracking..." -ForegroundColor Cyan
Test-Endpoint -Name "Track Order (Public)" -Method "GET" -Endpoint "/public/orders/ORD-20260310-0001/track" `
    -ExpectedStatus 200

# ==========================================
# 6. ERROR CASES & EDGE CASES
# ==========================================
Write-Host ""
Write-Host "⚠️  STEP 6: Error Handling & Edge Cases" -ForegroundColor Yellow
Write-Host "=====================================" -ForegroundColor Yellow

# 6.1 MISSING JWT
Write-Host ""
Write-Host "Testing auth protection (missing JWT)..." -ForegroundColor Cyan
Test-Endpoint -Name "GET /admin/trucks (no JWT)" -Method "GET" -Endpoint "/admin/trucks" `
    -ExpectedStatus 401

# 6.2 INVALID ID
Write-Host ""
Write-Host "Testing invalid ID handling..." -ForegroundColor Cyan
Test-Endpoint -Name "GET /admin/trucks/invalid-id" -Method "GET" -Endpoint "/admin/trucks/invalid-id" `
    -Headers "Authorization=Bearer $token" -ExpectedStatus 400

# 6.3 NOT FOUND
Write-Host ""
Write-Host "Testing 404 cases..." -ForegroundColor Cyan
Test-Endpoint -Name "GET /admin/trucks/999 (not found)" -Method "GET" -Endpoint "/admin/trucks/999" `
    -Headers "Authorization=Bearer $token" -ExpectedStatus 404

Test-Endpoint -Name "GET /admin/orders/999 (not found)" -Method "GET" -Endpoint "/admin/orders/999" `
    -Headers "Authorization=Bearer $token" -ExpectedStatus 404

# 6.4 VALIDATION
Write-Host ""
Write-Host "Testing validation errors..." -ForegroundColor Cyan

$invalidOrderBody = @{
    order_number = ""  # empty - should fail
    origin       = "Jakarta"
    destination  = "Bandung"
}
Test-Endpoint -Name "POST /admin/orders (empty order_number)" -Method "POST" -Endpoint "/admin/orders" `
    -Headers "Authorization=Bearer $token" -Body $invalidOrderBody -ExpectedStatus 422

# ==========================================
# SUMMARY REPORT
# ==========================================
Write-Host ""
Write-Host "=" * 60 -ForegroundColor Cyan
Write-Host "📊 TEST SUMMARY" -ForegroundColor Cyan
Write-Host "=" * 60 -ForegroundColor Cyan
Write-Host ""

$passed = $TestResults | Where-Object { $_.Success } | Measure-Object | Select-Object -ExpandProperty Count
$failed = $TestResults | Where-Object { -not $_.Success } | Measure-Object | Select-Object -ExpandProperty Count
$total = $TestResults.Count

Write-Host "Total Tests: $total" -ForegroundColor White
Write-Host "✅ Passed:   $passed" -ForegroundColor Green
Write-Host "❌ Failed:   $failed" -ForegroundColor Red
Write-Host ""

if ($failed -gt 0) {
    Write-Host "Failed Tests:" -ForegroundColor Red
    $TestResults | Where-Object { -not $_.Success } | ForEach-Object {
        Write-Host "  ❌ [$($_.Method)] $($_.Endpoint) - Expected: $($_.ExpectedStatus), Got: $($_.Status)" -ForegroundColor Red
    }
    Write-Host ""
}

$passRate = [math]::Round(($passed / $total) * 100, 2)
Write-Host "Pass Rate: $passRate%" -ForegroundColor (if ($passRate -eq 100) { "Green" } else { "Yellow" })
Write-Host ""

if ($failed -eq 0) {
    Write-Host "🎉 ALL TESTS PASSED! Ready for deployment." -ForegroundColor Green
}
else {
    Write-Host "⚠️  Some tests failed. Review the errors above." -ForegroundColor Yellow
}

Write-Host ""
Write-Host "Test execution completed at: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')" -ForegroundColor Gray
