# Test all member functions
$ErrorActionPreference = "Stop"

# Get auth token
$loginBody = @{
    email = "admin@tausigmasigma.org"
    password = "BlueLedger2026!"
} | ConvertTo-Json

$loginResp = Invoke-WebRequest -Uri "http://localhost:8081/v1/auth/login" -Method POST -ContentType "application/json" -Body $loginBody | ConvertFrom-Json
$token = $loginResp.data.access_token

# Get first member
$membersResp = Invoke-WebRequest -Uri "http://localhost:8081/v1/members?per_page=1" -Headers @{ Authorization = "Bearer $token" } | ConvertFrom-Json
$member = $membersResp.data[0]
$memberId = $member.id

Write-Host "=== MEMBER FUNCTION TESTS ===" -ForegroundColor Cyan
Write-Host "Subject: $($member.first_name) $($member.last_name)" -ForegroundColor Yellow
Write-Host "Member ID: $memberId" -ForegroundColor Yellow
Write-Host ""

# TEST 1: GET
Write-Host "TEST 1: GET /members/:id" -ForegroundColor Green
$getResp = Invoke-WebRequest -Uri "http://localhost:8081/v1/members/$memberId" -Headers @{ Authorization = "Bearer $token" } | ConvertFrom-Json
if ($getResp.data -and $getResp.data.id -eq $memberId) {
    Write-Host "✅ PASS: Retrieved member successfully" -ForegroundColor Green
    Write-Host "   Name: $($getResp.data.first_name) $($getResp.data.last_name)"
    Write-Host "   Email: $($getResp.data.email)"
    Write-Host "   Status: $($getResp.data.status)"
} else {
    Write-Host "❌ FAIL: Could not retrieve member" -ForegroundColor Red
}
Write-Host ""

# TEST 2: UPDATE
Write-Host "TEST 2: PUT /members/:id (Update Status)" -ForegroundColor Green
$updateBody = @{
    status = "inactive"
    job_title = "Test Manager"
    city = "San Francisco"
} | ConvertTo-Json
$updateResp = Invoke-WebRequest -Uri "http://localhost:8081/v1/members/$memberId" -Method PUT -Headers @{ Authorization = "Bearer $token" } -ContentType "application/json" -Body $updateBody | ConvertFrom-Json
if ($updateResp.data -and $updateResp.data.status -eq "inactive") {
    Write-Host "✅ PASS: Member updated successfully" -ForegroundColor Green
    Write-Host "   New Status: $($updateResp.data.status)"
    Write-Host "   New Job Title: $($updateResp.data.job_title)"
    Write-Host "   New City: $($updateResp.data.city)"
} else {
    Write-Host "❌ FAIL: Could not update member" -ForegroundColor Red
    Write-Host "   Error: $($updateResp.error.message)"
}
Write-Host ""

# TEST 3: GET XP HISTORY
Write-Host "TEST 3: GET /members/:id/xp-history" -ForegroundColor Green
$xpResp = Invoke-WebRequest -Uri "http://localhost:8081/v1/members/$memberId/xp-history" -Headers @{ Authorization = "Bearer $token" } | ConvertFrom-Json
if ($xpResp.meta) {
    Write-Host "✅ PASS: XP history retrieved" -ForegroundColor Green
    Write-Host "   Total entries: $($xpResp.meta.total)"
} else {
    Write-Host "❌ FAIL: Could not retrieve XP history" -ForegroundColor Red
}
Write-Host ""

# TEST 4: DELETE (Soft delete)
Write-Host "TEST 4: DELETE /members/:id (Deactivate)" -ForegroundColor Green
$deleteResp = Invoke-WebRequest -Uri "http://localhost:8081/v1/members/$memberId" -Method DELETE -Headers @{ Authorization = "Bearer $token" } -UseBasicParsing | ConvertFrom-Json
if ($deleteResp.data.message -and $deleteResp.data.message -eq "member removed") {
    Write-Host "✅ PASS: Member deactivated successfully" -ForegroundColor Green
    Write-Host "   Message: $($deleteResp.data.message)"
    
    # Verify the member status changed - it should be soft deleted
    try {
        $verifyResp = Invoke-WebRequest -Uri "http://localhost:8081/v1/members/$memberId" -Headers @{ Authorization = "Bearer $token" } -UseBasicParsing -ErrorAction Stop
        $verifyData = $verifyResp.Content | ConvertFrom-Json
        if ($verifyData.data.status -eq "inactive") {
            Write-Host "✅ PASS: Member status is now inactive" -ForegroundColor Green
        }
    } catch {
        if ($_.Exception.Response.StatusCode -eq 404) {
            Write-Host "✅ PASS: Member no longer appears in API (soft deleted)" -ForegroundColor Green
        } else {
            Write-Host "⚠️  INFO: Delete verification returned: $($_.Exception.Response.StatusCode)" -ForegroundColor Yellow
        }
    }
} else {
    Write-Host "❌ FAIL: Could not delete member" -ForegroundColor Red
    if ($deleteResp.error) {
        Write-Host "   Error: $($deleteResp.error.message)"
    }
}
Write-Host ""

Write-Host "=== TESTS COMPLETED ===" -ForegroundColor Cyan
