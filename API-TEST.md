# API Testing Guide - Meeting Slot Recommendation System

This guide demonstrates how to use the Meeting Slot Service API to solve the problem of finding optimal meeting times for geographically distributed teams.

## Problem Statement

**Challenge**: In a geographically distributed team, it is very hard to find common time to meet that works for everyone.

**Solution**: This API helps organizers:
1. Create events with multiple proposed time slots
2. Collect availability from all participants
3. Get smart recommendations for meeting times that work for everyone
4. Get fallback options when perfect alignment isn't possible

---

## Prerequisites

- API server running at `http://localhost:8080` (or your deployed ALB URL)
- API testing tool (Postman, Thunder Client, or similar)
- Basic understanding of REST APIs and JSON

## How to Use This Guide

Each API call shows:
1. **HTTP Method + Endpoint** (e.g., `POST /api/v1/users`)
2. **Request JSON** - Copy this payload and send it to the endpoint

Simply copy the JSON request body and use it in your API testing tool.

---

## Step-by-Step API Test Scenario

### Scenario Description

**Event**: "Brainstorming Meeting" for Q1 Planning  
**Organizer**: Sarah Johnson (Product Manager, Pacific Time)  
**Participants**: 
- Sarah Johnson (San Francisco, PST/UTC-8)
- Raj Kumar (Bangalore, IST/UTC+5:30)
- Emma Schmidt (Berlin, CET/UTC+1)
- Carlos Rivera (New York, EST/UTC-5)

**Proposed Slots by Organizer**:
- Option 1: January 15, 2026, 2:00 PM - 5:00 PM PST (10:00 PM PST → 6:00 AM IST → 11:00 PM CET → 5:00 PM EST)
- Option 2: January 16, 2026, 8:00 AM - 11:00 AM PST (4:00 AM PST → 9:30 PM IST → 5:00 PM CET → 11:00 AM EST)
- Option 3: January 17, 2026, 6:00 PM - 9:00 PM PST (Next day 12:00 AM PST → 7:30 AM IST → 3:00 AM CET → 9:00 PM EST)

**Meeting Duration Required**: 90 minutes (1.5 hours)

---

## API Testing Steps

### Step 1: Check API Health

Verify the API is running.

**Endpoint:** `GET /health`

**No request body required**

**Expected Response:**
```
ok
```

---

### Step 2: Create the Organizer User

Create Sarah Johnson (organizer).

**Endpoint:** `POST /api/v1/users`

**Request:**
```json
{
  "name": "Sarah Johnson",
  "email": "sarah.johnson@company.com"
}
```

**Expected Response:**
```json
{
  "id": "usr_1a2b3c4d",
  "name": "Sarah Johnson",
  "email": "sarah.johnson@company.com",
  "created_at": "2026-01-14T10:00:00Z"
}
```

**💡 Save the user ID**: `usr_1a2b3c4d` (use this in next steps)

---

### Step 3: Create Additional Participants

**3.1 Create Raj Kumar:**

**Endpoint:** `POST /api/v1/users`

**Request:**
```json
{
  "name": "Raj Kumar",
  "email": "raj.kumar@company.com"
}
```

**Expected Response:**
```json
{
  "id": "usr_2e3f4g5h",
  "name": "Raj Kumar",
  "email": "raj.kumar@company.com",
  "created_at": "2026-01-14T10:01:00Z"
}
```

**3.2 Create Emma Schmidt:**

**Endpoint:** `POST /api/v1/users`

**Request:**
```json
{
  "name": "Emma Schmidt",
  "email": "emma.schmidt@company.com"
}
```

**Expected Response:**
```json
{
  "id": "usr_3i4j5k6l",
  "name": "Emma Schmidt",
  "email": "emma.schmidt@company.com",
  "created_at": "2026-01-14T10:02:00Z"
}
```

**3.3 Create Carlos Rivera:**

**Endpoint:** `POST /api/v1/users`

**Request:**
```json
{
  "name": "Carlos Rivera",
  "email": "carlos.rivera@company.com"
}
```

**Expected Response:**
```json
{
  "id": "usr_4m5n6o7p",
  "name": "Carlos Rivera",
  "email": "carlos.rivera@company.com",
  "created_at": "2026-01-14T10:03:00Z"
}
```

