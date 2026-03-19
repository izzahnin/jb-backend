try {
    # Test the auth/login endpoint
    $body = @{
        username = "admin"
        password = "admin123"
    } | ConvertTo-Json
    
    Write-Host "Sending login request..."
    Write-Host "Body: $body"
    
    $response = Invoke-WebRequest -Uri 'http://localhost:8080/auth/login' -Method POST -ContentType 'application/json' -Body $body -UseBasicParsing
    Write-Host "Status: $($response.StatusCode)"
    Write-Host "Content: $($response.Content)"
}
catch {
    Write-Host "Status Code: $($_.Exception.Response.StatusCode)" 
    Write-Host "Status Description: $($_.Exception.Response.StatusDescription)"
    $streamReader = [System.IO.StreamReader]::new($_.Exception.Response.GetResponseStream())
    $content = $streamReader.ReadToEnd()
    Write-Host "Response Content: $content"
}
