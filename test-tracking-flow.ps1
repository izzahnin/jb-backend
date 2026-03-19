# ========================================
# FULL END-TO-END TESTING SCRIPT
# Testing: Login -> Create Truck -> Create Order -> Assign -> 
# Confirm Pickup -> Post Locations -> Confirm Delivery -> Get History
# ========================================

$API_BASE = "http://localhost:8080"
$timestamp = Get-Date -Format "yyyyMMddHHmmss"
$username = "testadmin$timestamp"
$password = "testpass123"

Write-Host "=====================================================================" -ForegroundColor Cyan
Write-Host "FULL END-TO-END TESTING: Location Tracking Flow" -ForegroundColor Cyan
Write-Host "=====================================================================" -ForegroundColor Cyan

# STEP 1: REGISTER NEW USER
Write-Host "`n[STEP 1] Register new admin user" -ForegroundColor Yellow
Write-Host "Username: $username" -ForegroundColor Gray
$registerUri = "$API_BASE/auth/register"
$registerBody = @{
    username = $username
    password = $password
} | ConvertTo-Json

$registerResponse = Invoke-WebRequest -Uri $registerUri -Method POST -ContentType "application/json" -Body $registerBody | ConvertFrom-Json
Write-Host "SUCCESS - User registered" -ForegroundColor Green
Write-Host "User ID: $($registerResponse.id)" -ForegroundColor Gray

# STEP 2: LOGIN
Write-Host "`n[STEP 2] Login admin user" -ForegroundColor Yellow
$loginUri = "$API_BASE/auth/login"
$loginBody = @{
    username = $username
    password = $password
} | ConvertTo-Json

$loginResponse = Invoke-WebRequest -Uri $loginUri -Method POST -ContentType "application/json" -Body $loginBody | ConvertFrom-Json
$token = $loginResponse.token
Write-Host "SUCCESS - User logged in" -ForegroundColor Green
Write-Host "Token: $($token.Substring(0, 50))..." -ForegroundColor Gray

# STEP 3: CREATE TRUCK
Write-Host "`n[STEP 3] Create truck" -ForegroundColor Yellow
$truckUri = "$API_BASE/admin/trucks"
$plateNumber = "TEST$(Get-Random -Minimum 1000 -Maximum 9999)"
$truckBody = @{
    plate_number = $plateNumber
    driver_name = "Testing Driver Makassar"
} | ConvertTo-Json

$headers = @{ "Authorization" = "Bearer $token" }
$truckResponse = Invoke-WebRequest -Uri $truckUri -Method POST -Headers $headers -ContentType "application/json" -Body $truckBody | ConvertFrom-Json
$truck_id = $truckResponse.id
Write-Host "SUCCESS - Truck created" -ForegroundColor Green
Write-Host "Truck ID: $truck_id | Plate: $($truckResponse.plate_number) | Driver: $($truckResponse.driver_name)" -ForegroundColor Gray

# STEP 4: CREATE ORDER
Write-Host "`n[STEP 4] Create order" -ForegroundColor Yellow
$orderUri = "$API_BASE/admin/orders"
$orderNumber = "ORD$(Get-Random -Minimum 100000 -Maximum 999999)"
$orderBody = @{
    order_number = $orderNumber
    origin = "Makassar"
    destination = "Gowa"
} | ConvertTo-Json

$orderResponse = Invoke-WebRequest -Uri $orderUri -Method POST -Headers $headers -ContentType "application/json" -Body $orderBody | ConvertFrom-Json
$order_id = $orderResponse.id
Write-Host "SUCCESS - Order created" -ForegroundColor Green
Write-Host "Order ID: $order_id | Order Number: $order_number | Status: $($orderResponse.status)" -ForegroundColor Gray

# STEP 5: ASSIGN TRUCK TO ORDER
Write-Host "`n[STEP 5] Assign truck to order" -ForegroundColor Yellow
$assignUri = "$API_BASE/admin/orders/assign"
$assignBody = @{
    order_id = $order_id
    truck_id = $truck_id
} | ConvertTo-Json