**💡 Save all user IDs:**
- Sarah (Organizer): `usr_1a2b3c4d`
- Raj: `usr_2e3f4g5h`
- Emma: `usr_3i4j5k6l`
- Carlos: `usr_4m5n6o7p`

---

### Step 4: Create the Event with Proposed Slots

Sarah creates the "Brainstorming Meeting" event with 3 proposed time slots.

**Endpoint:** `POST /api/v1/events`

**Request:**
```json
{
  "title": "Q1 Brainstorming Meeting",
  "description": "Strategic planning session for Q1 initiatives. We need to align on priorities and resource allocation.",
  "organizer_id": "usr_1a2b3c4d",
  "duration_minutes": 90,
  "proposed_slots": [
    {
      "start_time": "2026-01-15T22:00:00Z",
      "end_time": "2026-01-16T01:00:00Z",
      "timezone": "America/Los_Angeles"
    },
    {
      "start_time": "2026-01-16T16:00:00Z",
      "end_time": "2026-01-16T19:00:00Z",
      "timezone": "America/Los_Angeles"
    },
    {
      "start_time": "2026-01-18T02:00:00Z",
      "end_time": "2026-01-18T05:00:00Z",
      "timezone": "America/Los_Angeles"
    }
  ]
}
```

**Expected Response:**
```json
{
  "id": "evt_abc123xyz",
  "title": "Q1 Brainstorming Meeting",
  "description": "Strategic planning session for Q1 initiatives. We need to align on priorities and resource allocation.",
  "organizer_id": "usr_1a2b3c4d",
  "duration_minutes": 90,
  "status": "pending",
  "proposed_slots": [
    {
      "id": "slot_001",
      "start_time": "2026-01-15T22:00:00Z",
      "end_time": "2026-01-16T01:00:00Z",
      "timezone": "America/Los_Angeles"
    },
    {
      "id": "slot_002",
      "start_time": "2026-01-16T16:00:00Z",
      "end_time": "2026-01-16T19:00:00Z",
      "timezone": "America/Los_Angeles"
    },
    {
      "id": "slot_003",
      "start_time": "2026-01-18T02:00:00Z",
      "end_time": "2026-01-18T05:00:00Z",
      "timezone": "America/Los_Angeles"
    }
  ],
  "created_at": "2026-01-14T10:05:00Z"
}
```

**💡 Save the event ID**: `evt_abc123xyz`

---

### Step 5: Add Participants to the Event

Add all participants to the event in a single API call.

**Endpoint:** `POST /api/v1/events/evt_abc123xyz/participants`

**Request:**
```json
{
  "participants": [
    {
      "user_id": "usr_1a2b3c4d"
    },
    {
      "user_id": "usr_2e3f4g5h"
    },
    {
      "user_id": "usr_3i4j5k6l"
    },
    {
      "user_id": "usr_4m5n6o7p"
    }
  ]
}
```

**Expected Response:**
```json
{
  "event_id": "evt_abc123xyz",
  "participants_added": 4,
  "participants": [
    {
      "user_id": "usr_1a2b3c4d",
      "status": "pending",
      "added_at": "2026-01-14T10:06:00Z"
    },
    {
      "user_id": "usr_2e3f4g5h",
      "status": "pending",
      "added_at": "2026-01-14T10:06:00Z"
    },
    {
      "user_id": "usr_3i4j5k6l",
      "status": "pending",
      "added_at": "2026-01-14T10:06:00Z"
    },
    {
      "user_id": "usr_4m5n6o7p",
      "status": "pending",
      "added_at": "2026-01-14T10:06:00Z"
    }
  ]
}
```

---

### Step 6: Participants Submit Their Availability

Now each participant submits their available time windows within the proposed slots.

**6.1 Sarah Johnson's Availability (Organizer - PST):**

Sarah is available during all proposed times except early morning on Jan 16.

**Endpoint:** `POST /api/v1/events/evt_abc123xyz/participants/usr_1a2b3c4d/availability`

**Request:**
```json
{
  "slots": [
    {
      "start_time": "2026-01-15T22:00:00Z",
      "end_time": "2026-01-16T01:00:00Z",
      "timezone": "America/Los_Angeles"
    },
    {
      "start_time": "2026-01-18T02:00:00Z",
      "end_time": "2026-01-18T04:30:00Z",
      "timezone": "America/Los_Angeles"
    }
  ]
}
```

