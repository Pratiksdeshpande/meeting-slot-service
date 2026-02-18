# Test script to populate database
$ErrorActionPreference = "Continue"

# Step 1: Create 5 users
Write-Host "Creating users..."
$userABody = @{ name = "User A"; email = "usera@example.com" } | ConvertTo-Json
$userA = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/users" -Method Post -Body $userABody -ContentType "application/json"

$userBBody = @{ name = "User B"; email = "userb@example.com" } | ConvertTo-Json
$userB = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/users" -Method Post -Body $userBBody -ContentType "application/json"

$userCBody = @{ name = "User C"; email = "userc@example.com" } | ConvertTo-Json
$userC = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/users" -Method Post -Body $userCBody -ContentType "application/json"

$userDBody = @{ name = "User D"; email = "userd@example.com" } | ConvertTo-Json
$userD = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/users" -Method Post -Body $userDBody -ContentType "application/json"

$userEBody = @{ name = "User E"; email = "usere@example.com" } | ConvertTo-Json
$userE = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/users" -Method Post -Body $userEBody -ContentType "application/json"

Write-Host "Users created:"
Write-Host "User A: $($userA.data.id)"
Write-Host "User B: $($userB.data.id)"
Write-Host "User C: $($userC.data.id)"
Write-Host "User D: $($userD.data.id)"
Write-Host "User E: $($userE.data.id)"

# Step 2: Create event with proposed slots (1st Feb 2026, 2-4PM IST and 5-7PM IST)
Write-Host "`nCreating event..."
$eventBody = @{
  title = "Brainstorming meeting"
  organizer_id = $userA.data.id
  duration_minutes = 60
  proposed_slots = @(
    @{
      start_time = "2026-02-01T14:00:00+05:30"
      end_time = "2026-02-01T16:00:00+05:30"
      timezone = "Asia/Kolkata"
    },
    @{
      start_time = "2026-02-01T17:00:00+05:30"
      end_time = "2026-02-01T19:00:00+05:30"
      timezone = "Asia/Kolkata"
    }
  )
} | ConvertTo-Json -Depth 10

$event = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/events" -Method Post -Body $eventBody -ContentType "application/json"
Write-Host "Event created: $($event.data.id)"

# Step 3: Add participants B, C, D, E using batch API
Write-Host "`nAdding participants..."
$participantsBody = @{
  user_ids = @($userB.data.id, $userC.data.id, $userD.data.id, $userE.data.id)
} | ConvertTo-Json

$addResult = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/events/$($event.data.id)/participants" -Method Post -Body $participantsBody -ContentType "application/json"
Write-Host "Participants added: $($addResult.data.added_count)"

# Step 4: Add availability for User A (2-4PM & 5-7PM IST)
Write-Host "`nAdding availability for User A..."
$availABody = @{
  available_slots = @(
    @{
      start_time = "2026-02-01T14:00:00+05:30"
      end_time = "2026-02-01T16:00:00+05:30"
      timezone = "Asia/Kolkata"
    },
    @{
      start_time = "2026-02-01T17:00:00+05:30"
      end_time = "2026-02-01T19:00:00+05:30"
      timezone = "Asia/Kolkata"
    }
  )
} | ConvertTo-Json -Depth 10
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/events/$($event.data.id)/participants/$($userA.data.id)/availability" -Method Post -Body $availABody -ContentType "application/json" | Out-Null

# Step 5: Add availability for User B (1-4PM IST)
Write-Host "Adding availability for User B..."
$availBBody = @{
  available_slots = @(
    @{
      start_time = "2026-02-01T13:00:00+05:30"
      end_time = "2026-02-01T16:00:00+05:30"
      timezone = "Asia/Kolkata"
    }
  )
} | ConvertTo-Json -Depth 10
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/events/$($event.data.id)/participants/$($userB.data.id)/availability" -Method Post -Body $availBBody -ContentType "application/json" | Out-Null

# Step 6: Add availability for User C (3-7PM IST)
Write-Host "Adding availability for User C..."
$availCBody = @{
  available_slots = @(
    @{
      start_time = "2026-02-01T15:00:00+05:30"
      end_time = "2026-02-01T19:00:00+05:30"
      timezone = "Asia/Kolkata"
    }
  )
} | ConvertTo-Json -Depth 10
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/events/$($event.data.id)/participants/$($userC.data.id)/availability" -Method Post -Body $availCBody -ContentType "application/json" | Out-Null

# Step 7: Add availability for User D (3-4PM & 5-7PM IST)
Write-Host "Adding availability for User D..."
$availDBody = @{
  available_slots = @(
    @{
      start_time = "2026-02-01T15:00:00+05:30"
      end_time = "2026-02-01T16:00:00+05:30"
      timezone = "Asia/Kolkata"
    },
    @{
      start_time = "2026-02-01T17:00:00+05:30"
      end_time = "2026-02-01T19:00:00+05:30"
      timezone = "Asia/Kolkata"
    }
  )
} | ConvertTo-Json -Depth 10
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/events/$($event.data.id)/participants/$($userD.data.id)/availability" -Method Post -Body $availDBody -ContentType "application/json" | Out-Null

# Step 8: Add availability for User E (2-4PM IST)
Write-Host "Adding availability for User E..."
$availEBody = @{
  available_slots = @(
    @{
      start_time = "2026-02-01T14:00:00+05:30"
      end_time = "2026-02-01T16:00:00+05:30"
      timezone = "Asia/Kolkata"
    }
  )
} | ConvertTo-Json -Depth 10
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/events/$($event.data.id)/participants/$($userE.data.id)/availability" -Method Post -Body $availEBody -ContentType "application/json" | Out-Null

# Get recommendations
Write-Host "`nGetting recommendations..."
$recommendations = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/events/$($event.data.id)/recommendations" -Method Get

Write-Host "`n=== RESULTS ==="
Write-Host "Event ID: $($event.data.id)"
Write-Host "Recommendations:"
$recommendations.data | ConvertTo-Json -Depth 10