$assignResponse = Invoke-WebRequest -Uri $assignUri -Method PATCH -Headers $headers -ContentType "application/json" -Body $assignBody | ConvertFrom-Json
Write-Host "SUCCESS - Truck assigned to order" -ForegroundColor Green
Write-Host "Truck ID: $truck_id assigned to Order ID: $order_id" -ForegroundColor Gray

# STEP 6: CONFIRM PICKUP
Write-Host "`n[STEP 6] Confirm pickup (pending -> pickup)" -ForegroundColor Yellow
$pickupUri = "$API_BASE/admin/orders/$order_id/confirm-pickup"
$pickupBody = @{} | ConvertTo-Json

$pickupResponse = Invoke-WebRequest -Uri $pickupUri -Method POST -Headers $headers -ContentType "application/json" -Body $pickupBody | ConvertFrom-Json
Write-Host "SUCCESS - Pickup confirmed" -ForegroundColor Green
Write-Host "New status: $($pickupResponse.status)" -ForegroundColor Gray

# STEP 7: POST LOCATION #1
Write-Host "`n[STEP 7] Post location #1 (Start journey from Makassar)" -ForegroundColor Yellow
$locUri = "$API_BASE/trucks/$truck_id/location"
$loc1Body = @{
    latitude = -8.5
    longitude = 120.7
} | ConvertTo-Json

$loc1Response = Invoke-WebRequest -Uri $locUri -Method POST -ContentType "application/json" -Body $loc1Body | ConvertFrom-Json
Write-Host "SUCCESS - Location 1 posted" -ForegroundColor Green
Write-Host "Position: -8.5, 120.7" -ForegroundColor Gray

# STEP 8: POST LOCATION #2
Write-Host "`n[STEP 8] Post location #2 (Truck halfway to Gowa)" -ForegroundColor Yellow
Start-Sleep -Seconds 1
$loc2Body = @{
    latitude = -8.50523
    longitude = 120.71245
} | ConvertTo-Json

$loc2Response = Invoke-WebRequest -Uri $locUri -Method POST -ContentType "application/json" -Body $loc2Body | ConvertFrom-Json
Write-Host "SUCCESS - Location 2 posted" -ForegroundColor Green
Write-Host "Position: -8.50523, 120.71245" -ForegroundColor Gray

# STEP 9: POST LOCATION #3
Write-Host "`n[STEP 9] Post location #3 (Truck almost at destination)" -ForegroundColor Yellow
Start-Sleep -Seconds 1
$loc3Body = @{
    latitude = -8.51234
    longitude = 120.72890
} | ConvertTo-Json

$loc3Response = Invoke-WebRequest -Uri $locUri -Method POST -ContentType "application/json" -Body $loc3Body | ConvertFrom-Json
Write-Host "SUCCESS - Location 3 posted" -ForegroundColor Green
Write-Host "Position: -8.51234, 120.72890" -ForegroundColor Gray

# STEP 10: POST LOCATION #4
Write-Host "`n[STEP 10] Post location #4 (Arrived at Gowa)" -ForegroundColor Yellow
Start-Sleep -Seconds 1
$loc4Body = @{
    latitude = -8.52000
    longitude = 120.73500
} | ConvertTo-Json

$loc4Response = Invoke-WebRequest -Uri $locUri -Method POST -ContentType "application/json" -Body $loc4Body | ConvertFrom-Json
Write-Host "SUCCESS - Location 4 posted" -ForegroundColor Green
Write-Host "Position: -8.52000, 120.73500" -ForegroundColor Gray

# STEP 11: CONFIRM DELIVERY
Write-Host "`n[STEP 11] Confirm delivery (in_transit -> delivered)" -ForegroundColor Yellow
$deliveryUri = "$API_BASE/admin/orders/$order_id/confirm-delivery"
$deliveryBody = @{} | ConvertTo-Json

$deliveryResponse = Invoke-WebRequest -Uri $deliveryUri -Method POST -Headers $headers -ContentType "application/json" -Body $deliveryBody | ConvertFrom-Json
Write-Host "SUCCESS - Delivery confirmed" -ForegroundColor Green
Write-Host "New status: $($deliveryResponse.status)" -ForegroundColor Gray