**6.2 Raj Kumar's Availability (Bangalore - IST):**

Raj can attend only during his working hours (9 AM - 6 PM IST). This maps to:
- Jan 16, 2:30 AM - 11:30 AM UTC (corresponds to Jan 16, 8:00 AM - 5:00 PM IST)

**Endpoint:** `POST /api/v1/events/evt_abc123xyz/participants/usr_2e3f4g5h/availability`

**Request:**
```json
{
  "slots": [
    {
      "start_time": "2026-01-16T02:30:00Z",
      "end_time": "2026-01-16T11:30:00Z",
      "timezone": "Asia/Kolkata"
    }
  ]
}
```

**6.3 Emma Schmidt's Availability (Berlin - CET):**

Emma is available during business hours CET (9 AM - 6 PM).

**Endpoint:** `POST /api/v1/events/evt_abc123xyz/participants/usr_3i4j5k6l/availability`

**Request:**
```json
{
  "slots": [
    {
      "start_time": "2026-01-16T08:00:00Z",
      "end_time": "2026-01-16T17:00:00Z",
      "timezone": "Europe/Berlin"
    },
    {
      "start_time": "2026-01-17T08:00:00Z",
      "end_time": "2026-01-17T12:00:00Z",
      "timezone": "Europe/Berlin"
    }
  ]
}
```

**6.4 Carlos Rivera's Availability (New York - EST):**

Carlos is flexible and available across multiple time windows.

**Endpoint:** `POST /api/v1/events/evt_abc123xyz/participants/usr_4m5n6o7p/availability`

**Request:**
```json
{
  "slots": [
    {
      "start_time": "2026-01-15T20:00:00Z",
      "end_time": "2026-01-16T02:00:00Z",
      "timezone": "America/New_York"
    },
    {
      "start_time": "2026-01-16T14:00:00Z",
      "end_time": "2026-01-16T20:00:00Z",
      "timezone": "America/New_York"
    }
  ]
}
```

**Expected Response (for each availability submission):**
```json
{
  "event_id": "evt_abc123xyz",
  "user_id": "usr_...",
  "slots_count": 2,
  "submitted_at": "2026-01-14T10:10:00Z",
  "message": "Availability submitted successfully"
}
```

---

### Step 7: Get Recommendations

Now request the system to analyze all availabilities and recommend optimal meeting times.

**Endpoint:** `GET /api/v1/events/evt_abc123xyz/recommendations`

**No request body required**

**Expected Response:**

The system will return recommendations sorted by the number of available participants.