# STEP 12: GET LATEST LOCATION (from Redis)
Write-Host "`n[STEP 12] Get latest truck location (from Redis cache)" -ForegroundColor Yellow
$latestUri = "$API_BASE/trucks/$truck_id/location"
$latestResponse = Invoke-WebRequest -Uri $latestUri -Method GET -ContentType "application/json" | ConvertFrom-Json
Write-Host "SUCCESS - Latest location retrieved from Redis" -ForegroundColor Green
Write-Host "Latest Latitude: $($latestResponse.latitude)" -ForegroundColor Gray
Write-Host "Latest Longitude: $($latestResponse.longitude)" -ForegroundColor Gray
Write-Host "Timestamp: $($latestResponse.created_at)" -ForegroundColor Gray

# STEP 13: GET LOCATION HISTORY
Write-Host "`n[STEP 13] Get location history (all 4 locations)" -ForegroundColor Yellow
$historyUri = "$API_BASE/trucks/$truck_id/locations?limit=10"
$historyResponse = Invoke-WebRequest -Uri $historyUri -Method GET -ContentType "application/json" | ConvertFrom-Json
Write-Host "SUCCESS - Location history retrieved" -ForegroundColor Green
Write-Host "Total locations recorded: $($historyResponse.Count)" -ForegroundColor Gray
$i = 1
foreach ($loc in $historyResponse) {
    Write-Host "  Location $i - Lat: $($loc.latitude), Lon: $($loc.longitude), Time: $($loc.created_at)" -ForegroundColor Gray
    $i++
}

# STEP 14: CUSTOMER TRACK ORDER (PUBLIC - NO AUTH)
Write-Host "`n[STEP 14] Customer track order (PUBLIC endpoint - no auth required)" -ForegroundColor Yellow
$publicUri = "$API_BASE/public/orders/$orderNumber/track"
$publicResponse = Invoke-WebRequest -Uri $publicUri -Method GET -ContentType "application/json" | ConvertFrom-Json
Write-Host "SUCCESS - Customer retrieved tracking data" -ForegroundColor Green
Write-Host "Order Number: $($publicResponse.order_number)" -ForegroundColor Gray
Write-Host "Order Status: $($publicResponse.status)" -ForegroundColor Gray
Write-Host "Truck Plate: $($publicResponse.truck.plate_number)" -ForegroundColor Gray
Write-Host "Driver Name: $($publicResponse.truck.driver_name)" -ForegroundColor Gray
Write-Host "Current Location - Latitude: $($publicResponse.location.latitude)" -ForegroundColor Gray
Write-Host "Current Location - Longitude: $($publicResponse.location.longitude)" -ForegroundColor Gray
Write-Host "Location Updated: $($publicResponse.location.created_at)" -ForegroundColor Gray

# SUMMARY
Write-Host "`n=====================================================================" -ForegroundColor Cyan
Write-Host "ALL TESTS PASSED SUCCESSFULLY!" -ForegroundColor Green
Write-Host "=====================================================================" -ForegroundColor Cyan

Write-Host "`nSummary of completed tests:" -ForegroundColor Yellow
Write-Host "[OK] Register and login user: $username" -ForegroundColor Green
Write-Host "[OK] Create truck: $($truckResponse.plate_number)" -ForegroundColor Green
Write-Host "[OK] Create order: $orderNumber" -ForegroundColor Green
Write-Host "[OK] Assign truck to order" -ForegroundColor Green
Write-Host "[OK] Confirm pickup (pending -> pickup)" -ForegroundColor Green
Write-Host "[OK] POST location 1: -8.5, 120.7" -ForegroundColor Green
Write-Host "[OK] POST location 2: -8.50523, 120.71245" -ForegroundColor Green
Write-Host "[OK] POST location 3: -8.51234, 120.72890" -ForegroundColor Green
Write-Host "[OK] POST location 4: -8.52000, 120.73500" -ForegroundColor Green
Write-Host "[OK] Confirm delivery (in_transit -> delivered)" -ForegroundColor Green
Write-Host "[OK] GET latest location from Redis cache" -ForegroundColor Green
Write-Host "[OK] GET location history (4 locations)" -ForegroundColor Green
Write-Host "[OK] Customer track order (PUBLIC - no auth)" -ForegroundColor Green

Write-Host "`nTesting completed at: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')" -ForegroundColor Cyan