```json
{
  "event_id": "evt_abc123xyz",
  "event_title": "Q1 Brainstorming Meeting",
  "duration_minutes": 90,
  "total_participants": 4,
  "recommendations": [
    {
      "rank": 1,
      "start_time": "2026-01-16T16:00:00Z",
      "end_time": "2026-01-16T17:30:00Z",
      "timezone": "UTC",
      "available_count": 4,
      "availability_rate": 100.0,
      "status": "perfect_match",
      "available_participants": [
        {
          "user_id": "usr_1a2b3c4d",
          "name": "Sarah Johnson",
          "email": "sarah.johnson@company.com"
        },
        {
          "user_id": "usr_2e3f4g5h",
          "name": "Raj Kumar",
          "email": "raj.kumar@company.com"
        },
        {
          "user_id": "usr_3i4j5k6l",
          "name": "Emma Schmidt",
          "email": "emma.schmidt@company.com"
        },
        {
          "user_id": "usr_4m5n6o7p",
          "name": "Carlos Rivera",
          "email": "carlos.rivera@company.com"
        }
      ],
      "unavailable_participants": [],
      "local_times": {
        "America/Los_Angeles": "Jan 16, 2026 8:00 AM PST",
        "Asia/Kolkata": "Jan 16, 2026 9:30 PM IST",
        "Europe/Berlin": "Jan 16, 2026 5:00 PM CET",
        "America/New_York": "Jan 16, 2026 11:00 AM EST"
      }
    },
    {
      "rank": 2,
      "start_time": "2026-01-15T22:00:00Z",
      "end_time": "2026-01-15T23:30:00Z",
      "timezone": "UTC",
      "available_count": 3,
      "availability_rate": 75.0,
      "status": "partial_match",
      "available_participants": [
        {
          "user_id": "usr_1a2b3c4d",
          "name": "Sarah Johnson",
          "email": "sarah.johnson@company.com"
        },
        {
          "user_id": "usr_3i4j5k6l",
          "name": "Emma Schmidt",
          "email": "emma.schmidt@company.com"
        },
        {
          "user_id": "usr_4m5n6o7p",
          "name": "Carlos Rivera",
          "email": "carlos.rivera@company.com"
        }
      ],
      "unavailable_participants": [
        {
          "user_id": "usr_2e3f4g5h",
          "name": "Raj Kumar",
          "email": "raj.kumar@company.com",
          "reason": "Outside working hours (IST)"
        }
      ],
      "local_times": {
        "America/Los_Angeles": "Jan 15, 2026 2:00 PM PST",
        "Asia/Kolkata": "Jan 16, 2026 3:30 AM IST",
        "Europe/Berlin": "Jan 15, 2026 11:00 PM CET",
        "America/New_York": "Jan 15, 2026 5:00 PM EST"
      }
    },
    {
      "rank": 3,
      "start_time": "2026-01-18T02:00:00Z",
      "end_time": "2026-01-18T03:30:00Z",
      "timezone": "UTC",
      "available_count": 2,
      "availability_rate": 50.0,
      "status": "partial_match",
      "available_participants": [
        {
          "user_id": "usr_1a2b3c4d",
          "name": "Sarah Johnson",
          "email": "sarah.johnson@company.com"
        },
        {
          "user_id": "usr_4m5n6o7p",
          "name": "Carlos Rivera",
          "email": "carlos.rivera@company.com"
        }
      ],
      "unavailable_participants": [
        {
          "user_id": "usr_2e3f4g5h",
          "name": "Raj Kumar",
          "email": "raj.kumar@company.com",
          "reason": "No availability submitted"
        },
        {
          "user_id": "usr_3i4j5k6l",
          "name": "Emma Schmidt",
          "email": "emma.schmidt@company.com",
          "reason": "No availability submitted"
        }
      ],
      "local_times": {
        "America/Los_Angeles": "Jan 17, 2026 6:00 PM PST",
        "Asia/Kolkata": "Jan 18, 2026 7:30 AM IST",
        "Europe/Berlin": "Jan 18, 2026 3:00 AM CET",
        "America/New_York": "Jan 17, 2026 9:00 PM EST"
      }
    }
  ],
  "summary": {
    "total_recommendations": 3,
    "perfect_matches": 1,
    "partial_matches": 2,
    "best_recommendation": {
      "time": "2026-01-16T16:00:00Z",
      "availability_rate": 100.0,
      "participants_available": 4
    }
  },
  "generated_at": "2026-01-14T10:15:00Z"
}
```

---

## Understanding the Results

### ✅ Best Recommendation (Rank 1)
**Time**: January 16, 2026, 4:00 PM UTC (90 minutes)
- **Sarah**: 8:00 AM PST ✅ (Working hours)
- **Raj**: 9:30 PM IST ✅ (Late evening, acceptable)
- **Emma**: 5:00 PM CET ✅ (End of work day)
- **Carlos**: 11:00 AM EST ✅ (Mid-morning)

**Result**: **100% availability** - Works for everyone! 🎉

## Additional API Operations

### Get Event Details

**Endpoint:** `GET /api/v1/events/evt_abc123xyz`

**No request body required**

### List All Participants

**Endpoint:** `GET /api/v1/events/evt_abc123xyz/participants`

**No request body required**

### Get Participant's Availability

**Endpoint:** `GET /api/v1/events/evt_abc123xyz/participants/usr_1a2b3c4d/availability`

**No request body required**

### Update Availability

**Endpoint:** `PUT /api/v1/events/evt_abc123xyz/participants/usr_1a2b3c4d/availability`

**Request:**
```json
{
  "slots": [
    {
      "start_time": "2026-01-16T14:00:00Z",
      "end_time": "2026-01-16T20:00:00Z",
      "timezone": "America/Los_Angeles"
    }
  ]
}
```

### Remove Participant

**Endpoint:** `DELETE /api/v1/events/evt_abc123xyz/participants/usr_4m5n6o7p`

**No request body required**

### Delete Event

**Endpoint:** `DELETE /api/v1/events/evt_abc123xyz`

**No request body required**

---